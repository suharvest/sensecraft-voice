package recording

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/errors"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/config"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/service"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/util/httpclient"
	"k8s.io/klog/v2"
)

type RecordingGetter interface {
	Recording() Interface
}

type Interface interface {
	Save(ctx context.Context, in SaveRequest) (*model.Recording, error)
	List(ctx context.Context, in ListRequest) (*ListResponse, error)
	Query(ctx context.Context, in QueryRequest) (*QueryResponse, error)
	GetKeywordMatches(ctx context.Context, in KeywordMatchRequest) (*KeywordMatchResponse, error)
}

type ListRequest struct {
	// 分页参数 - 可选
	Offset int `form:"offset"`
	Limit  int `form:"limit"`

	// 时间范围 - 可选
	StartTime int64 `form:"start_time"`
	EndTime   int64 `form:"end_time"`

	// 设备时间范围 - 可选
	DeviceStartTime int64 `form:"device_start_time"`
	DeviceEndTime   int64 `form:"device_end_time"`

	// 业务参数 - 可选
	StoreID    int      `form:"store_id"`
	LocationID int      `form:"location_id"`
	MacAddress []string `form:"mac_address"` // 支持多个MAC地址

	// 类型参数 - 可选
	Type int `form:"type"` // 1: 查询音频记录并返回play_url, 0或不传: 不查询音频记录

	// 状态过滤 - 可选
	Status *int8 `form:"status"` // 支持状态过滤
}

type ListResponse struct {
	Total int64              `json:"total"`
	Items []*model.Recording `json:"items"`
}

type QueryRequest struct {
	DeviceId       string `json:"deviceId"`
	StartTimestamp int64  `json:"startTimestamp"`
	EndTimestamp   int64  `json:"endTimestamp"`
	Sid            string `json:"sid"`
}

type QueryResponse struct {
	Total int64              `json:"total"`
	Items []*model.Recording `json:"items"`
}

type SaveRequest struct {
	SessionID  string `json:"sessionID"`
	AudioID    string `json:"audioID"`
	MacAddress string `json:"mac_address"`
	Speaker    struct {
		Confidence  float64 `json:"confidence"`
		Identified  bool    `json:"identified"`
		SpeakerID   string  `json:"speaker_id"`
		SpeakerName string  `json:"speaker_name"`
	} `json:"speaker"`
	Text       string `json:"text"`
	TextLength int    `json:"textLength"`
	Timestamp  int64  `json:"timestamp"`
	Type       string `json:"type"`
	WordCount  int    `json:"wordCount"`
	Status     int8   `json:"status"`
}

// KeywordMatchRequest 关键词匹配查询请求
type KeywordMatchRequest struct {
	// 分页参数
	Offset int `form:"offset"` // 移除验证标签，在路由中处理
	Limit  int `form:"limit"`  // 移除验证标签，在路由中处理

	// 查询条件
	MacAddress  string `form:"mac_address"`  // MAC地址
	KeywordID   int64  `form:"keyword_id"`   // 关键词ID
	RecordingID int64  `form:"recording_id"` // 录音ID
	StoreID     int64  `form:"store_id"`     // 门店ID

	// 时间范围
	StartTime int64 `form:"start_time"` // 开始时间（毫秒时间戳）
	EndTime   int64 `form:"end_time"`   // 结束时间（毫秒时间戳）
}

// KeywordMatchResponse 关键词匹配查询响应
type KeywordMatchResponse struct {
	Total  int64                 `json:"total"`
	Items  []*model.KeywordMatch `json:"items"`
	Limit  int                   `json:"limit"`
	Offset int                   `json:"offset"`
}

type recording struct {
	cc      config.Config
	factory db.ShareDaoFactory
}

