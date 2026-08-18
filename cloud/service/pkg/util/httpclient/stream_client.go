package httpclient

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// StreamClient 流式HTTP客户端，支持Server-Sent Events (SSE)
type StreamClient struct {
	client  *http.Client
	config  *StreamConfig
	headers map[string]string
}

// StreamConfig 流式客户端配置
type StreamConfig struct {
	Timeout     time.Duration
	EnableDebug bool
	BaseURL     string
	Headers     map[string]string
}

// StreamEvent SSE事件结构
type StreamEvent struct {
	Event string                 `json:"event"`
	Data  map[string]interface{} `json:"data,omitempty"`
	ID    string                 `json:"id,omitempty"`
	Retry int                    `json:"retry,omitempty"`
	Raw   string                 `json:"raw,omitempty"`
}

// StreamEventHandler 流事件处理器
type StreamEventHandler func(event *StreamEvent) error

// DefaultStreamConfig 默认流式配置
func DefaultStreamConfig() *StreamConfig {
	return &StreamConfig{
		Timeout:     60 * time.Second,
		EnableDebug: false,
		Headers:     make(map[string]string),
	}
}

// NewStreamClient 创建新的流式客户端
func NewStreamClient(config *StreamConfig) *StreamClient {
	if config == nil {
		config = DefaultStreamConfig()
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	return &StreamClient{
		client:  client,
		config:  config,
		headers: make(map[string]string),
	}
}

// SetAuthToken 设置认证令牌
func (c *StreamClient) SetAuthToken(token string) *StreamClient {
	c.headers["Authorization"] = "Bearer " + token
	return c
}

// SetHeader 设置请求头
func (c *StreamClient) SetHeader(key, value string) *StreamClient {
	c.headers[key] = value
	return c
}

// SetHeaders 批量设置请求头
func (c *StreamClient) SetHeaders(headers map[string]string) *StreamClient {
	for k, v := range headers {
		c.headers[k] = v
	}
	return c
}

// PostStream 发送POST请求并处理流式响应
func (c *StreamClient) PostStream(ctx context.Context, url string, body interface{}, handler StreamEventHandler) error {
	// 构建完整URL
	fullURL := url
	if c.config.BaseURL != "" && !strings.HasPrefix(url, "http") {
		fullURL = strings.TrimRight(c.config.BaseURL, "/") + "/" + strings.TrimLeft(url, "/")
	}

	// 序列化请求体
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body failed: %w", err)
		}
		bodyReader = strings.NewReader(string(bodyBytes))
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", fullURL, bodyReader)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	// 设置默认头部
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	// 设置自定义头部
	for k, v := range c.config.Headers {
		req.Header.Set(k, v)
	}
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	if c.config.EnableDebug {
		fmt.Printf("[StreamClient] Request: %s %s\n", req.Method, req.URL.String())
		for k, v := range req.Header {
			fmt.Printf("[StreamClient] Header: %s: %s\n", k, v)
		}
	}

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 处理流式响应
	return c.handleStream(resp.Body, handler)
}

// GetStream 发送GET请求并处理流式响应
func (c *StreamClient) GetStream(ctx context.Context, url string, handler StreamEventHandler) error {
	// 构建完整URL
	fullURL := url
	if c.config.BaseURL != "" && !strings.HasPrefix(url, "http") {
		fullURL = strings.TrimRight(c.config.BaseURL, "/") + "/" + strings.TrimLeft(url, "/")
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	// 设置默认头部
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	// 设置自定义头部
	for k, v := range c.config.Headers {
		req.Header.Set(k, v)
	}
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 处理流式响应
	return c.handleStream(resp.Body, handler)
}

// handleStream 处理流式响应
func (c *StreamClient) handleStream(reader io.Reader, handler StreamEventHandler) error {
	scanner := bufio.NewScanner(reader)
	var event *StreamEvent

	for scanner.Scan() {
		line := scanner.Text()

		if c.config.EnableDebug {
			fmt.Printf("[StreamClient] Received: %s\n", line)
		}

		// 空行表示事件结束
		if line == "" {
			if event != nil && event.Data != nil {
				if err := handler(event); err != nil {
					return fmt.Errorf("event handler failed: %w", err)
				}
			}
			event = nil
			continue
		}

		// 解析SSE格式
		if strings.HasPrefix(line, "data: ") {
			dataStr := strings.TrimPrefix(line, "data: ")

			// 初始化事件对象
			if event == nil {
				event = &StreamEvent{
					Data: make(map[string]interface{}),
				}
			}

			// 尝试解析JSON数据
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
				// 如果不是JSON，保存原始字符串
				event.Raw = dataStr
			} else {
				// 提取事件类型
				if eventType, ok := data["event"].(string); ok {
					event.Event = eventType
				}
				event.Data = data
			}
		} else if strings.HasPrefix(line, "event: ") {
			if event == nil {
				event = &StreamEvent{
					Data: make(map[string]interface{}),
				}
			}
			event.Event = strings.TrimPrefix(line, "event: ")
		} else if strings.HasPrefix(line, "id: ") {
			if event == nil {
				event = &StreamEvent{
					Data: make(map[string]interface{}),
				}
			}
			event.ID = strings.TrimPrefix(line, "id: ")
		} else if strings.HasPrefix(line, "retry: ") {
			if event == nil {
				event = &StreamEvent{
					Data: make(map[string]interface{}),
				}
			}
			// 解析重试时间（这里简化处理）
			event.Retry = 1000
		}
	}

	// 处理最后一个事件
	if event != nil && event.Data != nil {
		if err := handler(event); err != nil {
			return fmt.Errorf("final event handler failed: %w", err)
		}
	}

	return scanner.Err()
}

// ChatRequest Dify API 聊天请求结构
type ChatRequest struct {
	Inputs         map[string]interface{} `json:"inputs"`
	Query          string                 `json:"query"`
	ResponseMode   string                 `json:"response_mode"`
	ConversationID string                 `json:"conversation_id"`
	User           string                 `json:"user"`
	Files          []ChatFile             `json:"files,omitempty"`
}

// ChatFile 聊天文件结构
type ChatFile struct {
	Type           string `json:"type"`
	TransferMethod string `json:"transfer_method"`
	URL            string `json:"url"`
}

// PostChatStream 发送聊天请求并处理流式响应（专门用于Dify API）
func (c *StreamClient) PostChatStream(ctx context.Context, url string, req *ChatRequest, handler StreamEventHandler) error {
	return c.PostStream(ctx, url, req, handler)
}
