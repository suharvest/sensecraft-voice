package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/types"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/util/httpclient"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/util/openai"
	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
)

// ChatGetter 聊天获取器接口
type ChatGetter interface {
	Chat() Interface
}

// Interface 聊天控制器接口
type Interface interface {
	// SendMessage 发送聊天消息
	SendMessage(ctx context.Context, req *types.ChatRequest) error
	// StreamMessage 流式发送聊天消息
	StreamMessage(ctx context.Context, req *types.ChatRequest, streamChan chan<- interface{}) error
	// GetChatHistory 获取聊天历史
	GetChatHistory(ctx context.Context, sessionID string, limit int) ([]*types.ChatMessage, error)
	// GetChatSession 获取聊天会话
	GetChatSession(ctx context.Context, sessionID string) (*types.ChatSession, error)
	// GetChatSessions 获取聊天会话列表
	GetChatSessions(ctx context.Context, userID string) ([]*types.ChatSession, error)
	// DeleteChatSession 删除聊天会话
	DeleteChatSession(ctx context.Context, sessionID string) error
	// DeleteChatSessions 批量删除聊天会话
	DeleteChatSessions(ctx context.Context, sessionIDs []string) error
	// UpdateChatSessionTitle 更新聊天会话标题
	UpdateChatSessionTitle(ctx context.Context, sessionID, title string) error
}

// Controller 聊天控制器实现
type Controller struct {
	factory        db.ShareDaoFactory
	config         *types.ChatConfig
	titleGenerator *openai.TitleGenerator
}

// NewController 创建聊天控制器
func NewController(factory db.ShareDaoFactory, config *types.ChatConfig, openaiConfig *types.OpenAIConfig) Interface {
	// 创建OpenAI客户端用于标题生成
	openaiClient := openai.NewClient(&openai.Config{
		APIKey:      openaiConfig.APIKey,
		BaseURL:     openaiConfig.BaseURL,
		Timeout:     openaiConfig.Timeout,
		MaxTokens:   openaiConfig.MaxTokens,
		Temperature: openaiConfig.Temperature,
		Model:       openaiConfig.Model,
	}, nil) // 标题生成不需要上下文管理器

	// 创建标题生成器
	titleGenerator := openai.NewTitleGenerator(openaiClient, &openai.Config{
		APIKey:      openaiConfig.APIKey,
		BaseURL:     openaiConfig.BaseURL,
		Timeout:     openaiConfig.Timeout,
		MaxTokens:   openaiConfig.MaxTokens,
		Temperature: openaiConfig.Temperature,
		Model:       openaiConfig.Model,
	})

	return &Controller{
		factory:        factory,
		config:         config,
		titleGenerator: titleGenerator,
	}
}