func (r *recording) Save(ctx context.Context, in SaveRequest) (*model.Recording, error) {
	if in.MacAddress == "" {
		return nil, errors.ErrInvalidRequest
	}
	mac := strings.ToLower(in.MacAddress)

	obj := &model.Recording{
		SessionID:   in.SessionID,
		AudioID:     in.AudioID,
		MacAddress:  mac,
		SpeakerId:   in.Speaker.SpeakerID,
		SpeakerName: in.Speaker.SpeakerName,
		Text:        in.Text,
		Status:      in.Status,
		CreatedAtMs: time.Now().UnixNano() / 1e6,
		DeviceTime:  in.Timestamp,
	}
	out, err := r.factory.Recording().Create(ctx, obj)
	if err != nil {
		return nil, errors.ErrServerInternal
	}

	// 关键词匹配处理
	go r.processKeywordMatching(context.Background(), out.Id, mac, in.Text)

	return out, nil
}

// processKeywordMatching 处理关键词匹配（异步执行）
func (r *recording) processKeywordMatching(ctx context.Context, recordingID int64, macAddress, text string) {
	// 创建关键词匹配服务
	// 使用带缓存的关键词匹配器（如果可用）
	// 这里暂以原有构造函数保留，缓存注入将在启动流程统一注入到服务层单例中
	matcher := service.NewKeywordMatcher(r.factory)

	// 执行关键词匹配
	results, err := matcher.MatchKeywords(ctx, text, macAddress)
	if err != nil {
		klog.Errorf("Failed to match keywords for recording %d: %v", recordingID, err)
		return
	}

	// 如果没有匹配到关键词，直接返回
	if len(results) == 0 {
		klog.V(4).Infof("No keywords matched for recording %d", recordingID)
		return
	}

	// 保存匹配结果到数据库
	if err := matcher.SaveMatches(ctx, recordingID, macAddress, results); err != nil {
		klog.Errorf("Failed to save keyword matches for recording %d: %v", recordingID, err)
		return
	}

	klog.Infof("Successfully matched %d keywords for recording %d", len(results), recordingID)
}

func (r *recording) List(ctx context.Context, in ListRequest) (*ListResponse, error) {
	// 设置默认分页参数
	if in.Offset <= 0 {
		in.Offset = 0
	}

	if in.Limit <= 0 || in.Limit > 5000 {
		in.Limit = 3000 // 默认每页50条
	}

	// 构建DAO查询请求
	daoReq := db.ListRequest{
		Offset:          in.Offset,
		Limit:           in.Limit,
		StartTime:       in.StartTime,
		EndTime:         in.EndTime,
		DeviceStartTime: in.DeviceStartTime,
		DeviceEndTime:   in.DeviceEndTime,
		StoreID:         in.StoreID,
		LocationID:      in.LocationID,
		MacAddress:      in.MacAddress,
		Status:          in.Status,
	}

	// 获取总数
	total, err := r.factory.Recording().Count(ctx, daoReq)
	if err != nil {
		return nil, errors.ErrServerInternal
	}

	// 获取列表数据
	items, err := r.factory.Recording().List(ctx, daoReq)
	if err != nil {
		return nil, errors.ErrServerInternal
	}

	// 根据type字段决定是否生成播放链接
	if in.Type == 1 {
		// type=1: 查询音频记录并返回play_url
		// 批量查询存在的音频文件
		audioExistsMap := r.batchCheckAudioFiles(ctx, items)

		// 为每个录音项生成播放链接（只对有音频文件的录音生成）
		for _, item := range items {
			key := fmt.Sprintf("%s:%s", item.SessionID, item.AudioID)
			if audioExistsMap[key] {
				item.PlayURL = r.generatePlayURL(item.SessionID, item.AudioID)
			}
			// 如果没有音频文件，PlayURL保持为空字符串
		}
	}
	// type=0或不传: 不查询音频记录，不返回play_url

	return &ListResponse{
		Total: total,
		Items: items,
	}, nil
}

