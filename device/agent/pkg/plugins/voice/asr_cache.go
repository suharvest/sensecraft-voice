package voice

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ASR缓存状态常量
const (
	StatusOnlineReported = 0 // 在线已上报
	StatusOfflinePending = 1 // 离线待上报
	StatusReportFailed   = 2 // 上报失败
	StatusExpired        = 3 // 已过期
)

// ASRCacheData ASR缓存数据结构
type ASRCacheData struct {
	ID          string                 `json:"id"`
	Timestamp   int64                  `json:"timestamp"`
	CreatedAt   string                 `json:"created_at"`
	Status      int                    `json:"status"`
	RetryCount  int                    `json:"retry_count"`
	LastRetryAt *int64                 `json:"last_retry_at"`
	NextRetryAt int64                  `json:"next_retry_at"`
	ExpiresAt   int64                  `json:"expires_at"`
	Payload     map[string]interface{} `json:"payload"` // 完整的ASR结果
}

// ASRCacheConfig ASR缓存配置
type ASRCacheConfig struct {
	Enabled         bool          `yaml:"enabled"`
	CacheDir        string        `yaml:"cache_dir"`
	MaxRetries      int           `yaml:"max_retries"`
	RetryInterval   time.Duration `yaml:"retry_interval"`
	CacheExpiry     time.Duration `yaml:"cache_expiry"`
	MaxCacheSize    int           `yaml:"max_cache_size"`
	CleanupInterval time.Duration `yaml:"cleanup_interval"`

	// HTTP批量上传配置
	HTTPBatch HTTPBatchConfig `yaml:"http_batch"`
}

// HTTPBatchConfig HTTP批量上传配置
type HTTPBatchConfig struct {
	Enabled          bool          `yaml:"enabled"`
	BatchSize        int           `yaml:"batch_size"`
	UploadInterval   time.Duration `yaml:"upload_interval"`
	MaxRetryAttempts int           `yaml:"max_retry_attempts"`
	Timeout          time.Duration `yaml:"timeout"`
}

// ASRCacheMetrics 缓存监控指标
type ASRCacheMetrics struct {
	TotalCached    int64 `json:"total_cached"`
	OfflinePending int64 `json:"offline_pending"`
	ReportFailed   int64 `json:"report_failed"`
	Expired        int64 `json:"expired"`
	RetrySuccess   int64 `json:"retry_success"`
	RetryFailed    int64 `json:"retry_failed"`
	LastCleanupAt  int64 `json:"last_cleanup_at"`
	LastUploadAt   int64 `json:"last_upload_at"`
}

// NewASRCacheData 创建新的ASR缓存数据
func NewASRCacheData(result map[string]interface{}, status int, config ASRCacheConfig) *ASRCacheData {
	now := time.Now()

	return &ASRCacheData{
		ID:          generateCacheID(now),
		Timestamp:   now.UnixMilli(),
		CreatedAt:   now.Format(time.RFC3339),
		Status:      status,
		RetryCount:  0,
		LastRetryAt: nil,
		NextRetryAt: now.Add(config.RetryInterval).UnixMilli(),
		ExpiresAt:   now.Add(config.CacheExpiry).UnixMilli(),
		Payload:     result,
	}
}

// generateCacheID 生成缓存ID
func generateCacheID(t time.Time) string {
	return fmt.Sprintf("asr_%d_%s", t.Unix(), generateShortUUID())
}

// generateShortUUID 生成短UUID
func generateShortUUID() string {
	// 简单的短UUID生成，实际项目中可以使用更复杂的算法
	return fmt.Sprintf("%x", time.Now().UnixNano())[:8]
}

// IsExpired 检查是否已过期
// 优化：暂时禁用过期检查，确保不漏数据
func (c *ASRCacheData) IsExpired() bool {
	return false // 暂时禁用过期检查，确保数据安全
}

// CanRetry 检查是否可以重试
// 优化：去掉所有限制，确保所有数据都会被重试
func (c *ASRCacheData) CanRetry(maxRetries int) bool {
	return true // 去掉过期和重试次数限制，确保不漏数据
}

// ShouldRetry 检查是否应该重试
func (c *ASRCacheData) ShouldRetry() bool {
	return time.Now().UnixMilli() >= c.NextRetryAt
}

// UpdateRetryInfo 更新重试信息
func (c *ASRCacheData) UpdateRetryInfo(retryInterval time.Duration) {
	now := time.Now()
	nowUnix := now.UnixMilli()
	c.RetryCount++
	c.LastRetryAt = &nowUnix
	c.NextRetryAt = now.Add(retryInterval).UnixMilli()
}

// ToJSON 转换为JSON字符串
func (c *ASRCacheData) ToJSON() ([]byte, error) {
	return json.Marshal(c)
}

// FromJSON 从JSON字符串创建
func FromJSON(data []byte) (*ASRCacheData, error) {
	var cache ASRCacheData
	err := json.Unmarshal(data, &cache)
	return &cache, err
}

// GetCacheFilePath 获取缓存文件路径
func (c *ASRCacheData) GetCacheFilePath(cacheDir string) string {
	return filepath.Join(cacheDir, fmt.Sprintf("%s.json", c.ID))
}

// SaveToFile 保存到文件
func (c *ASRCacheData) SaveToFile(cacheDir string) error {
	// 确保目录存在
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// 转换为JSON
	data, err := c.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	// 写入文件
	filePath := c.GetCacheFilePath(cacheDir)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// LoadFromFile 从文件加载
func LoadFromFile(filePath string) (*ASRCacheData, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	return FromJSON(data)
}

// DeleteFile 删除缓存文件
func (c *ASRCacheData) DeleteFile(cacheDir string) error {
	filePath := c.GetCacheFilePath(cacheDir)
	return os.Remove(filePath)
}