// SendMessage 发送聊天消息
func (c *Controller) SendMessage(ctx context.Context, req *types.ChatRequest) error {
	// 创建或获取会话（使用临时会话ID，后续会从API响应中获取真实的conversation_id）
	tempSessionID := req.ConversationID

	// 创建临时会话
	session := &types.ChatSession{
		SessionID: tempSessionID,
		UserID:    req.User,
		Status:    "active",
	}

	if err := c.factory.Chat().SaveChatSession(ctx, session); err != nil {
		klog.Errorf("保存聊天会话失败: %v", err)
		return fmt.Errorf("保存会话失败: %w", err)
	}

	// 创建流式客户端
	streamClient := httpclient.NewStreamClient(&httpclient.StreamConfig{
		BaseURL:     c.config.BaseURL,
		Timeout:     time.Duration(c.config.Timeout) * time.Second,
		EnableDebug: c.config.EnableDebug,
	}).SetAuthToken(c.config.APIKey)

	// 转换请求格式
	chatReq := &httpclient.ChatRequest{
		Inputs:         req.Inputs,
		Query:          req.Query,
		ResponseMode:   req.ResponseMode,
		ConversationID: tempSessionID, // 使用临时会话ID
		User:           req.User,
		Files:          convertFiles(req.Files),
	}

	// 定义事件处理器
	var chatStats *types.ChatStats
	var fullMessage string
	var conversationID string

	handler := func(event *httpclient.StreamEvent) error {
		// 从API响应中获取conversation_id
		if conversationID == "" && event.Data != nil {
			if convID, ok := event.Data["conversation_id"].(string); ok && convID != "" {
				conversationID = convID
			}
		}

		// 保存消息到数据库
		message := &types.ChatMessage{
			SessionID: tempSessionID, // 使用临时会话ID
			MessageID: generateMessageID(),
			Event:     event.Event,
			Content:   extractContent(event),
			Data:      serializeData(event.Data),
		}

		if err := c.factory.Chat().SaveChatMessage(ctx, message); err != nil {
			klog.Errorf("保存聊天消息失败: %v", err)
		}

		// 处理不同类型的事件
		switch event.Event {
		case "message":
			if answer, ok := event.Data["answer"].(string); ok {
				fullMessage += answer
			}
		case "message_end":
			// 提取统计信息
			if metadata, ok := event.Data["metadata"].(map[string]interface{}); ok {
				if usage, ok := metadata["usage"].(map[string]interface{}); ok {
					chatStats = &types.ChatStats{
						SessionID:      tempSessionID,  // 使用临时会话ID
						ConversationID: conversationID, // 使用从API获取的conversation_id
						CreatedAt:      time.Now().UnixMilli(),
					}

					if totalTokens, ok := usage["total_tokens"].(float64); ok {
						chatStats.TotalTokens = int64(totalTokens)
					}
					if totalPrice, ok := usage["total_price"].(string); ok {
						if price, err := parseFloat(totalPrice); err == nil {
							chatStats.TotalPrice = price
						}
					}
					if currency, ok := usage["currency"].(string); ok {
						chatStats.Currency = currency
					}
					if latency, ok := usage["latency"].(float64); ok {
						chatStats.Latency = latency
					}
				}
			}

			// 保存统计信息（如果会话ID不为空）
			if chatStats != nil && tempSessionID != "" {
				if err := c.factory.Chat().SaveChatStats(ctx, chatStats); err != nil {
					klog.Warningf("保存聊天统计失败: %v，继续处理", err)
				}
			}
		}

		return nil
	}

	// 发送流式请求
	if err := streamClient.PostChatStream(ctx, "/v1/chat-messages", chatReq, handler); err != nil {
		klog.Errorf("发送聊天消息失败: %v", err)
		return fmt.Errorf("发送消息失败: %w", err)
	}

	return nil
}

// GetChatHistory 获取聊天历史
func (c *Controller) GetChatHistory(ctx context.Context, sessionID string, limit int) ([]*types.ChatMessage, error) {
	return c.factory.Chat().GetChatHistory(ctx, sessionID, limit)
}

// GetChatSession 获取聊天会话
func (c *Controller) GetChatSession(ctx context.Context, sessionID string) (*types.ChatSession, error) {
	return c.factory.Chat().GetChatSession(ctx, sessionID)
}