func (r *recording) Query(ctx context.Context, in QueryRequest) (*QueryResponse, error) {
	// 1. 先保存查询日志
	queryLog := &model.QueryLog{
		DeviceID:       in.DeviceId,
		StartTimestamp: in.StartTimestamp,
		EndTimestamp:   in.EndTimestamp,
		SessionID:      in.Sid,
		SeeedAPIStatus: 0, // 初始状态：未处理
	}

	savedLog, err := r.factory.QueryLog().Create(ctx, queryLog)
	if err != nil {
		klog.Errorf("Failed to create query log: %v", err)
		// 记录日志失败不影响主流程，继续执行
	}

	// 2. 构建DAO查询请求
	daoReq := db.QueryRequest{
		DeviceId:       in.DeviceId,
		StartTimestamp: in.StartTimestamp,
		EndTimestamp:   in.EndTimestamp,
		Sid:            in.Sid,
	}

	// 3. 获取查询结果
	items, err := r.factory.Recording().Query(ctx, daoReq)
	if err != nil {
		return nil, errors.ErrServerInternal
	}

	// 4. 异步处理：调用OpenAI API，然后调用Seeed API，并更新查询日志
	if savedLog != nil {
		go r.processRecordingsAsync(items, in.Sid, savedLog.Id)
	} else {
		go r.processRecordingsAsync(items, in.Sid, 0)
	}

	return &QueryResponse{
		Total: int64(len(items)),
		Items: items,
	}, nil
}

// compressRecordingData 压缩录音数据为简单格式
// 将speaker_id和text组合成字符串，使用基础压缩策略
func (r *recording) compressRecordingData(recordings []*model.Recording) string {
	if len(recordings) == 0 {
		return ""
	}

	var result strings.Builder
	currentSpeaker := ""
	currentText := ""

	for _, rec := range recordings {
		// 简化speaker_id: speaker_01 -> S1, speaker_02 -> S2
		speaker := strings.Replace(rec.SpeakerId, "speaker_", "S", 1)
		if speaker == "" {
			speaker = "Unknown"
		}

		if speaker == currentSpeaker {
			// 同一speaker，合并文本
			if currentText != "" {
				currentText += " " + rec.Text
			} else {
				currentText = rec.Text
			}
		} else {
			// 不同speaker，先输出之前的
			if currentSpeaker != "" {
				result.WriteString(fmt.Sprintf("%s: %s; ", currentSpeaker, currentText))
			}
			currentSpeaker = speaker
			currentText = rec.Text
		}
	}

	// 输出最后一个
	if currentSpeaker != "" {
		result.WriteString(fmt.Sprintf("%s: %s", currentSpeaker, currentText))
	}

	compressed := result.String()

	// 限制长度，避免过长
	if len(compressed) > 2000 {
		compressed = compressed[:2000] + "..."
	}

	return compressed
}

// buildEnhancedContent 构建包含收集表链接的增强内容
func (r *recording) buildEnhancedContent(openAIAnswer string) string {
	// 收集表链接
	surveyLink := "https://doc.weixin.qq.com/forms/AGEAZwfLABEAUsA4QaAAGkCNR743d6xqf?page=1"

	// 吸引用户点击填写的文案 - 使用 Markdown 链接格式
	surveyText := `
💡 您的意见，是我们前进的动力!!!!
只需1分钟，帮助我们打造更好的智能会议体验!!!!!
[🔗 立即反馈](` + surveyLink + `)
感谢您的宝贵时间!!!!✨`

	// 组合原始内容和收集表文案
	enhancedContent := openAIAnswer + surveyText

	klog.Infof("构建增强内容完成，原始长度: %d, 增强后长度: %d", len(openAIAnswer), len(enhancedContent))

	return enhancedContent
}

