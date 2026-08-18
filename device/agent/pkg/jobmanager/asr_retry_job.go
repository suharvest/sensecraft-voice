package jobmanager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/plugins/voice"
	"github.com/sirupsen/logrus"
)

// ASRRetryJob ASR重试任务
type ASRRetryJob struct {
	config  voice.ASRCacheConfig
	metrics *voice.ASRCacheMetrics
	// 添加获取baseURL的函数
	getBaseURL func() string
	// 优化：添加原子性处理相关字段
	mu        sync.Mutex // 互斥锁，确保任务执行原子性
	isRunning bool       // 执行状态标志
	lockFile  *os.File   // 文件锁句柄
}

// NewASRRetryJob 创建新的ASR重试任务
func NewASRRetryJob(config voice.ASRCacheConfig) *ASRRetryJob {
	return &ASRRetryJob{
		config:     config,
		metrics:    &voice.ASRCacheMetrics{},
		getBaseURL: func() string { return "" }, // 默认返回空字符串
	}
}

// NewASRRetryJobWithBaseURL 创建带baseURL获取函数的ASR重试任务
func NewASRRetryJobWithBaseURL(config voice.ASRCacheConfig, getBaseURL func() string) *ASRRetryJob {
	return &ASRRetryJob{
		config:     config,
		metrics:    &voice.ASRCacheMetrics{},
		getBaseURL: getBaseURL,
	}
}

// Name 返回任务名称
func (j *ASRRetryJob) Name() string {
	return "asr-retry-job"
}

// CronSpec 返回cron表达式
func (j *ASRRetryJob) CronSpec() string {
	// 每30秒执行一次
	return "@every 5s"
}

// Do 执行重试任务
func (j *ASRRetryJob) Do(ctx *JobContext) error {
	if !j.config.Enabled {
		return nil
	}

	// 优化：添加原子性检查，防止并发执行
	j.mu.Lock()
	if j.isRunning {
		j.mu.Unlock()
		logrus.Warn("ASR重试任务正在执行中，跳过本次执行")
		return nil
	}
	j.isRunning = true
	j.mu.Unlock()

	// 确保在函数结束时重置执行状态
	defer func() {
		j.mu.Lock()
		j.isRunning = false
		j.mu.Unlock()
	}()

	// 直接处理缓存文件，不依赖Voice Manager
	if err := j.ProcessCacheFilesDirectly(); err != nil {
		logrus.Error(fmt.Errorf("直接处理缓存文件失败: %w", err))
	}

	return nil
}

// ProcessCacheFilesDirectly 直接处理缓存文件
func (j *ASRRetryJob) ProcessCacheFilesDirectly() error {
	// 优化：使用文件锁确保原子性处理
	if err := j.acquireFileLock(); err != nil {
		return fmt.Errorf("获取文件锁失败: %w", err)
	}
	defer j.releaseFileLock()

	// 扫描缓存目录
	cacheFiles, err := j.scanCacheFiles()
	if err != nil {
		return fmt.Errorf("扫描缓存文件失败: %w", err)
	}

	if len(cacheFiles) == 0 {
		return nil
	}

	// 过滤出需要重试的文件
	// 优化：简化过滤逻辑，确保不漏数据
	var pendingFiles []voice.ASRCacheData
	for _, file := range cacheFiles {
		// 只检查状态，不检查过期时间和重试次数限制
		if file.Status == voice.StatusOfflinePending || file.Status == voice.StatusReportFailed {
			pendingFiles = append(pendingFiles, file)
		}
	}

	if len(pendingFiles) == 0 {
		return nil
	}

	logrus.Infof("找到 %d 个待重试的缓存文件", len(pendingFiles))

	// 优化：逐个处理文件，成功即删除，失败即更新重试信息
	var successCount int
	var failedCount int

	for _, cacheFile := range pendingFiles {
		if err := j.uploadCacheFile(cacheFile); err != nil {
			// 上传失败：更新重试信息，保留文件
			logrus.Warnf("上传缓存文件失败，文件ID: %s, 错误: %v", cacheFile.ID, err)
			if err := j.updateFailedFile(cacheFile); err != nil {
				logrus.Warnf("更新失败文件重试信息失败，文件ID: %s, 错误: %v", cacheFile.ID, err)
			}
			failedCount++
		} else {
			// 上传成功：立即删除文件
			if err := j.deleteCacheFile(cacheFile.ID); err != nil {
				logrus.Warnf("删除成功上传文件失败，文件ID: %s, 错误: %v", cacheFile.ID, err)
			} else {
				logrus.Debugf("成功上传并删除缓存文件: %s", cacheFile.ID)
			}
			successCount++
		}
	}

	// 更新指标
	j.metrics.RetrySuccess += int64(successCount)
	j.metrics.RetryFailed += int64(failedCount)
	j.metrics.LastUploadAt = time.Now().UnixMilli()

	logrus.Infof("缓存文件处理完成: 成功 %d, 失败 %d", successCount, failedCount)

	return nil
}

