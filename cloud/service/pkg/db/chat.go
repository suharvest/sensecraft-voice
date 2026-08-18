package db

import (
	"context"
	"time"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/types"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// ChatInterface 聊天数据访问接口
type ChatInterface interface {
	// SaveChatSession 保存聊天会话
	SaveChatSession(ctx context.Context, session *types.ChatSession) error
	// GetChatSession 获取聊天会话
	GetChatSession(ctx context.Context, sessionID string) (*types.ChatSession, error)
	// GetChatSessions 获取聊天会话列表
	GetChatSessions(ctx context.Context, userID string) ([]*types.ChatSession, error)
	// DeleteChatSession 删除聊天会话（同时删除相关消息）
	DeleteChatSession(ctx context.Context, sessionID string) error
	// DeleteChatSessions 批量删除聊天会话（同时删除相关消息）
	DeleteChatSessions(ctx context.Context, sessionIDs []string) error
	// SaveChatMessage 保存聊天消息
	SaveChatMessage(ctx context.Context, message *types.ChatMessage) error
	// GetChatHistory 获取聊天历史
	GetChatHistory(ctx context.Context, sessionID string, limit int) ([]*types.ChatMessage, error)
	// SaveChatStats 保存聊天统计
	SaveChatStats(ctx context.Context, stats *types.ChatStats) error
	// BatchSaveMessages 批量保存聊天消息
	BatchSaveMessages(ctx context.Context, messages []*types.ChatMessage) error
	// SaveChatWithMetadata 保存聊天会话、消息和统计信息（事务）
	SaveChatWithMetadata(ctx context.Context, session *types.ChatSession, messages []*types.ChatMessage, stats *types.ChatStats) error
	// GetChatHistoryForContext 获取聊天历史（用于上下文管理，按时间正序）
	GetChatHistoryForContext(ctx context.Context, sessionID string, limit int) ([]*types.ChatMessage, error)
	// UpdateChatSessionTitle 更新聊天会话标题
	UpdateChatSessionTitle(ctx context.Context, sessionID, title string) error
}

// chatImpl 聊天数据访问实现
type chatImpl struct {
	db *gorm.DB
}

// NewChat 创建聊天DAO实例
func NewChat(db *gorm.DB) ChatInterface {
	return &chatImpl{db: db}
}

// SaveChatSession 保存聊天会话
func (c *chatImpl) SaveChatSession(ctx context.Context, session *types.ChatSession) error {
	if session.CreatedAt == 0 {
		session.CreatedAt = time.Now().UnixMilli()
	}
	session.UpdatedAt = time.Now().UnixMilli()

	return c.db.WithContext(ctx).Save(session).Error
}

// GetChatSession 获取聊天会话
func (c *chatImpl) GetChatSession(ctx context.Context, sessionID string) (*types.ChatSession, error) {
	var session types.ChatSession
	err := c.db.WithContext(ctx).Where("session_id = ?", sessionID).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// SaveChatMessage 保存聊天消息
func (c *chatImpl) SaveChatMessage(ctx context.Context, message *types.ChatMessage) error {
	if message.CreatedAt == 0 {
		message.CreatedAt = time.Now().UnixMilli()
	}

	klog.Infof("数据库层保存聊天消息: SessionID=%s, MessageID=%s, Event=%s, Content长度=%d",
		message.SessionID, message.MessageID, message.Event, len(message.Content))

	err := c.db.WithContext(ctx).Create(message).Error
	if err != nil {
		klog.Errorf("数据库层保存聊天消息失败: %v", err)
	} else {
		klog.Infof("数据库层保存聊天消息成功: SessionID=%s, MessageID=%s",
			message.SessionID, message.MessageID)
	}

	return err
}

// GetChatHistory 获取聊天历史
func (c *chatImpl) GetChatHistory(ctx context.Context, sessionID string, limit int) ([]*types.ChatMessage, error) {
	var messages []*types.ChatMessage
	err := c.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at asc"). // 改为正序，保持对话顺序
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

// GetChatSessions 获取聊天会话列表
func (c *chatImpl) GetChatSessions(ctx context.Context, userID string) ([]*types.ChatSession, error) {
	var sessions []*types.ChatSession
	query := c.db.WithContext(ctx).Model(&types.ChatSession{})

	// 如果指定了用户ID，则按用户过滤
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	// 按创建时间倒序排列
	query = query.Order("created_at DESC")

	err := query.Find(&sessions).Error
	return sessions, err
}

// DeleteChatSession 删除聊天会话（同时删除相关消息）
func (c *chatImpl) DeleteChatSession(ctx context.Context, sessionID string) error {
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除会话相关的消息
		if err := tx.Where("session_id = ?", sessionID).Delete(&types.ChatMessage{}).Error; err != nil {
			return err
		}

		// 删除会话
		if err := tx.Where("session_id = ?", sessionID).Delete(&types.ChatSession{}).Error; err != nil {
			return err
		}

		return nil
	})
}

// DeleteChatSessions 批量删除聊天会话（同时删除相关消息）
func (c *chatImpl) DeleteChatSessions(ctx context.Context, sessionIDs []string) error {
	if len(sessionIDs) == 0 {
		return nil
	}

	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除会话相关的消息
		if err := tx.Where("session_id IN ?", sessionIDs).Delete(&types.ChatMessage{}).Error; err != nil {
			return err
		}

		// 删除会话
		if err := tx.Where("session_id IN ?", sessionIDs).Delete(&types.ChatSession{}).Error; err != nil {
			return err
		}

		return nil
	})
}

