package voice

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	logutil "github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/util/log"
)

// CacheManager 缓存文件管理器
type CacheManager struct {
	config ASRCacheConfig
}

// NewCacheManager 创建新的缓存管理器
func NewCacheManager(config ASRCacheConfig) *CacheManager {
	return &CacheManager{
		config: config,
	}
}

// ScanCacheFiles 扫描缓存目录，返回所有缓存文件
func (m *CacheManager) ScanCacheFiles() ([]ASRCacheData, error) {
	var cacheFiles []ASRCacheData

	// 确保缓存目录存在
	if err := os.MkdirAll(m.config.CacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	// 读取目录
	entries, err := os.ReadDir(m.config.CacheDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache directory: %w", err)
	}

	// 遍历文件
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// 只处理.json文件
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		// 读取缓存文件
		filePath := filepath.Join(m.config.CacheDir, entry.Name())
		cacheData, err := LoadFromFile(filePath)
		if err != nil {
			logutil.Warnf("Cache manager: failed to load cache file %s: %v", filePath, err)
			continue
		}

		cacheFiles = append(cacheFiles, *cacheData)
	}

	// 按创建时间排序
	sort.Slice(cacheFiles, func(i, j int) bool {
		return cacheFiles[i].Timestamp < cacheFiles[j].Timestamp
	})

	return cacheFiles, nil
}

// GetCacheFilesByStatus 根据状态获取缓存文件
func (m *CacheManager) GetCacheFilesByStatus(status int) ([]ASRCacheData, error) {
	allFiles, err := m.ScanCacheFiles()
	if err != nil {
		return nil, err
	}

	var filteredFiles []ASRCacheData
	for _, file := range allFiles {
		if file.Status == status {
			filteredFiles = append(filteredFiles, file)
		}
	}

	return filteredFiles, nil
}

// GetPendingCacheFiles 获取待处理的缓存文件
func (m *CacheManager) GetPendingCacheFiles(limit int) ([]ASRCacheData, error) {
	allFiles, err := m.ScanCacheFiles()
	if err != nil {
		return nil, err
	}

	var pendingFiles []ASRCacheData
	for _, file := range allFiles {
		// 只处理待上报和上报失败的文件
		if (file.Status == StatusOfflinePending || file.Status == StatusReportFailed) &&
			file.CanRetry(m.config.MaxRetries) && file.ShouldRetry() {
			pendingFiles = append(pendingFiles, file)
		}

		// 限制数量
		if limit > 0 && len(pendingFiles) >= limit {
			break
		}
	}

	return pendingFiles, nil
}

// DeleteCacheFile 删除指定的缓存文件
func (m *CacheManager) DeleteCacheFile(cacheID string) error {
	filePath := filepath.Join(m.config.CacheDir, fmt.Sprintf("%s.json", cacheID))

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			logutil.Debugf("Cache manager: cache file %s not found", filePath)
			return nil
		}
		return fmt.Errorf("failed to delete cache file %s: %w", filePath, err)
	}

	logutil.Debugf("Cache manager: deleted cache file %s", filePath)
	return nil
}

