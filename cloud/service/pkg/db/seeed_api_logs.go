package db

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
)

type SeeedAPILogInterface interface {
	Create(ctx context.Context, object *model.SeeedAPILog) (*model.SeeedAPILog, error)
	GetByQueryLogID(ctx context.Context, queryLogID int64) ([]*model.SeeedAPILog, error)
	GetBySessionID(ctx context.Context, sessionID string) ([]*model.SeeedAPILog, error)
}

type seeedAPILog struct {
	db *gorm.DB
}

func newSeeedAPILog(db *gorm.DB) SeeedAPILogInterface { return &seeedAPILog{db: db} }

func (s *seeedAPILog) Create(ctx context.Context, object *model.SeeedAPILog) (*model.SeeedAPILog, error) {
	object.CreatedAt = time.Now().UnixMilli()

	if err := s.db.WithContext(ctx).Create(object).Error; err != nil {
		return nil, err
	}
	return object, nil
}

func (s *seeedAPILog) GetByQueryLogID(ctx context.Context, queryLogID int64) ([]*model.SeeedAPILog, error) {
	var logs []*model.SeeedAPILog
	err := s.db.WithContext(ctx).
		Where("query_log_id = ?", queryLogID).
		Order("created_at asc").
		Find(&logs).Error
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func (s *seeedAPILog) GetBySessionID(ctx context.Context, sessionID string) ([]*model.SeeedAPILog, error) {
	var logs []*model.SeeedAPILog
	err := s.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at asc").
		Find(&logs).Error
	if err != nil {
		return nil, err
	}
	return logs, nil
}