// acquireFileLock 获取文件锁
func (j *ASRRetryJob) acquireFileLock() error {
	lockFilePath := filepath.Join(j.config.CacheDir, ".asr_retry.lock")

	// 确保缓存目录存在
	if err := os.MkdirAll(j.config.CacheDir, 0755); err != nil {
		return fmt.Errorf("创建缓存目录失败: %w", err)
	}

	// 打开或创建锁文件
	lockFile, err := os.OpenFile(lockFilePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("创建锁文件失败: %w", err)
	}

	// 尝试获取排他锁
	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		lockFile.Close()
		return fmt.Errorf("获取文件锁失败，可能有其他实例正在运行: %w", err)
	}

	j.lockFile = lockFile
	logrus.Debug("成功获取文件锁")
	return nil
}

// releaseFileLock 释放文件锁
func (j *ASRRetryJob) releaseFileLock() {
	if j.lockFile != nil {
		syscall.Flock(int(j.lockFile.Fd()), syscall.LOCK_UN)
		j.lockFile.Close()
		j.lockFile = nil
		logrus.Debug("成功释放文件锁")
	}
}

// scanCacheFiles 扫描缓存目录
func (j *ASRRetryJob) scanCacheFiles() ([]voice.ASRCacheData, error) {
	// 确保缓存目录存在
	if err := os.MkdirAll(j.config.CacheDir, 0755); err != nil {
		return nil, fmt.Errorf("创建缓存目录失败: %w", err)
	}

	// 读取目录
	entries, err := os.ReadDir(j.config.CacheDir)
	if err != nil {
		return nil, fmt.Errorf("读取缓存目录失败: %w", err)
	}

	var cacheFiles []voice.ASRCacheData

	// 遍历文件
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// 只处理.json文件
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		// 读取文件
		filePath := filepath.Join(j.config.CacheDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			logrus.Warnf("读取缓存文件失败: %s, 错误: %v", entry.Name(), err)
			continue
		}

		// 解析JSON
		var cacheFile voice.ASRCacheData
		if err := json.Unmarshal(data, &cacheFile); err != nil {
			logrus.Warnf("解析缓存文件失败: %s, 错误: %v", entry.Name(), err)
			continue
		}

		cacheFiles = append(cacheFiles, cacheFile)
	}

	return cacheFiles, nil
}

// uploadCacheFile 上传单个缓存文件
func (j *ASRRetryJob) uploadCacheFile(cacheFile voice.ASRCacheData) error {
	// 构造请求数据
	requestData := make(map[string]interface{})

	// 复制原始ASR结果
	for key, value := range cacheFile.Payload {
		requestData[key] = value
	}

	// 确保包含必要字段
	if _, exists := requestData["mac_address"]; !exists {
		requestData["mac_address"] = "6e:8e:84:f9:73:d6" // 默认MAC地址
	}

	if _, exists := requestData["timestamp"]; !exists {
		requestData["timestamp"] = time.Now().UnixMilli()
	}

	if _, exists := requestData["status"]; !exists {
		requestData["status"] = 1 // 离线重试状态
	}

	// 序列化请求数据
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("序列化ASR结果失败: %w", err)
	}

	// 获取当前配置的baseURL
	baseURL := j.getBaseURL()
	if baseURL == "" {
		return fmt.Errorf("远程服务地址未配置")
	}

	// 构造HTTP请求
	url := baseURL + "/api/v1/recordings"
	req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "SenseCraft-Voice-Client/1.0")

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: j.config.HTTPBatch.Timeout,
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("HTTP请求失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// deleteCacheFile 删除单个缓存文件
func (j *ASRRetryJob) deleteCacheFile(cacheID string) error {
	filePath := filepath.Join(j.config.CacheDir, fmt.Sprintf("%s.json", cacheID))
	if err := os.Remove(filePath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("删除缓存文件失败: %w", err)
		}
	}
	return nil
}