// processRecordingsAsync 异步处理录音数据
// 1. 压缩数据
// 2. 调用OpenAI API (重试3次)
// 3. 调用Seeed API (重试3次)
// 4. 更新查询日志
func (r *recording) processRecordingsAsync(recordings []*model.Recording, sessionId string, queryLogID int64) {
	klog.Infof("开始异步处理录音数据，会话ID: %s, 录音数量: %d, 查询日志ID: %d", sessionId, len(recordings), queryLogID)

	var openAIAnswer string
	var seeedAPIStatus int8 = 0 // 0-未处理，1-成功，2-失败
	var processingError string

	// 1. 压缩数据
	compressedData := r.compressRecordingData(recordings)
	if compressedData == "" {
		klog.Warningf("录音数据为空，发送友好提示，会话ID: %s", sessionId)
		// 设置友好的提示语
		openAIAnswer = "当前暂无会议数据，如有疑问，请联系方案团队"

		// 构建包含收集表链接的增强内容
		enhancedContent := r.buildEnhancedContent(openAIAnswer)

		// 直接推送增强后的提示语到 Seeed API
		klog.Infof("开始推送空数据提示到 Seeed API，会话ID: %s", sessionId)
		seeedSuccess := r.callSeeedAPIWithRetry(sessionId, enhancedContent, queryLogID)
		if seeedSuccess {
			klog.Infof("空数据提示推送成功，会话ID: %s", sessionId)
			seeedAPIStatus = 1
		} else {
			klog.Errorf("空数据提示推送失败，会话ID: %s", sessionId)
			processingError = "推送空数据提示失败"
			seeedAPIStatus = 2
		}

		// 更新查询日志
		if queryLogID > 0 {
			r.updateQueryLog(queryLogID, openAIAnswer, seeedAPIStatus, processingError)
		}
		return
	}

	klog.Infof("录音数据压缩完成，会话ID: %s, 压缩后长度: %d", sessionId, len(compressedData))
	klog.V(4).Infof("压缩后的录音数据: %s", compressedData)

	// 2. 调用OpenAI API
	klog.Infof("开始调用OpenAI API，会话ID: %s", sessionId)
	openAIAnswer, openAISuccess := r.callOpenAIAPIWithRetry(compressedData, sessionId)
	if !openAISuccess {
		klog.Errorf("OpenAI API调用失败，会话ID: %s", sessionId)
		processingError = "OpenAI API调用失败"
		seeedAPIStatus = 2
		// 更新查询日志
		if queryLogID > 0 {
			r.updateQueryLog(queryLogID, openAIAnswer, seeedAPIStatus, processingError)
		}
		return
	}

	klog.Infof("OpenAI API调用成功，会话ID: %s, 回复长度: %d", sessionId, len(openAIAnswer))
	klog.V(4).Infof("OpenAI回复内容: %s", openAIAnswer)

	// 3. 构建包含收集表链接的完整内容
	enhancedContent := r.buildEnhancedContent(openAIAnswer)

	// 4. 调用Seeed API，使用增强后的content
	klog.Infof("开始调用Seeed API，会话ID: %s", sessionId)
	seeedSuccess := r.callSeeedAPIWithRetry(sessionId, enhancedContent, queryLogID)
	if seeedSuccess {
		klog.Infof("Seeed API调用成功，会话ID: %s", sessionId)
		seeedAPIStatus = 1
	} else {
		klog.Errorf("Seeed API调用失败，会话ID: %s", sessionId)
		processingError = "Seeed API调用失败"
		seeedAPIStatus = 2
	}

	// 4. 更新查询日志
	if queryLogID > 0 {
		r.updateQueryLog(queryLogID, openAIAnswer, seeedAPIStatus, processingError)
	}
}

// updateQueryLog 更新查询日志
func (r *recording) updateQueryLog(queryLogID int64, openAIAnswer string, seeedAPIStatus int8, processingError string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := r.factory.QueryLog().UpdateProcessingResult(ctx, queryLogID, openAIAnswer, seeedAPIStatus, processingError)
	if err != nil {
		klog.Errorf("Failed to update query log %d: %v", queryLogID, err)
	} else {
		klog.Infof("Successfully updated query log %d", queryLogID)
	}
}

// generateSeeedToken 生成Seeed API Token
// 规则: MD5(固定秘钥+当天日期)
// 固定秘钥: 由IT部门提供
// 当天日期: yyyyMMdd格式
func (r *recording) generateSeeedToken() string {
	// 从配置获取固定秘钥
	secretKey := r.cc.Seeed.SecretKey

	// 获取当天日期，格式为 yyyyMMdd
	today := time.Now().Format("20060102")

	// 拼接固定秘钥和当天日期
	rawToken := secretKey + today

	// 计算MD5
	hash := md5.Sum([]byte(rawToken))
	token := fmt.Sprintf("%x", hash)

	return token
}

