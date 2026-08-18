package voice

import (
	"fmt"
	"sync"
	"time"

	logutil "github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/util/log"
)

// ASRCacheSink ASR缓存Sink，负责缓存ASR结果到磁盘
type ASRCacheSink struct {
	mu           sync.RWMutex
	config       ASRCacheConfig
	macAddress   string
	remoteSink   *RemoteSink
	httpClient   *HTTPBatchClient
	cacheManager *CacheManager
	metrics      ASRCacheMetrics
	closed       bool
}

// NewASRCacheSink 创建新的ASR缓存Sink
func NewASRCacheSink(config ASRCacheConfig, macAddress string, remoteSink *RemoteSink) *ASRCacheSink {
	// 创建缓存管理器
	cacheManager := NewCacheManager(config)

	// 创建HTTP客户端（如果启用）
	var httpClient *HTTPBatchClient
	if config.HTTPBatch.Enabled {
		// 从RemoteSink获取baseURL，这里需要根据实际情况调整
		baseURL := ""
		if remoteSink != nil {
			baseURL = remoteSink.GetBaseURL()
		}

		if baseURL != "" {
			httpClient = NewHTTPBatchClient(baseURL, nil, config.HTTPBatch.Timeout)
		}
	}

	return &ASRCacheSink{
		config:       config,
		macAddress:   macAddress,
		remoteSink:   remoteSink,
		httpClient:   httpClient,
		cacheManager: cacheManager,
		metrics:      ASRCacheMetrics{},
	}
}

// OnASRResult 处理ASR结果
func (s *ASRCacheSink) OnASRResult(result map[string]interface{}) {
	if s.closed {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查RemoteSink连接状态
	if s.remoteSink != nil && s.isRemoteSinkConnected() {
		// WebSocket连接正常，只进行实时上报，不缓存
		s.remoteSink.OnASRResult(result)
		logutil.Debugf("ASR cache sink: real-time report sent, not caching")
		s.metrics.RetrySuccess++
	} else {
		// WebSocket断开或RemoteSink不可用，缓存到磁盘
		logutil.Debugf("ASR cache sink: remote sink not connected, caching to disk")
		s.cacheToDisk(result, StatusOfflinePending)
	}
}

// cacheToDisk 缓存ASR结果到磁盘
func (s *ASRCacheSink) cacheToDisk(result map[string]interface{}, status int) {
	// 确保mac_address已添加到result中（与RemoteSink逻辑一致）
	if s.macAddress != "" {
		result["mac_address"] = s.macAddress
	}

	// 创建缓存数据
	cacheData := NewASRCacheData(result, status, s.config)

	// 保存到文件
	if err := cacheData.SaveToFile(s.config.CacheDir); err != nil {
		logutil.Errorf("ASR cache sink: failed to save cache file: %v", err)
		return
	}

	// 更新指标
	s.updateMetrics(status)

	logutil.Infof("ASR cache sink: cached ASR result to disk, id=%s, status=%d", cacheData.ID, status)
}

// isRemoteSinkConnected 检查RemoteSink是否连接
func (s *ASRCacheSink) isRemoteSinkConnected() bool {
	if s.remoteSink == nil {
		return false
	}

	return s.remoteSink.IsConnected()
}

// updateMetrics 更新监控指标
func (s *ASRCacheSink) updateMetrics(status int) {
	switch status {
	case StatusOfflinePending:
		s.metrics.OfflinePending++
	case StatusReportFailed:
		s.metrics.ReportFailed++
	}
	s.metrics.TotalCached++
}

// GetMetrics 获取监控指标
func (s *ASRCacheSink) GetMetrics() ASRCacheMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.metrics
}

// ResetMetrics 重置监控指标
func (s *ASRCacheSink) ResetMetrics() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metrics = ASRCacheMetrics{}
}

// Close 关闭缓存Sink
func (s *ASRCacheSink) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	logutil.Infof("ASR cache sink: closed")
}

// IsClosed 检查是否已关闭
func (s *ASRCacheSink) IsClosed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.closed
}

// GetConfig 获取配置
func (s *ASRCacheSink) GetConfig() ASRCacheConfig {
	return s.config
}

// UpdateConfig 更新配置
func (s *ASRCacheSink) UpdateConfig(config ASRCacheConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = config
	logutil.Infof("ASR cache sink: config updated")
}

// GetCacheDir 获取缓存目录
func (s *ASRCacheSink) GetCacheDir() string {
	return s.config.CacheDir
}

// GetMacAddress 获取MAC地址
func (s *ASRCacheSink) GetMacAddress() string {
	return s.macAddress
}

// SetRemoteSink 设置RemoteSink引用
func (s *ASRCacheSink) SetRemoteSink(remoteSink *RemoteSink) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.remoteSink = remoteSink
	logutil.Infof("ASR cache sink: remote sink updated")
}

// GetPendingCacheCount 获取待处理缓存数量
func (s *ASRCacheSink) GetPendingCacheCount() (int, error) {
	// 扫描缓存目录，统计待处理的文件数量
	// 这里需要实现缓存文件扫描逻辑
	return 0, nil
}

// CleanupExpiredCache 清理过期缓存
func (s *ASRCacheSink) CleanupExpiredCache() error {
	if s.closed {
		return fmt.Errorf("ASR cache sink is closed")
	}

	// 清理过期文件
	expiredCount, err := s.cacheManager.CleanupExpiredFiles()
	if err != nil {
		return fmt.Errorf("failed to cleanup expired files: %w", err)
	}

	// 清理旧文件（按数量限制）
	oldCount, err := s.cacheManager.CleanupOldFiles()
	if err != nil {
		return fmt.Errorf("failed to cleanup old files: %w", err)
	}

	// 更新指标
	s.mu.Lock()
	s.metrics.LastCleanupAt = time.Now().UnixMilli()
	s.mu.Unlock()

	if expiredCount > 0 || oldCount > 0 {
		logutil.Infof("ASR cache sink: cleanup completed, expired=%d, old=%d", expiredCount, oldCount)
	}

	return nil
}

