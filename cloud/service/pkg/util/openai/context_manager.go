package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/types"
	"k8s.io/klog/v2"
)

// 基于现有chat表的上下文管理器
type ChatContextManager struct {
	factory db.ShareDaoFactory
}

// 创建上下文管理器
func NewContextManager(factory db.ShareDaoFactory) ContextManager {
	return &ChatContextManager{
		factory: factory,
	}
}

// 保存消息到数据库
func (cm *ChatContextManager) SaveMessage(sessionID, role, content string, metadata map[string]interface{}) error {
	ctx := context.Background()

	message := &types.ChatMessage{
		SessionID: sessionID,
		MessageID: generateMessageID(),
		Event:     role, // 直接使用role作为event类型
		Content:   content,
		Data:      serializeMetadata(metadata),
		CreatedAt: time.Now().UnixMilli(),
	}

	if err := cm.factory.Chat().SaveChatMessage(ctx, message); err != nil {
		return fmt.Errorf("保存消息失败: %w", err)
	}

	klog.Infof("保存消息成功: session=%s, role=%s, content_len=%d", sessionID, role, len(content))
	return nil
}

// 获取会话上下文
func (cm *ChatContextManager) GetConversationContext(sessionID string, maxMessages int) ([]Message, error) {
	ctx := context.Background()

	// 从数据库获取消息历史（按时间正序）
	dbMessages, err := cm.factory.Chat().GetChatHistoryForContext(ctx, sessionID, maxMessages)
	if err != nil {
		return nil, fmt.Errorf("获取聊天历史失败: %w", err)
	}

	// 转换为OpenAI格式并分类
	var systemMessages []Message
	var otherMessages []Message

	for _, msg := range dbMessages {
		// 处理所有类型的消息，包括系统消息，但过滤掉空内容
		if (msg.Event == "user" || msg.Event == "assistant" || msg.Event == "system") && len(msg.Content) > 0 {
			message := Message{
				Role:    msg.Event,
				Content: msg.Content,
			}

			if msg.Event == "system" {
				systemMessages = append(systemMessages, message)
				klog.Infof("包含系统消息到上下文: role=%s, content_len=%d", msg.Event, len(msg.Content))
			} else {
				otherMessages = append(otherMessages, message)
				klog.Infof("包含消息到上下文: role=%s, content_len=%d", msg.Event, len(msg.Content))
			}
		} else {
			klog.Infof("跳过消息: role=%s, content_len=%d", msg.Event, len(msg.Content))
		}
	}

	// 重新组合消息：系统消息在前，其他消息按时间顺序在后
	var messages []Message
	messages = append(messages, systemMessages...)
	messages = append(messages, otherMessages...)

	klog.Infof("获取会话上下文: session=%s, messages_count=%d (system=%d, other=%d)", sessionID, len(messages), len(systemMessages), len(otherMessages))
	return messages, nil
}

// 创建新会话
func (cm *ChatContextManager) CreateSession(userID string) (string, error) {
	ctx := context.Background()

	sessionID := generateSessionID()
	session := &types.ChatSession{
		SessionID: sessionID,
		UserID:    userID,
		Status:    "active",
		CreatedAt: time.Now().UnixMilli(),
		UpdatedAt: time.Now().UnixMilli(),
	}

	if err := cm.factory.Chat().SaveChatSession(ctx, session); err != nil {
		return "", fmt.Errorf("创建会话失败: %w", err)
	}

	klog.Infof("创建新会话: session=%s, user=%s", sessionID, userID)
	return sessionID, nil
}

// 关闭会话
func (cm *ChatContextManager) CloseSession(sessionID string) error {
	ctx := context.Background()

	// 获取会话
	session, err := cm.factory.Chat().GetChatSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("获取会话失败: %w", err)
	}

	// 更新会话状态
	session.Status = "closed"
	session.UpdatedAt = time.Now().UnixMilli()

	// 这里需要添加更新会话的方法，暂时跳过
	klog.Infof("关闭会话: session=%s", sessionID)
	return nil
}

// 获取会话历史（用于显示）
func (cm *ChatContextManager) GetConversationHistory(sessionID string, limit int) ([]*types.ChatMessage, error) {
	ctx := context.Background()
	return cm.factory.Chat().GetChatHistory(ctx, sessionID, limit)
}

// 清理旧消息（保持上下文在合理范围内）
func (cm *ChatContextManager) CleanupOldMessages(sessionID string, maxMessages int) error {
	ctx := context.Background()

	// 获取所有消息
	allMessages, err := cm.factory.Chat().GetChatHistory(ctx, sessionID, 1000)
	if err != nil {
		return fmt.Errorf("获取消息历史失败: %w", err)
	}

	// 如果消息数量超过限制，删除最旧的消息
	if len(allMessages) > maxMessages {
		messagesToDelete := allMessages[:len(allMessages)-maxMessages]
		for _, msg := range messagesToDelete {
			// 这里需要添加删除消息的方法，暂时跳过
			klog.Infof("需要删除旧消息: %s", msg.MessageID)
		}
	}

	return nil
}

// 辅助函数
func generateSessionID() string {
	return fmt.Sprintf("openai_%d", time.Now().UnixMilli())
}

func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixMilli())
}

func serializeMetadata(metadata map[string]interface{}) string {
	if metadata == nil {
		return ""
	}

	bytes, err := json.Marshal(metadata)
	if err != nil {
		klog.Warningf("序列化元数据失败: %v", err)
		return ""
	}

	return string(bytes)
}