// callOpenAIAPIWithRetry 调用OpenAI API，带重试逻辑
// 失败时每隔1秒重试，最多重试3次
// 返回OpenAI API的回复内容
func (r *recording) callOpenAIAPIWithRetry(compressedData, sessionId string) (string, bool) {
	klog.Infof("准备调用OpenAI API，会话ID: %s, 模型: %s, 超时: %ds", sessionId, r.cc.OpenAI.Model, r.cc.OpenAI.Timeout)

	// 创建HTTP客户端
	client := httpclient.NewClient(&httpclient.Config{
		Timeout:    time.Duration(r.cc.OpenAI.Timeout) * time.Second,
		RetryCount: 0, // 我们手动控制重试
		Headers: map[string]string{
			"Authorization": "Bearer " + r.cc.OpenAI.APIKey,
			"Content-Type":  "application/json",
		},
	})

	// 构建OpenAI请求
	request := map[string]interface{}{
		"model": r.cc.OpenAI.Model,
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": "你是一个智能助手，请分析以下录音内容并给出简洁的总结。",
			},
			{
				"role":    "user",
				"content": compressedData,
			},
		},
		"max_tokens":  r.cc.OpenAI.MaxTokens,
		"temperature": r.cc.OpenAI.Temperature,
		"user":        sessionId,
	}

	klog.V(4).Infof("OpenAI请求参数: model=%s, max_tokens=%d, temperature=%.2f",
		r.cc.OpenAI.Model, r.cc.OpenAI.MaxTokens, r.cc.OpenAI.Temperature)

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		klog.Infof("OpenAI API调用尝试 %d/%d，会话ID: %s", i+1, maxRetries, sessionId)

		// 创建新的context
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.cc.OpenAI.Timeout)*time.Second)

		// 调用OpenAI API
		var response map[string]interface{}
		apiURL := r.cc.OpenAI.BaseURL
		// 如果 base_url 已经包含 /v1 路径，则直接使用；否则添加 /v1/chat/completions
		if strings.HasSuffix(apiURL, "/v1") {
			apiURL += "/chat/completions"
		} else {
			apiURL += "/v1/chat/completions"
		}
		klog.V(4).Infof("OpenAI API URL: %s", apiURL)
		err := client.PostJSON(ctx, apiURL, request, &response)
		cancel()

		if err == nil {
			klog.Infof("OpenAI API调用成功，会话ID: %s", sessionId)

			// 提取回复内容
			if choices, ok := response["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if message, ok := choice["message"].(map[string]interface{}); ok {
						if content, ok := message["content"].(string); ok {
							klog.Infof("成功提取OpenAI回复内容，会话ID: %s, 长度: %d", sessionId, len(content))
							return content, true
						}
					}
				}
			}
			klog.Warningf("OpenAI API响应格式异常，会话ID: %s", sessionId)
		} else {
			klog.Warningf("OpenAI API调用失败，会话ID: %s, 错误: %v", sessionId, err)
		}

		// 如果不是最后一次尝试，等待1秒后重试
		if i < maxRetries-1 {
			klog.Infof("等待1秒后重试，会话ID: %s", sessionId)
			time.Sleep(1 * time.Second)
		}
	}

	klog.Errorf("OpenAI API调用最终失败，会话ID: %s, 已重试 %d 次", sessionId, maxRetries)
	return "", false
}

