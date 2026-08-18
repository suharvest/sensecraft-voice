package httpclient

import (
	"context"
	"testing"
	"time"
)

func TestStreamClient_Basic(t *testing.T) {
	// 创建客户端
	client := NewStreamClient(&StreamConfig{
		Timeout:     10 * time.Second,
		EnableDebug: true,
	})

	if client == nil {
		t.Fatal("创建StreamClient失败")
	}

	// 测试设置头部
	client.SetAuthToken("test-token").
		SetHeader("Custom-Header", "test-value")

	if client.headers["Authorization"] != "Bearer test-token" {
		t.Errorf("设置认证令牌失败，期望: Bearer test-token, 实际: %s", client.headers["Authorization"])
	}

	if client.headers["Custom-Header"] != "test-value" {
		t.Errorf("设置自定义头部失败，期望: test-value, 实际: %s", client.headers["Custom-Header"])
	}
}

func TestChatRequest_Structure(t *testing.T) {
	// 测试聊天请求结构
	req := &ChatRequest{
		Inputs:         map[string]interface{}{"key": "value"},
		Query:          "测试查询",
		ResponseMode:   "streaming",
		ConversationID: "test-conversation",
		User:           "test-user",
		Files: []ChatFile{
			{
				Type:           "image",
				TransferMethod: "remote_url",
				URL:            "https://example.com/image.png",
			},
		},
	}

	if req.Query != "测试查询" {
		t.Errorf("查询字段设置错误，期望: 测试查询, 实际: %s", req.Query)
	}

	if len(req.Files) != 1 {
		t.Errorf("文件数量错误，期望: 1, 实际: %d", len(req.Files))
	}

	if req.Files[0].Type != "image" {
		t.Errorf("文件类型错误，期望: image, 实际: %s", req.Files[0].Type)
	}
}

func TestStreamEvent_Structure(t *testing.T) {
	// 测试流事件结构
	event := &StreamEvent{
		Event: "message",
		Data: map[string]interface{}{
			"answer":    "测试回答",
			"timestamp": 1679586595,
		},
		ID: "test-id",
	}

	if event.Event != "message" {
		t.Errorf("事件类型错误，期望: message, 实际: %s", event.Event)
	}

	if event.Data["answer"] != "测试回答" {
		t.Errorf("事件数据错误，期望: 测试回答, 实际: %v", event.Data["answer"])
	}
}

// 注意：这个测试需要实际的API端点，在CI/CD中应该跳过
func TestStreamClient_RealAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过实际API测试")
	}

	client := NewStreamClient(&StreamConfig{
		BaseURL:     "https://dify.example.com",
		Timeout:     30 * time.Second,
		EnableDebug: true,
	}).SetAuthToken("test-api-key")

	req := &ChatRequest{
		Inputs:       map[string]interface{}{},
		Query:        "Hello",
		ResponseMode: "streaming",
		User:         "test-user",
	}

	eventCount := 0
	handler := func(event *StreamEvent) error {
		eventCount++
		t.Logf("接收到事件 #%d: %s", eventCount, event.Event)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 这个测试可能会因为网络问题或API密钥问题而失败，这是正常的
	err := client.PostChatStream(ctx, "/v1/chat-messages", req, handler)
	if err != nil {
		t.Logf("API调用失败（这是预期的，因为使用了测试密钥）: %v", err)
	}
}
