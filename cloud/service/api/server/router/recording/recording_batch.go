package recording

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	ctrl "github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller/recording"
)

// batchDeviceInfo 与 client 的 voice.DeviceInfo 对齐
type batchDeviceInfo struct {
	MacAddress    string `json:"mac_address"`
	ClientVersion string `json:"client_version"`
	Timestamp     int64  `json:"timestamp"`
}

// batchItem 与 client 的 voice.ASRCacheData 对齐；Payload 是完整的 ASR 结果
type batchItem struct {
	ID        string          `json:"id"`
	Timestamp int64           `json:"timestamp"`
	CreatedAt string          `json:"created_at"`
	Status    int             `json:"status"`
	Payload   json.RawMessage `json:"payload"`
}

// batchUploadRequest 客户端离线缓存批量上报请求
// 报文格式见 sensecraft-voice-client/pkg/plugins/voice/http_batch_client.go
type batchUploadRequest struct {
	DeviceInfo batchDeviceInfo `json:"device_info"`
	BatchData  []batchItem     `json:"batch_data"`
}

// batchUploadResponse 客户端直接按此结构解析响应体（不套 httputils.Response）
type batchUploadResponse struct {
	Success        bool     `json:"success"`
	Message        string   `json:"message"`
	ProcessedCount int      `json:"processed_count"`
	FailedItems    []string `json:"failed_items"`
	Timestamp      int64    `json:"timestamp"`
}

// saveBatch POST /api/v1/recordings/batch
// 客户端在设备离线期间把 ASR 结果缓存在本地，恢复后批量补报。
func (r *recordingRouter) saveBatch(c *gin.Context) {
	var req batchUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, batchUploadResponse{
			Success:     false,
			Message:     err.Error(),
			FailedItems: []string{},
			Timestamp:   time.Now().UnixMilli(),
		})
		return
	}

	defaultMac := strings.ToLower(strings.TrimSpace(req.DeviceInfo.MacAddress))
	failed := make([]string, 0, len(req.BatchData))
	processed := 0

	for _, item := range req.BatchData {
		var save ctrl.SaveRequest
		if len(item.Payload) > 0 {
			if err := json.Unmarshal(item.Payload, &save); err != nil {
				klog.Errorf("batch item %s: invalid payload: %v", item.ID, err)
				failed = append(failed, item.ID)
				continue
			}
		}
		if save.MacAddress == "" {
			save.MacAddress = defaultMac
		}
		if save.Timestamp == 0 {
			save.Timestamp = item.Timestamp
		}
		if save.MacAddress == "" {
			klog.Errorf("batch item %s: mac address missing", item.ID)
			failed = append(failed, item.ID)
			continue
		}

		if _, err := r.c.Recording().Save(c.Request.Context(), save); err != nil {
			klog.Errorf("batch item %s: save failed: %v", item.ID, err)
			failed = append(failed, item.ID)
			continue
		}
		processed++
	}

	c.JSON(http.StatusOK, batchUploadResponse{
		Success:        len(failed) == 0,
		Message:        "ok",
		ProcessedCount: processed,
		FailedItems:    failed,
		Timestamp:      time.Now().UnixMilli(),
	})
}