// StreamMessage 流式发送聊天消息
func (c *Controller) StreamMessage(ctx context.Context, req *types.ChatRequest, streamChan chan<- interface{}) error {
	defer func() {
		if r := recover(); r != nil {
			klog.Errorf("StreamMessage panic: %v", r)
			// 发送错误事件
			select {
			case streamChan <- gin.H{"event": "error", "data": fmt.Sprintf("内部错误: %v", r)}:
			case <-ctx.Done():
				klog.Infof("上下文已取消，停止发送错误事件")
			default:
				klog.Warningf("流式通道已满，丢弃错误事件")
			}
		}
	}()

	// 创建或获取会话（使用临时会话ID，后续会从API响应中获取真实的conversation_id）
	tempSessionID := req.ConversationID
	// 只有当conversation_id完全不存在时才自动生成
	// 如果用户明确提供了空字符串，则保持为空
	if req.ConversationID == "" {
		// 检查是否明确提供了conversation_id字段
		// 这里我们简化处理：如果用户明确提供了空字符串，就不自动生成
		// 因为Gin会绑定所有存在的字段，包括空字符串
		klog.Infof("用户提供了空的conversation_id，保持为空")
	} else {
		klog.Infof("使用用户提供的会话ID: '%s'", tempSessionID)
	}

	klog.Infof("开始流式聊天，临时会话ID: '%s'", tempSessionID)

	// 1. 立即存储用户输入
	userMessageID := generateMessageID()
	userMessage := &types.ChatMessage{
		SessionID: tempSessionID,
		MessageID: userMessageID,
		Event:     "user",    // 用户角色
		Content:   req.Query, // 用户问题
		CreatedAt: time.Now().UnixMilli(),
	}

	klog.Infof("准备保存用户输入: SessionID=%s, MessageID=%s, Content=%s",
		tempSessionID, userMessageID, req.Query)

	if err := c.factory.Chat().SaveChatMessage(ctx, userMessage); err != nil {
		klog.Errorf("保存用户输入失败: %v，继续处理", err)
	} else {
		klog.Infof("成功保存用户输入，SessionID=%s, MessageID=%s, 内容长度: %d",
			tempSessionID, userMessageID, len(req.Query))
	}

	// 2. 等待最终结果时再存储会话和AI回复
	klog.Infof("等待最终结果，暂不创建临时会话")

	// 创建流式客户端
	streamClient := httpclient.NewStreamClient(&httpclient.StreamConfig{
		BaseURL:     c.config.BaseURL,
		Timeout:     time.Duration(c.config.Timeout) * time.Second,
		EnableDebug: c.config.EnableDebug,
	}).SetAuthToken(c.config.APIKey)

	// 转换请求格式
	chatReq := &httpclient.ChatRequest{
		Inputs:         req.Inputs,
		Query:          req.Query,
		ResponseMode:   req.ResponseMode,
		ConversationID: tempSessionID, // 使用临时会话ID
		User:           req.User,
		Files:          convertFiles(req.Files),
	}

	// 定义事件处理器
	var chatStats *types.ChatStats
	var fullMessage string
	var conversationID string

	handler := func(event *httpclient.StreamEvent) error {
		defer func() {
			if r := recover(); r != nil {
				klog.Errorf("事件处理器panic: %v", r)
				// 发送错误事件
				select {
				case streamChan <- gin.H{"event": "error", "data": fmt.Sprintf("事件处理错误: %v", r)}:
				case <-ctx.Done():
					klog.Infof("上下文已取消，停止发送事件处理错误")
				default:
					klog.Warningf("流式通道已满，丢弃事件处理错误")
				}
			}
		}()

		// 从API响应中获取conversation_id
		if conversationID == "" && event.Data != nil {
			if convID, ok := event.Data["conversation_id"].(string); ok && convID != "" {
				conversationID = convID
				klog.Infof("获取到真实conversation_id: %s", conversationID)
			}
		}

		// 发送事件到流式通道（不存储到数据库）
		select {
		case streamChan <- event.Data:
		case <-ctx.Done():
			klog.Infof("上下文已取消，停止发送事件")
			return ctx.Err()
		default:
			klog.Warningf("流式通道已满，丢弃事件: %s", event.Event)
		}

		// 处理不同类型的事件
		switch event.Event {
		case "message":
			// 累积消息内容，但不存储
			if answer, ok := event.Data["answer"].(string); ok {
				fullMessage += answer
			}
		case "message_end":
			// 只在最终结果时存储数据
			klog.Infof("收到message_end事件，开始存储最终结果")

			// 1. 存储最终会话（使用真实的conversation_id）
			if conversationID != "" {
				finalSession := &types.ChatSession{
					SessionID: conversationID, // 使用真实的conversation_id
					UserID:    req.User,
					Status:    "active",
					CreatedAt: time.Now().UnixMilli(),
					UpdatedAt: time.Now().UnixMilli(),
				}

				if err := c.factory.Chat().SaveChatSession(ctx, finalSession); err != nil {
					klog.Warningf("保存最终聊天会话失败: %v，继续处理", err)
				} else {
					klog.Infof("成功保存最终聊天会话: %s", conversationID)

					// 异步生成会话标题
					go c.generateSessionTitleAsync(ctx, conversationID, req.Query)
				}

				// 2. 存储 AI 回复消息
				aiMessageID := generateMessageID()
				aiMessage := &types.ChatMessage{
					SessionID: conversationID, // 使用真实的conversation_id
					MessageID: aiMessageID,
					Event:     "assistant", // AI 助手角色
					Content:   fullMessage, // 完整的 AI 回复内容
					Data:      serializeData(event.Data),
					CreatedAt: time.Now().UnixMilli(),
				}

				klog.Infof("准备保存 AI 回复: SessionID=%s, MessageID=%s, Content长度=%d",
					conversationID, aiMessageID, len(fullMessage))

				if err := c.factory.Chat().SaveChatMessage(ctx, aiMessage); err != nil {
					klog.Errorf("保存 AI 回复消息失败: %v，继续处理", err)
				} else {
					klog.Infof("成功保存 AI 回复消息，SessionID=%s, MessageID=%s, 内容长度: %d",
						conversationID, aiMessageID, len(fullMessage))
				}

				// 3. 存储元数据消息（可选）
				metadataMessageID := generateMessageID()
				metadataContent := serializeData(event.Data)
				metadataMessage := &types.ChatMessage{
					SessionID: conversationID,
					MessageID: metadataMessageID,
					Event:     "metadata", // 元数据
					Content:   metadataContent,
					CreatedAt: time.Now().UnixMilli(),
				}

				klog.Infof("准备保存元数据: SessionID=%s, MessageID=%s, Content长度=%d",
					conversationID, metadataMessageID, len(metadataContent))

				if err := c.factory.Chat().SaveChatMessage(ctx, metadataMessage); err != nil {
					klog.Errorf("保存元数据消息失败: %v，继续处理", err)
				} else {
					klog.Infof("成功保存元数据消息，SessionID=%s, MessageID=%s",
						conversationID, metadataMessageID)
				}
			}

			// 3. 提取并存储统计信息
			if metadata, ok := event.Data["metadata"].(map[string]interface{}); ok {
				if usage, ok := metadata["usage"].(map[string]interface{}); ok {
					chatStats = &types.ChatStats{
						SessionID:      conversationID, // 使用真实的conversation_id
						ConversationID: conversationID,
						CreatedAt:      time.Now().UnixMilli(),
					}

					if totalTokens, ok := usage["total_tokens"].(float64); ok {
						chatStats.TotalTokens = int64(totalTokens)
					}
					if totalPrice, ok := usage["total_price"].(string); ok {
						if price, err := parseFloat(totalPrice); err == nil {
							chatStats.TotalPrice = price
						}
					}
					if currency, ok := usage["currency"].(string); ok {
						chatStats.Currency = currency
					}
					if latency, ok := usage["latency"].(float64); ok {
						chatStats.Latency = latency
					}

					// 保存统计信息
					if err := c.factory.Chat().SaveChatStats(ctx, chatStats); err != nil {
						klog.Warningf("保存聊天统计失败: %v，继续处理", err)
					} else {
						klog.Infof("成功保存聊天统计，总tokens: %d", chatStats.TotalTokens)
					}
				}
			}

			// 发送完成事件
			select {
			case streamChan <- gin.H{"event": "completed", "conversation_id": conversationID}:
			case <-ctx.Done():
				klog.Infof("上下文已取消，停止发送完成事件")
				return ctx.Err()
			default:
				klog.Warningf("流式通道已满，丢弃完成事件")
			}
		}

		return nil
	}

	// 发送流式请求
	if err := streamClient.PostChatStream(ctx, "/v1/chat-messages", chatReq, handler); err != nil {
		klog.Errorf("发送聊天消息失败: %v", err)
		// 发送错误事件
		select {
		case streamChan <- gin.H{"event": "error", "data": fmt.Sprintf("发送消息失败: %v", err)}:
		case <-ctx.Done():
			klog.Infof("上下文已取消，停止发送错误事件")
		default:
			klog.Warningf("流式通道已满，丢弃错误事件")
		}
		return fmt.Errorf("发送消息失败: %w", err)
	}

	klog.Infof("流式聊天完成，会话ID: %s, conversation_id: %s", tempSessionID, conversationID)
	return nil
}

