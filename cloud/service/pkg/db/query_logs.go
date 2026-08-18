package db

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
)

type QueryLogInterface interface {
	Create(ctx context.Context, object *model.QueryLog) (*model.QueryLog, error)
	UpdateProcessingResult(ctx context.Context, id int64, openAIAnswer string, seeedAPIStatus int8, processingError string) error
	GetByID(ctx context.Context, id int64) (*model.QueryLog, error)
	GetBySessionID(ctx context.Context, sessionID string) (*model.QueryLog, error)
}

type queryLog struct {
	db *gorm.DB
}

func newQueryLog(db *gorm.DB) QueryLogInterface { return &queryLog{db: db} }

func (q *queryLog) Create(ctx context.Context, object *model.QueryLog) (*model.QueryLog, error) {
	now := time.Now().UnixMilli()
	object.CreatedAt = now
	object.UpdatedAt = now

	if err := q.db.WithContext(ctx).Create(object).Error; err != nil {
		return nil, err
	}
	return object, nil
}

func (q *queryLog) UpdateProcessingResult(ctx context.Context, id int64, openAIAnswer string, seeedAPIStatus int8, processingError string) error {
	updates := map[string]interface{}{
		"openai_answer":     openAIAnswer,
		"seeed_api_status":  seeedAPIStatus,
		"processing_error":  processingError,
		"updated_at":        time.Now().UnixMilli(),
	}

	return q.db.WithContext(ctx).
		Model(&model.QueryLog{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (q *queryLog) GetByID(ctx context.Context, id int64) (*model.QueryLog, error) {
	var log model.QueryLog
	err := q.db.WithContext(ctx).Where("id = ?", id).First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (q *queryLog) GetBySessionID(ctx context.Context, sessionID string) (*model.QueryLog, error) {
	var log model.QueryLog
	err := q.db.WithContext(ctx).Where("session_id = ?", sessionID).First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}
