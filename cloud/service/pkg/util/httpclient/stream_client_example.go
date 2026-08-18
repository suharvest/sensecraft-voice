package httpclient

import (
	"context"
	"fmt"
	"log"
	"time"
)

// ExampleUsage 展示如何使用StreamClient
func ExampleUsage() {
	// 创建流式客户端
	config := &StreamConfig{
		Timeout:     60 * time.Second,
		EnableDebug: true,
		BaseURL:     "https://dify.example.com",
	}

	client := NewStreamClient(config).
		SetAuthToken("your_api_key_here").
		SetHeader("Content-Type", "application/json")

	// 构建聊天请求
	chatReq := &ChatRequest{
		Inputs:         map[string]interface{}{},
		Query:          "What are the specs of the iPhone 13 Pro Max?",
		ResponseMode:   "streaming",
		ConversationID: "",
		User:           "abc-123",
		Files: []ChatFile{
			{
				Type:           "image",
				TransferMethod: "remote_url",
				URL:            "https://cloud.dify.ai/logo/logo-site.png",
			},
		},
	}

	// 定义事件处理器
	handler := func(event *StreamEvent) error {
		switch event.Event {
		case "workflow_started":
			fmt.Printf("工作流开始: %v\n", event.Data)
		case "node_started":
			fmt.Printf("节点开始: %s\n", event.Data["title"])
		case "node_finished":
			fmt.Printf("节点完成: %s (耗时: %v秒)\n",
				event.Data["title"], event.Data["elapsed_time"])
		case "workflow_finished":
			fmt.Printf("工作流完成: 状态=%s, 总耗时=%v秒\n",
				event.Data["status"], event.Data["elapsed_time"])
		case "message":
			// 流式消息片段
			fmt.Printf("%s", event.Data["answer"])
		case "message_end":
			fmt.Printf("\n消息完成: 总token=%v, 总价格=%s %s\n",
				event.Data["metadata"].(map[string]interface{})["usage"].(map[string]interface{})["total_tokens"],
				event.Data["metadata"].(map[string]interface{})["usage"].(map[string]interface{})["total_price"],
				event.Data["metadata"].(map[string]interface{})["usage"].(map[string]interface{})["currency"])
		case "tts_message":
			fmt.Printf("TTS音频片段接收\n")
		case "tts_message_end":
			fmt.Printf("TTS音频完成\n")
		default:
			fmt.Printf("未知事件: %s\n", event.Event)
		}
		return nil
	}

	// 发送流式聊天请求
	ctx := context.Background()
	if err := client.PostChatStream(ctx, "/v1/chat-messages", chatReq, handler); err != nil {
		log.Fatalf("流式请求失败: %v", err)
	}
}

// ExampleDifyAPIUsage 具体的Dify API使用示例
func ExampleDifyAPIUsage() {
	client := NewStreamClient(&StreamConfig{
		BaseURL:     "https://dify.example.com",
		Timeout:     60 * time.Second,
		EnableDebug: false,
	}).SetAuthToken("your_actual_api_key")

	// 简单的文本对话
	req := &ChatRequest{
		Inputs:       map[string]interface{}{},
		Query:        "Hello, how are you?",
		ResponseMode: "streaming",
		User:         "user-123",
	}

	var fullMessage string
	handler := func(event *StreamEvent) error {
		if event.Event == "message" {
			if answer, ok := event.Data["answer"].(string); ok {
				fullMessage += answer
				fmt.Print(answer) // 实时输出
			}
		} else if event.Event == "message_end" {
			fmt.Printf("\n[完整消息]: %s\n", fullMessage)
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.PostChatStream(ctx, "/v1/chat-messages", req, handler); err != nil {
		log.Printf("请求失败: %v", err)
	}
}