// 辅助函数
func generateSessionID() string {
	// 生成标准UUID格式的会话ID，符合Dify API要求
	// 格式: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	now := time.Now().UnixNano()
	rand.Seed(now)

	// 生成32位随机数
	rand1 := rand.Uint32()
	rand2 := rand.Uint32()
	rand3 := rand.Uint32()
	rand4 := rand.Uint32()

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		rand1,
		rand2&0xffff,
		rand3&0xffff,
		rand4&0xffff,
		now&0xffffffffffff)
}

func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixMilli())
}

func convertFiles(files []types.ChatFile) []httpclient.ChatFile {
	if files == nil {
		return nil
	}

	result := make([]httpclient.ChatFile, len(files))
	for i, file := range files {
		result[i] = httpclient.ChatFile{
			Type:           file.Type,
			TransferMethod: file.TransferMethod,
			URL:            file.URL,
		}
	}
	return result
}

func extractContent(event *httpclient.StreamEvent) string {
	if event.Raw != "" {
		return event.Raw
	}

	if answer, ok := event.Data["answer"].(string); ok {
		return answer
	}

	return ""
}

func serializeData(data map[string]interface{}) string {
	if data == nil {
		return ""
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	return string(bytes)
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

// GetChatSessions 获取聊天会话列表
func (c *Controller) GetChatSessions(ctx context.Context, userID string) ([]*types.ChatSession, error) {
	// 获取会话列表
	sessions, err := c.factory.Chat().GetChatSessions(ctx, userID)
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

// DeleteChatSession 删除聊天会话
func (c *Controller) DeleteChatSession(ctx context.Context, sessionID string) error {
	// 调用数据库层删除会话（同时删除相关消息）
	return c.factory.Chat().DeleteChatSession(ctx, sessionID)
}

// DeleteChatSessions 批量删除聊天会话
func (c *Controller) DeleteChatSessions(ctx context.Context, sessionIDs []string) error {
	// 调用数据库层批量删除会话（同时删除相关消息）
	return c.factory.Chat().DeleteChatSessions(ctx, sessionIDs)
}

// UpdateChatSessionTitle 更新聊天会话标题
func (c *Controller) UpdateChatSessionTitle(ctx context.Context, sessionID, title string) error {
	return c.factory.Chat().UpdateChatSessionTitle(ctx, sessionID, title)
}

// generateSessionTitleAsync 异步生成会话标题
func (c *Controller) generateSessionTitleAsync(ctx context.Context, sessionID, userQuery string) {
	// 使用标题生成器的异步方法
	c.titleGenerator.GenerateTitleAsync(ctx, userQuery, func(title string, err error) {
		if err != nil {
			klog.Errorf("生成会话标题失败: %v", err)
			return
		}

		// 更新会话标题
		if err := c.factory.Chat().UpdateChatSessionTitle(ctx, sessionID, title); err != nil {
			klog.Errorf("更新会话标题失败: %v", err)
		} else {
			klog.Infof("成功更新会话标题: sessionID=%s, title=%s", sessionID, title)
		}
	})
}