// GetCacheFiles 获取缓存文件列表
func (s *ASRCacheSink) GetCacheFiles(status int) ([]ASRCacheData, error) {
	// 扫描缓存目录，返回指定状态的文件
	// 这里需要实现文件扫描逻辑
	return nil, nil
}

// DeleteCacheFile 删除指定缓存文件
func (s *ASRCacheSink) DeleteCacheFile(cacheID string) error {
	// 删除指定的缓存文件
	// 这里需要实现删除逻辑
	return nil
}

// RetryFailedCache 重试失败的缓存
func (s *ASRCacheSink) RetryFailedCache() error {
	if s.closed {
		return fmt.Errorf("ASR cache sink is closed")
	}

	// 获取待重试的缓存文件
	pendingFiles, err := s.cacheManager.GetPendingCacheFiles(s.config.HTTPBatch.BatchSize)
	if err != nil {
		return fmt.Errorf("failed to get pending cache files: %w", err)
	}

	if len(pendingFiles) == 0 {
		logutil.Debugf("ASR cache sink: no pending files to retry")
		return nil
	}

	logutil.Infof("ASR cache sink: retrying %d cache files", len(pendingFiles))

	// 优先尝试WebSocket实时上报
	if s.remoteSink != nil && s.isRemoteSinkConnected() {
		return s.retryWebSocketReport(pendingFiles)
	}

	// 如果WebSocket不可用，使用HTTP批量上传
	if s.httpClient != nil {
		return s.retryHTTPBatchUpload(pendingFiles)
	}

	return fmt.Errorf("no retry method available")
}

// retryWebSocketReport 通过WebSocket重试上报
func (s *ASRCacheSink) retryWebSocketReport(pendingFiles []ASRCacheData) error {
	var successIDs []string
	var failCount int

	for _, cacheFile := range pendingFiles {
		s.remoteSink.OnASRResult(cacheFile.Payload)
		logutil.Debugf("ASR cache sink: WebSocket retry sent for %s", cacheFile.ID)
		successIDs = append(successIDs, cacheFile.ID)
	}

	// 删除成功上报的文件
	if len(successIDs) > 0 {
		if err := s.cacheManager.DeleteCacheFiles(successIDs); err != nil {
			logutil.Warnf("ASR cache sink: failed to delete success files: %v", err)
		}
	}

	// 更新指标
	s.mu.Lock()
	s.metrics.RetrySuccess += int64(len(successIDs))
	s.metrics.RetryFailed += int64(failCount)
	s.mu.Unlock()

	logutil.Infof("ASR cache sink: WebSocket retry completed, success=%d, failed=%d", len(successIDs), failCount)
	return nil
}

// retryHTTPBatchUpload 通过HTTP批量上传重试
func (s *ASRCacheSink) retryHTTPBatchUpload(pendingFiles []ASRCacheData) error {
	response, err := s.httpClient.UploadBatchWithRetry(pendingFiles, s.macAddress, s.config.HTTPBatch.MaxRetryAttempts)
	if err != nil {
		// 更新重试信息
		for _, file := range pendingFiles {
			if err := s.cacheManager.UpdateCacheFileRetryInfo(file.ID); err != nil {
				logutil.Warnf("ASR cache sink: failed to update retry info for %s: %v", file.ID, err)
			}
		}
		return fmt.Errorf("HTTP batch upload failed: %w", err)
	}

	// 删除成功上传的文件
	if response.ProcessedCount > 0 {
		var successIDs []string
		for i := 0; i < response.ProcessedCount && i < len(pendingFiles); i++ {
			successIDs = append(successIDs, pendingFiles[i].ID)
		}

		if err := s.cacheManager.DeleteCacheFiles(successIDs); err != nil {
			logutil.Warnf("ASR cache sink: failed to delete success files: %v", err)
		}
	}

	// 更新指标
	s.mu.Lock()
	s.metrics.RetrySuccess += int64(response.ProcessedCount)
	s.metrics.RetryFailed += int64(len(response.FailedItems))
	s.metrics.LastUploadAt = time.Now().UnixMilli()
	s.mu.Unlock()

	logutil.Infof("ASR cache sink: HTTP batch upload completed, success=%d, failed=%d",
		response.ProcessedCount, len(response.FailedItems))
	return nil
}

// GetStatus 获取缓存状态信息
func (s *ASRCacheSink) GetStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := map[string]interface{}{
		"enabled":          s.config.Enabled,
		"cache_dir":        s.config.CacheDir,
		"max_retries":      s.config.MaxRetries,
		"retry_interval":   s.config.RetryInterval.String(),
		"cache_expiry":     s.config.CacheExpiry.String(),
		"max_cache_size":   s.config.MaxCacheSize,
		"cleanup_interval": s.config.CleanupInterval.String(),
		"mac_address":      s.macAddress,
		"closed":           s.closed,
		"metrics":          s.metrics,
	}

	return status
}

// GetCacheManager 获取缓存管理器
func (s *ASRCacheSink) GetCacheManager() *CacheManager {
	return s.cacheManager
}

// GetRemoteSink 获取远程Sink
func (s *ASRCacheSink) GetRemoteSink() *RemoteSink {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.remoteSink
}