// callSeeedAPIWithRetry 调用Seeed API，带重试逻辑
// 失败时每隔1秒重试，最多重试3次
// queryLogID: 关联的查询日志ID，用于记录API调用详情
func (r *recording) callSeeedAPIWithRetry(sessionId, content string, queryLogID int64) bool {
	klog.Infof("准备调用Seeed API，会话ID: %s, 内容长度: %d, 查询日志ID: %d", sessionId, len(content), queryLogID)

	// 创建HTTP客户端
	client := httpclient.NewClient(&httpclient.Config{
		Timeout:    time.Duration(r.cc.Seeed.Timeout) * time.Second,
		RetryCount: 0, // 我们手动控制重试
	})

	// 生成Token
	token := r.generateSeeedToken()

	// 创建请求
	request := &httpclient.SeeedModifyRecordingRequest{
		Token:   token,
		SId:     sessionId,
		Content: content,
	}

	// 构建完整URL
	requestURL := r.cc.Seeed.BaseURL + "/api/SF/ModifyRecordingMeetingContent"

	// 打印发送给Seeed API的详细内容
	klog.Infof("Seeed API请求参数: Token=%s, SId=%s, Content长度=%d", token, sessionId, len(content))
	klog.Infof("发送给Seeed API的内容: %s", content)

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		retryCount := i
		klog.Infof("Seeed API调用尝试 %d/%d，会话ID: %s", i+1, maxRetries, sessionId)

		// 创建新的context
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.cc.Seeed.Timeout)*time.Second)

		// 调用Seeed API
		response, err := client.SeeedModifyRecordingMeetingContent(ctx, r.cc.Seeed.BaseURL, request)

		// 准备日志记录
		apiLog := &model.SeeedAPILog{
			QueryLogID:     queryLogID,
			SessionID:      sessionId,
			RequestToken:   token,
			RequestContent: content,
			RequestURL:     requestURL,
			RetryCount:     retryCount,
		}

		if err == nil && response != nil {
			apiLog.ResponseCode = response.Code
			apiLog.ResponseMessage = response.Msg
			// 将响应体序列化为JSON
			if responseJSON, jsonErr := json.Marshal(response); jsonErr == nil {
				apiLog.ResponseBody = string(responseJSON)
			}

			if response.Code == 0 {
				apiLog.IsSuccess = true
				klog.Infof("Seeed API调用成功，会话ID: %s, 响应码: %d", sessionId, response.Code)
				klog.Infof("Seeed API成功发送的内容: %s", content)

				// 记录成功的API调用
				r.saveSeeedAPILog(apiLog)
				cancel()
				return true
			} else {
				apiLog.IsSuccess = false
				apiLog.ErrorMessage = fmt.Sprintf("API返回错误码: %d, 消息: %s", response.Code, response.Msg)
				klog.Warningf("Seeed API返回错误码，会话ID: %s, 错误码: %d, 消息: %s", sessionId, response.Code, response.Msg)
			}
		} else {
			apiLog.IsSuccess = false
			if err != nil {
				apiLog.ErrorMessage = err.Error()
				klog.Warningf("Seeed API调用失败，会话ID: %s, 错误: %v", sessionId, err)
			}
		}

		// 记录失败的API调用
		r.saveSeeedAPILog(apiLog)
		cancel()

		// 如果不是最后一次尝试，等待1秒后重试
		if i < maxRetries-1 {
			klog.Infof("等待1秒后重试Seeed API，会话ID: %s", sessionId)
			time.Sleep(1 * time.Second)
		}
	}

	klog.Errorf("Seeed API调用最终失败，会话ID: %s, 已重试 %d 次", sessionId, maxRetries)
	return false
}

// saveSeeedAPILog 保存Seeed API调用日志
func (r *recording) saveSeeedAPILog(apiLog *model.SeeedAPILog) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := r.factory.SeeedAPILog().Create(ctx, apiLog)
	if err != nil {
		klog.Errorf("Failed to save Seeed API log: %v", err)
	} else {
		klog.V(4).Infof("Successfully saved Seeed API log for session %s, retry %d", apiLog.SessionID, apiLog.RetryCount)
	}
}