// DeleteCacheFiles 批量删除缓存文件
func (m *CacheManager) DeleteCacheFiles(cacheIDs []string) error {
	var errors []string

	for _, cacheID := range cacheIDs {
		if err := m.DeleteCacheFile(cacheID); err != nil {
			errors = append(errors, fmt.Sprintf("failed to delete %s: %v", cacheID, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to delete some cache files: %s", strings.Join(errors, "; "))
	}

	return nil
}

// CleanupExpiredFiles 清理过期文件
func (m *CacheManager) CleanupExpiredFiles() (int, error) {
	allFiles, err := m.ScanCacheFiles()
	if err != nil {
		return 0, err
	}

	var expiredIDs []string

	for _, file := range allFiles {
		if file.IsExpired() || file.Status == StatusExpired {
			expiredIDs = append(expiredIDs, file.ID)
		}
	}

	if len(expiredIDs) == 0 {
		return 0, nil
	}

	// 删除过期文件
	if err := m.DeleteCacheFiles(expiredIDs); err != nil {
		return 0, err
	}

	logutil.Infof("Cache manager: cleaned up %d expired cache files", len(expiredIDs))
	return len(expiredIDs), nil
}

// CleanupOldFiles 清理旧文件（按数量限制）
func (m *CacheManager) CleanupOldFiles() (int, error) {
	if m.config.MaxCacheSize <= 0 {
		return 0, nil
	}

	allFiles, err := m.ScanCacheFiles()
	if err != nil {
		return 0, err
	}

	if len(allFiles) <= m.config.MaxCacheSize {
		return 0, nil
	}

	// 按时间排序，删除最旧的文件
	sort.Slice(allFiles, func(i, j int) bool {
		return allFiles[i].Timestamp < allFiles[j].Timestamp
	})

	// 计算需要删除的文件数量
	deleteCount := len(allFiles) - m.config.MaxCacheSize
	var deleteIDs []string

	for i := 0; i < deleteCount; i++ {
		deleteIDs = append(deleteIDs, allFiles[i].ID)
	}

	// 删除旧文件
	if err := m.DeleteCacheFiles(deleteIDs); err != nil {
		return 0, err
	}

	logutil.Infof("Cache manager: cleaned up %d old cache files", deleteCount)
	return deleteCount, nil
}

// GetCacheStatistics 获取缓存统计信息
func (m *CacheManager) GetCacheStatistics() (map[string]interface{}, error) {
	allFiles, err := m.ScanCacheFiles()
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_files":      len(allFiles),
		"offline_pending":  0,
		"report_failed":    0,
		"expired":          0,
		"online_reported":  0,
		"total_size_bytes": 0,
		"oldest_file":      nil,
		"newest_file":      nil,
	}

	var oldestTime, newestTime int64
	var totalSize int64

	for _, file := range allFiles {
		// 统计各状态文件数量
		switch file.Status {
		case StatusOfflinePending:
			stats["offline_pending"] = stats["offline_pending"].(int) + 1
		case StatusReportFailed:
			stats["report_failed"] = stats["report_failed"].(int) + 1
		case StatusExpired:
			stats["expired"] = stats["expired"].(int) + 1
		case StatusOnlineReported:
			stats["online_reported"] = stats["online_reported"].(int) + 1
		}

		// 统计时间范围
		if oldestTime == 0 || file.Timestamp < oldestTime {
			oldestTime = file.Timestamp
		}
		if newestTime == 0 || file.Timestamp > newestTime {
			newestTime = file.Timestamp
		}

		// 统计文件大小
		filePath := file.GetCacheFilePath(m.config.CacheDir)
		if info, err := os.Stat(filePath); err == nil {
			totalSize += info.Size()
		}
	}

	stats["total_size_bytes"] = totalSize

	if oldestTime > 0 {
		stats["oldest_file"] = time.UnixMilli(oldestTime).Format(time.RFC3339)
	}
	if newestTime > 0 {
		stats["newest_file"] = time.UnixMilli(newestTime).Format(time.RFC3339)
	}

	return stats, nil
}

// UpdateCacheFileStatus 更新缓存文件状态
func (m *CacheManager) UpdateCacheFileStatus(cacheID string, status int) error {
	filePath := filepath.Join(m.config.CacheDir, fmt.Sprintf("%s.json", cacheID))

	cacheData, err := LoadFromFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to load cache file: %w", err)
	}

	cacheData.Status = status

	// 保存更新后的文件
	if err := cacheData.SaveToFile(m.config.CacheDir); err != nil {
		return fmt.Errorf("failed to save updated cache file: %w", err)
	}

	logutil.Debugf("Cache manager: updated cache file %s status to %d", cacheID, status)
	return nil
}

// UpdateCacheFileRetryInfo 更新缓存文件重试信息
func (m *CacheManager) UpdateCacheFileRetryInfo(cacheID string) error {
	filePath := filepath.Join(m.config.CacheDir, fmt.Sprintf("%s.json", cacheID))

	cacheData, err := LoadFromFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to load cache file: %w", err)
	}

	cacheData.UpdateRetryInfo(m.config.RetryInterval)

	// 保存更新后的文件
	if err := cacheData.SaveToFile(m.config.CacheDir); err != nil {
		return fmt.Errorf("failed to save updated cache file: %w", err)
	}

	logutil.Debugf("Cache manager: updated cache file %s retry info", cacheID)
	return nil
}

// GetCacheFile 获取指定的缓存文件
func (m *CacheManager) GetCacheFile(cacheID string) (*ASRCacheData, error) {
	filePath := filepath.Join(m.config.CacheDir, fmt.Sprintf("%s.json", cacheID))
	return LoadFromFile(filePath)
}

// IsCacheFileExists 检查缓存文件是否存在
func (m *CacheManager) IsCacheFileExists(cacheID string) bool {
	filePath := filepath.Join(m.config.CacheDir, fmt.Sprintf("%s.json", cacheID))
	_, err := os.Stat(filePath)
	return err == nil
}