// deleteCacheFiles 删除缓存文件（批量版本，保留向后兼容）
func (j *ASRRetryJob) deleteCacheFiles(cacheIDs []string) error {
	for _, cacheID := range cacheIDs {
		filePath := filepath.Join(j.config.CacheDir, fmt.Sprintf("%s.json", cacheID))
		if err := os.Remove(filePath); err != nil {
			if !os.IsNotExist(err) {
				logrus.Warnf("删除缓存文件失败: %s, 错误: %v", filePath, err)
			}
		}
	}
	return nil
}

// updateFailedFile 更新单个失败文件的重试信息
func (j *ASRRetryJob) updateFailedFile(cacheFile voice.ASRCacheData) error {
	filePath := filepath.Join(j.config.CacheDir, fmt.Sprintf("%s.json", cacheFile.ID))

	// 更新重试信息
	cacheFile.UpdateRetryInfo(j.config.RetryInterval)

	// 写回文件
	newData, err := json.MarshalIndent(cacheFile, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化缓存文件失败: %w", err)
	}

	if err := os.WriteFile(filePath, newData, 0644); err != nil {
		return fmt.Errorf("写入缓存文件失败: %w", err)
	}

	return nil
}

// updateFailedFiles 更新失败文件的重试信息（批量版本，保留向后兼容）
func (j *ASRRetryJob) updateFailedFiles(cacheIDs []string) error {
	for _, cacheID := range cacheIDs {
		filePath := filepath.Join(j.config.CacheDir, fmt.Sprintf("%s.json", cacheID))

		// 读取文件
		data, err := os.ReadFile(filePath)
		if err != nil {
			logrus.Warnf("读取缓存文件失败: %s, 错误: %v", cacheID, err)
			continue
		}

		// 解析JSON
		var cacheFile voice.ASRCacheData
		if err := json.Unmarshal(data, &cacheFile); err != nil {
			logrus.Warnf("解析缓存文件失败: %s, 错误: %v", cacheID, err)
			continue
		}

		// 更新重试信息
		cacheFile.UpdateRetryInfo(j.config.RetryInterval)

		// 写回文件
		newData, err := json.MarshalIndent(cacheFile, "", "  ")
		if err != nil {
			logrus.Warnf("序列化缓存文件失败: %s, 错误: %v", cacheID, err)
			continue
		}

		if err := os.WriteFile(filePath, newData, 0644); err != nil {
			logrus.Warnf("写入缓存文件失败: %s, 错误: %v", cacheID, err)
			continue
		}
	}
	return nil
}

// retryWebSocketReport 重试WebSocket上报
func (j *ASRRetryJob) retryWebSocketReport(cacheManager *voice.CacheManager, remoteSink *voice.RemoteSink) error {
	// 获取待重试的缓存文件
	pendingFiles, err := cacheManager.GetPendingCacheFiles(50) // 每次最多处理50个文件
	if err != nil {
		return fmt.Errorf("获取待重试缓存文件失败: %w", err)
	}

	if len(pendingFiles) == 0 {
		return nil
	}

	var successCount int
	var successIDs []string

	for _, cacheFile := range pendingFiles {
		// 尝试通过WebSocket上报
		remoteSink.OnASRResult(cacheFile.Payload)
		successCount++
		successIDs = append(successIDs, cacheFile.ID)
	}

	// 删除成功上报的文件
	if len(successIDs) > 0 {
		if err := cacheManager.DeleteCacheFiles(successIDs); err != nil {
			logrus.Warnf("删除成功上报文件失败: %v", err)
		}
	}

	// 更新指标
	j.metrics.RetrySuccess += int64(successCount)

	return nil
}