// SaveChatStats 保存聊天统计
func (c *chatImpl) SaveChatStats(ctx context.Context, stats *types.ChatStats) error {
	if stats.CreatedAt == 0 {
		stats.CreatedAt = time.Now().UnixMilli()
	}

	return c.db.WithContext(ctx).Create(stats).Error
}

// BatchSaveMessages 批量保存聊天消息
func (c *chatImpl) BatchSaveMessages(ctx context.Context, messages []*types.ChatMessage) error {
	if len(messages) == 0 {
		return nil
	}

	// 设置创建时间
	now := time.Now().UnixMilli()
	for _, msg := range messages {
		if msg.CreatedAt == 0 {
			msg.CreatedAt = now
		}
	}

	return c.db.WithContext(ctx).CreateInBatches(messages, 100).Error
}

// SaveChatWithMetadata 保存聊天会话、消息和统计信息（事务）
func (c *chatImpl) SaveChatWithMetadata(ctx context.Context, session *types.ChatSession, messages []*types.ChatMessage, stats *types.ChatStats) error {
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 存储会话
		if session != nil {
			if session.CreatedAt == 0 {
				session.CreatedAt = time.Now().UnixMilli()
			}
			session.UpdatedAt = time.Now().UnixMilli()
			if err := tx.Save(session).Error; err != nil {
				return err
			}
		}

		// 批量存储消息
		if len(messages) > 0 {
			now := time.Now().UnixMilli()
			for _, msg := range messages {
				if msg.CreatedAt == 0 {
					msg.CreatedAt = now
				}
			}
			if err := tx.CreateInBatches(messages, 100).Error; err != nil {
				return err
			}
		}

		// 存储统计
		if stats != nil {
			if stats.CreatedAt == 0 {
				stats.CreatedAt = time.Now().UnixMilli()
			}
			if err := tx.Create(stats).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetChatHistoryForContext 获取聊天历史（用于上下文管理，按时间正序）
func (c *chatImpl) GetChatHistoryForContext(ctx context.Context, sessionID string, limit int) ([]*types.ChatMessage, error) {
	var messages []*types.ChatMessage
	err := c.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at asc"). // 正序排列，保持对话顺序
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

// UpdateChatSessionTitle 更新聊天会话标题
func (c *chatImpl) UpdateChatSessionTitle(ctx context.Context, sessionID, title string) error {
	return c.db.WithContext(ctx).
		Model(&types.ChatSession{}).
		Where("session_id = ?", sessionID).
		Update("title", title).Error
}
