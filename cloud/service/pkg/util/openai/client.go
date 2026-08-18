package openai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"k8s.io/klog/v2"
)

// OpenAI客户端实现
type OpenAIClient struct {
	config         *Config
	httpClient     *http.Client
	contextManager ContextManager
}

// GetContextManager 获取上下文管理器
func (c *OpenAIClient) GetContextManager() ContextManager {
	return c.contextManager
}

// 创建新的OpenAI客户端
func NewClient(config *Config, contextManager ContextManager) Client {
	return &OpenAIClient{
		config: config,
		httpClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
		contextManager: contextManager,
	}
}

// 创建聊天补全
func (c *OpenAIClient) CreateChatCompletion(req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	// 设置默认值
	if req.Model == "" {
		req.Model = c.config.Model
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = c.config.MaxTokens
	}
	if req.Temperature == 0 {
		req.Temperature = c.config.Temperature
	}

	// 序列化请求
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequest("POST", c.config.BaseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	// 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.Unmarshal(respBody, &errorResp); err != nil {
			return nil, fmt.Errorf("API请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
		}
		return nil, fmt.Errorf("API请求失败: %s", errorResp.Error.Message)
	}

	// 解析响应
	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	klog.Infof("OpenAI API调用成功，使用tokens: %d", chatResp.Usage.TotalTokens)
	return &chatResp, nil
}

// 创建流式聊天补全
func (c *OpenAIClient) CreateChatCompletionStream(req *ChatCompletionRequest) (<-chan *StreamEvent, error) {
	// 设置流式模式
	req.Stream = true

	// 序列化请求
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequest("POST", c.config.BaseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Cache-Control", "no-cache")

	// 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	// 创建事件通道
	eventChan := make(chan *StreamEvent, 100)

	// 在goroutine中处理流式响应
	go func() {
		defer resp.Body.Close()
		defer close(eventChan)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			// 跳过空行和注释
			if line == "" || strings.HasPrefix(line, ":") {
				continue
			}

			// 处理SSE格式：data: {...}
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")

				// 检查是否是结束标记
				if data == "[DONE]" {
					klog.Infof("流式响应结束")
					break
				}

				// 解析JSON数据
				var event StreamEvent
				if err := json.Unmarshal([]byte(data), &event); err != nil {
					klog.Errorf("解析流式事件失败: %v, data: %s", err, data)
					continue
				}

				// 发送事件到通道
				select {
				case eventChan <- &event:
				default:
					klog.Warningf("事件通道已满，丢弃事件")
				}
			}
		}

		if err := scanner.Err(); err != nil {
			klog.Errorf("读取流式响应失败: %v", err)
		}
	}()

	return eventChan, nil
}

// 带上下文的聊天
func (c *OpenAIClient) ChatWithContext(sessionID, userMessage, userID string) (*ChatCompletionResponse, error) {
	// 1. 保存用户消息
	if err := c.contextManager.SaveMessage(sessionID, "user", userMessage, map[string]interface{}{
		"user_id":   userID,
		"timestamp": time.Now().UnixMilli(),
	}); err != nil {
		return nil, fmt.Errorf("保存用户消息失败: %w", err)
	}

	// 2. 获取会话上下文
	contextMessages, err := c.contextManager.GetConversationContext(sessionID, 20)
	if err != nil {
		return nil, fmt.Errorf("获取会话上下文失败: %w", err)
	}

	// 3. 创建OpenAI请求
	req := &ChatCompletionRequest{
		Model:       c.config.Model,
		Messages:    contextMessages,
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
	}

	// 4. 调用OpenAI API
	resp, err := c.CreateChatCompletion(req)
	if err != nil {
		return nil, err
	}

	// 5. 保存AI回复
	if len(resp.Choices) > 0 {
		assistantMessage := resp.Choices[0].Message.Content
		if err := c.contextManager.SaveMessage(sessionID, "assistant", assistantMessage, map[string]interface{}{
			"user_id":       userID,
			"timestamp":     time.Now().UnixMilli(),
			"usage":         resp.Usage,
			"finish_reason": resp.Choices[0].FinishReason,
		}); err != nil {
			klog.Warningf("保存AI回复失败: %v", err)
		}
	}

	return resp, nil
}

// 流式聊天
func (c *OpenAIClient) ChatWithContextStream(sessionID, userMessage, userID string) (<-chan *StreamEvent, error) {
	// 1. 保存用户消息
	if err := c.contextManager.SaveMessage(sessionID, "user", userMessage, map[string]interface{}{
		"user_id":   userID,
		"timestamp": time.Now().UnixMilli(),
	}); err != nil {
		return nil, fmt.Errorf("保存用户消息失败: %w", err)
	}

	// 2. 获取会话上下文
	contextMessages, err := c.contextManager.GetConversationContext(sessionID, 20)
	if err != nil {
		return nil, fmt.Errorf("获取会话上下文失败: %w", err)
	}

	// 3. 创建OpenAI请求
	req := &ChatCompletionRequest{
		Model:       c.config.Model,
		Messages:    contextMessages,
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
		Stream:      true,
	}

	// 调试日志：打印发送给OpenAI的消息
	klog.Infof("发送给OpenAI的消息数量: %d", len(contextMessages))
	for i, msg := range contextMessages {
		klog.Infof("消息 %d: role=%s, content_len=%d, content_preview='%s'", i, msg.Role, len(msg.Content), truncateString(msg.Content, 100))
	}

	// 4. 调用流式API
	eventChan, err := c.CreateChatCompletionStream(req)
	if err != nil {
		return nil, err
	}

	// 5. 创建包装的事件通道，用于保存完整的AI回复
	wrappedChan := make(chan *StreamEvent, 100)
	var fullResponse strings.Builder

	go func() {
		defer close(wrappedChan)
		for event := range eventChan {
			// 转发事件
			select {
			case wrappedChan <- event:
			default:
				klog.Warningf("包装事件通道已满，丢弃事件")
			}

			// 累积完整回复
			if len(event.Choices) > 0 && event.Choices[0].Message.Content != "" {
				fullResponse.WriteString(event.Choices[0].Message.Content)
			}
		}

		// 保存完整的AI回复
		if fullResponse.Len() > 0 {
			if err := c.contextManager.SaveMessage(sessionID, "assistant", fullResponse.String(), map[string]interface{}{
				"user_id":         userID,
				"timestamp":       time.Now().UnixMilli(),
				"stream_complete": true,
			}); err != nil {
				klog.Warningf("保存完整AI回复失败: %v", err)
			}
		}
	}()

	return wrappedChan, nil
}

// 截断字符串用于日志显示
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