// retryHTTPBatchUpload 重试HTTP批量上传
func (j *ASRRetryJob) retryHTTPBatchUpload(cacheManager *voice.CacheManager, remoteSink *voice.RemoteSink) error {
	// 获取待重试的缓存文件
	pendingFiles, err := cacheManager.GetPendingCacheFiles(j.config.HTTPBatch.BatchSize)
	if err != nil {
		return fmt.Errorf("获取待重试缓存文件失败: %w", err)
	}

	if len(pendingFiles) == 0 {
		return nil
	}

	var successCount int
	var successIDs []string
	var failedIDs []string

	// 逐个发送ASR结果到 /api/v1/recordings 接口
	for _, cacheFile := range pendingFiles {
		if err := j.sendASRResultToAPI(cacheFile.Payload, remoteSink); err != nil {
			logrus.Warnf("发送ASR结果失败，文件ID: %s, 错误: %v", cacheFile.ID, err)
			failedIDs = append(failedIDs, cacheFile.ID)
		} else {
			successCount++
			successIDs = append(successIDs, cacheFile.ID)
		}
	}

	// 删除成功上报的文件
	if len(successIDs) > 0 {
		if err := cacheManager.DeleteCacheFiles(successIDs); err != nil {
			logrus.Warnf("删除成功上报文件失败: %v", err)
		}
	}

	// 更新失败文件的重试信息
	if len(failedIDs) > 0 {
		for _, fileID := range failedIDs {
			if err := cacheManager.UpdateCacheFileRetryInfo(fileID); err != nil {
				logrus.Warnf("更新重试信息失败，文件ID: %s, 错误: %v", fileID, err)
			}
		}
	}

	// 更新指标
	j.metrics.RetrySuccess += int64(successCount)
	j.metrics.RetryFailed += int64(len(failedIDs))
	j.metrics.LastUploadAt = time.Now().UnixMilli()

	return nil
}

