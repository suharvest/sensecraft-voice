package voice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	logutil "github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/util/log"
)

// HTTPBatchClient HTTP批量上传客户端
type HTTPBatchClient struct {
	baseURL    string
	httpClient *http.Client
	headers    map[string]string
	timeout    time.Duration
}

// BatchUploadRequest 批量上传请求
type BatchUploadRequest struct {
	DeviceInfo DeviceInfo     `json:"device_info"`
	BatchData  []ASRCacheData `json:"batch_data"`
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	MacAddress    string `json:"mac_address"`
	ClientVersion string `json:"client_version"`
	Timestamp     int64  `json:"timestamp"`
}

// BatchUploadResponse 批量上传响应
type BatchUploadResponse struct {
	Success        bool     `json:"success"`
	Message        string   `json:"message"`
	ProcessedCount int      `json:"processed_count"`
	FailedItems    []string `json:"failed_items"`
	Timestamp      int64    `json:"timestamp"`
}

// NewHTTPBatchClient 创建新的HTTP批量上传客户端
func NewHTTPBatchClient(baseURL string, headers map[string]string, timeout time.Duration) *HTTPBatchClient {
	return &HTTPBatchClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		headers: headers,
		timeout: timeout,
	}
}

// UploadBatch 批量上传ASR缓存数据
func (c *HTTPBatchClient) UploadBatch(cacheFiles []ASRCacheData, macAddress string) (*BatchUploadResponse, error) {
	if len(cacheFiles) == 0 {
		return &BatchUploadResponse{
			Success:        true,
			Message:        "No data to upload",
			ProcessedCount: 0,
			FailedItems:    []string{},
			Timestamp:      time.Now().UnixMilli(),
		}, nil
	}

	// 构造请求
	request := BatchUploadRequest{
		DeviceInfo: DeviceInfo{
			MacAddress:    macAddress,
			ClientVersion: "1.0.0",
			Timestamp:     time.Now().UnixMilli(),
		},
		BatchData: cacheFiles,
	}

	// 序列化请求
	requestData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch upload request: %w", err)
	}

	// 构造HTTP请求
	req, err := http.NewRequestWithContext(context.Background(), "POST", c.baseURL+"/api/v1/recordings/batch", bytes.NewBuffer(requestData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "SenseCraft-Voice-Client/1.0")
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	logutil.Infof("HTTP batch client: uploading %d cache files to %s", len(cacheFiles), c.baseURL)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var response BatchUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, response.Message)
	}

	logutil.Infof("HTTP batch client: upload completed, success=%t, processed=%d, failed=%d",
		response.Success, response.ProcessedCount, len(response.FailedItems))

	return &response, nil
}

// UploadBatchWithRetry 带重试的批量上传
func (c *HTTPBatchClient) UploadBatchWithRetry(cacheFiles []ASRCacheData, macAddress string, maxRetries int) (*BatchUploadResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// 指数退避
			backoff := time.Duration(attempt) * time.Second
			logutil.Infof("HTTP batch client: retry attempt %d after %v", attempt, backoff)
			time.Sleep(backoff)
		}

		response, err := c.UploadBatch(cacheFiles, macAddress)
		if err == nil {
			return response, nil
		}

		lastErr = err
		logutil.Warnf("HTTP batch client: upload attempt %d failed: %v", attempt+1, err)
	}

	return nil, fmt.Errorf("upload failed after %d attempts: %w", maxRetries+1, lastErr)
}

// TestConnection 测试连接
func (c *HTTPBatchClient) TestConnection() error {
	// 发送一个空的批量请求来测试连接
	emptyResponse, err := c.UploadBatch([]ASRCacheData{}, "")
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	if !emptyResponse.Success {
		return fmt.Errorf("connection test failed: %s", emptyResponse.Message)
	}

	logutil.Infof("HTTP batch client: connection test successful")
	return nil
}

// GetBaseURL 获取基础URL
func (c *HTTPBatchClient) GetBaseURL() string {
	return c.baseURL
}

// GetTimeout 获取超时时间
func (c *HTTPBatchClient) GetTimeout() time.Duration {
	return c.timeout
}

// SetTimeout 设置超时时间
func (c *HTTPBatchClient) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
	c.httpClient.Timeout = timeout
}

// SetHeaders 设置请求头
func (c *HTTPBatchClient) SetHeaders(headers map[string]string) {
	c.headers = headers
}

// AddHeader 添加请求头
func (c *HTTPBatchClient) AddHeader(key, value string) {
	if c.headers == nil {
		c.headers = make(map[string]string)
	}
	c.headers[key] = value
}

// RemoveHeader 移除请求头
func (c *HTTPBatchClient) RemoveHeader(key string) {
	if c.headers != nil {
		delete(c.headers, key)
	}
}

// GetHeaders 获取请求头
func (c *HTTPBatchClient) GetHeaders() map[string]string {
	return c.headers
}