// batchCheckAudioFiles 批量检查音频文件是否存在
func (r *recording) batchCheckAudioFiles(ctx context.Context, items []*model.Recording) map[string]bool {
	// 构建查询条件
	sessionAudioPairs := make([]db.SessionAudioPair, 0, len(items))

	for _, item := range items {
		sessionAudioPairs = append(sessionAudioPairs, db.SessionAudioPair{
			SessionID: item.SessionID,
			AudioID:   item.AudioID,
		})
	}

	// 批量查询audio_recordings表
	audioRecordings, err := r.factory.AudioRecording().GetBySessionIDAndAudioIDBatch(ctx, sessionAudioPairs)
	if err != nil {
		// 如果批量查询失败，返回空map（不生成任何播放链接）
		return make(map[string]bool)
	}

	// 构建存在映射
	existsMap := make(map[string]bool)
	for _, ar := range audioRecordings {
		key := fmt.Sprintf("%s:%s", ar.SessionID, ar.AudioID)
		existsMap[key] = true
	}

	return existsMap
}

// generatePlayURL 生成播放链接
func (r *recording) generatePlayURL(sessionID, audioID string) string {
	baseURL := r.cc.Default.BaseURL
	if baseURL == "" {
		baseURL = "http://127.0.0.1:3008" // 默认值
	}
	return fmt.Sprintf("%s/api/v1/audio/session/%s/audio/%s/play", baseURL, sessionID, audioID)
}

// GetKeywordMatches 获取关键词匹配记录
func (r *recording) GetKeywordMatches(ctx context.Context, in KeywordMatchRequest) (*KeywordMatchResponse, error) {
	var matches []*model.KeywordMatch
	var total int64
	var err error

	// 先执行 count 查询获取满足条件的总数（不应用 limit）
	if in.RecordingID > 0 {
		// 根据录音ID查询
		total, err = r.factory.KeywordMatch().CountByRecordingID(ctx, in.RecordingID)
		if err != nil {
			klog.Errorf("Failed to count keyword matches by recording ID: %v", err)
			return nil, errors.ErrServerInternal
		}
		matches, err = r.factory.KeywordMatch().GetByRecordingID(ctx, in.RecordingID)
	} else if in.StoreID > 0 {
		// 根据门店ID查询
		total, err = r.factory.KeywordMatch().CountByStoreID(ctx, in.StoreID, in.StartTime, in.EndTime)
		if err != nil {
			klog.Errorf("Failed to count keyword matches by store ID: %v", err)
			return nil, errors.ErrServerInternal
		}
		matches, err = r.factory.KeywordMatch().GetByStoreID(ctx, in.StoreID, in.StartTime, in.EndTime, in.Limit)
	} else if in.MacAddress != "" || in.KeywordID > 0 || (in.StartTime > 0 && in.EndTime > 0) {
		// 使用组合查询方法，支持多个条件的组合
		total, err = r.factory.KeywordMatch().CountByConditions(ctx, in.MacAddress, in.KeywordID, in.StartTime, in.EndTime)
		if err != nil {
			klog.Errorf("Failed to count keyword matches by conditions: %v", err)
			return nil, errors.ErrServerInternal
		}
		matches, err = r.factory.KeywordMatch().GetByConditions(ctx, in.MacAddress, in.KeywordID, in.StartTime, in.EndTime, in.Limit)
	} else {
		// 如果没有指定查询条件，查询所有匹配记录
		total, err = r.factory.KeywordMatch().CountAll(ctx)
		if err != nil {
			klog.Errorf("Failed to count all keyword matches: %v", err)
			return nil, errors.ErrServerInternal
		}
		matches, err = r.factory.KeywordMatch().GetAll(ctx, in.Limit)
	}

	if err != nil {
		klog.Errorf("Failed to get keyword matches: %v", err)
		return nil, errors.ErrServerInternal
	}

	// 应用分页（只有在设置了 limit 时才进行分页）
	if in.Limit > 0 {
		start := in.Offset
		end := start + in.Limit

		if start >= len(matches) {
			matches = []*model.KeywordMatch{}
		} else if end > len(matches) {
			matches = matches[start:]
		} else {
			matches = matches[start:end]
		}
	}

	return &KeywordMatchResponse{
		Total:  total,
		Items:  matches,
		Limit:  in.Limit,
		Offset: in.Offset,
	}, nil
}

func NewRecording(cfg config.Config, f db.ShareDaoFactory) *recording {
	return &recording{cc: cfg, factory: f}
}