// sendASRResultToAPI 发送ASR结果到API接口
func (j *ASRRetryJob) sendASRResultToAPI(asrResult map[string]interface{}, remoteSink *voice.RemoteSink) error {
	// 获取baseURL，优先从remoteSink获取
	baseURL := ""
	if remoteSink != nil {
		baseURL = remoteSink.GetBaseURL()
	}

	if baseURL == "" {
		return fmt.Errorf("HTTP上传缺少基础URL")
	}

	// 构造请求数据，确保包含必要的字段
	requestData := make(map[string]interface{})

	// 复制原始ASR结果
	for key, value := range asrResult {
		requestData[key] = value
	}

	// 确保包含mac_address
	if _, exists := requestData["mac_address"]; !exists {
		if remoteSink != nil {
			requestData["mac_address"] = remoteSink.GetMacAddress()
		}
	}

	// 确保包含timestamp
	if _, exists := requestData["timestamp"]; !exists {
		requestData["timestamp"] = time.Now().UnixMilli()
	}

	// 确保包含status字段，离线重试的status应该为1
	if _, exists := requestData["status"]; !exists {
		requestData["status"] = 1
	}

	// 序列化请求数据
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("序列化ASR结果失败: %w", err)
	}

	// 构造HTTP请求
	url := baseURL + "/api/v1/recordings"
	req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "SenseCraft-Voice-Client/1.0")

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: j.config.HTTPBatch.Timeout,
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("HTTP请求失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

// cleanupExpiredFiles 清理过期文件
func (j *ASRRetryJob) cleanupExpiredFiles(cacheManager *voice.CacheManager) error {
	_, err := cacheManager.CleanupExpiredFiles()
	if err != nil {
		return fmt.Errorf("清理过期文件失败: %w", err)
	}

	return nil
}

// cleanupOldFiles 清理旧文件
func (j *ASRRetryJob) cleanupOldFiles(cacheManager *voice.CacheManager) error {
	_, err := cacheManager.CleanupOldFiles()
	if err != nil {
		return fmt.Errorf("清理旧文件失败: %w", err)
	}

	return nil
}

// isRemoteSinkConnected 检查RemoteSink是否连接
func (j *ASRRetryJob) isRemoteSinkConnected(remoteSink *voice.RemoteSink) bool {
	if remoteSink == nil {
		return false
	}

	return remoteSink.IsConnected()
}

// GetMetrics 获取任务指标
func (j *ASRRetryJob) GetMetrics() voice.ASRCacheMetrics {
	return *j.metrics
}

// UpdateConfig 更新配置
func (j *ASRRetryJob) UpdateConfig(config voice.ASRCacheConfig) {
	j.config = config
}

// UpdateBaseURLGetter 更新baseURL获取函数
func (j *ASRRetryJob) UpdateBaseURLGetter(getBaseURL func() string) {
	j.getBaseURL = getBaseURL
}

// GetStatus 获取任务状态
func (j *ASRRetryJob) GetStatus() map[string]interface{} {
	// 获取Voice Manager中的组件
	voiceManager := voice.GetManager()
	asrCacheSink := voiceManager.GetASRCacheSink()

	status := map[string]interface{}{
		"job_name":           j.Name(),
		"cron_spec":          j.CronSpec(),
		"enabled":            j.config.Enabled,
		"http_batch_enabled": j.config.HTTPBatch.Enabled,
		"metrics":            j.metrics,
		"is_running":         j.isRunning,
	}

	// 添加缓存目录大小监控
	cacheSize, err := j.getCacheDirectorySize()
	if err != nil {
		logrus.Warnf("获取缓存目录大小失败: %v", err)
		status["cache_size_bytes"] = 0
		status["cache_size_mb"] = 0
	} else {
		status["cache_size_bytes"] = cacheSize
		status["cache_size_mb"] = float64(cacheSize) / (1024 * 1024)
	}

	if asrCacheSink != nil {
		cacheManager := asrCacheSink.GetCacheManager()
		remoteSink := asrCacheSink.GetRemoteSink()

		// 获取缓存统计信息
		stats, err := cacheManager.GetCacheStatistics()
		if err != nil {
			stats = make(map[string]interface{})
		}

		status["remote_connected"] = j.isRemoteSinkConnected(remoteSink)
		status["cache_stats"] = stats
	} else {
		status["remote_connected"] = false
		status["cache_stats"] = make(map[string]interface{})
	}

	return status
}

// getCacheDirectorySize 获取缓存目录大小
func (j *ASRRetryJob) getCacheDirectorySize() (int64, error) {
	var size int64

	err := filepath.Walk(j.config.CacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// ForceRetry 强制重试所有待处理文件
func (j *ASRRetryJob) ForceRetry() error {
	// 获取Voice Manager中的组件
	voiceManager := voice.GetManager()
	asrCacheSink := voiceManager.GetASRCacheSink()
	if asrCacheSink == nil {
		return fmt.Errorf("ASR缓存未初始化")
	}

	cacheManager := asrCacheSink.GetCacheManager()
	remoteSink := asrCacheSink.GetRemoteSink()

	// 获取所有待处理文件
	pendingFiles, err := cacheManager.GetCacheFilesByStatus(voice.StatusOfflinePending)
	if err != nil {
		return fmt.Errorf("获取离线待处理文件失败: %w", err)
	}

	failedFiles, err := cacheManager.GetCacheFilesByStatus(voice.StatusReportFailed)
	if err != nil {
		return fmt.Errorf("获取上报失败文件失败: %w", err)
	}

	allFiles := append(pendingFiles, failedFiles...)

	if len(allFiles) == 0 {
		return nil
	}

	// 执行重试
	if j.config.HTTPBatch.Enabled {
		return j.retryHTTPBatchUpload(cacheManager, remoteSink)
	} else if remoteSink != nil && j.isRemoteSinkConnected(remoteSink) {
		return j.retryWebSocketReport(cacheManager, remoteSink)
	}

	return fmt.Errorf("没有可用的重试方法")
}

// IsRunning 检查任务是否正在运行
func (j *ASRRetryJob) IsRunning() bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.isRunning
}

// Cleanup 清理资源
func (j *ASRRetryJob) Cleanup() {
	// 释放文件锁
	j.releaseFileLock()

	// 重置运行状态
	j.mu.Lock()
	j.isRunning = false
	j.mu.Unlock()

	logrus.Debug("ASR重试任务资源清理完成")
}
