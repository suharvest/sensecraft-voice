package db

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
)

type AsrServerInterface interface {
	Create(ctx context.Context, object *model.AsrServer) (*model.AsrServer, error)
	Update(ctx context.Context, id int64, updates map[string]interface{}) error
	Delete(ctx context.Context, id int64) error
	GetById(ctx context.Context, id int64) (*model.AsrServer, error)
	List(ctx context.Context, offset, limit int) ([]*model.AsrServer, error)
	Count(ctx context.Context) (int64, error)
	ListAll(ctx context.Context) ([]*model.AsrServer, error)
	CountDevices(ctx context.Context, id int64) (int64, error)
}

type asrServer struct {
	db *gorm.DB
}

func newAsrServer(db *gorm.DB) AsrServerInterface { return &asrServer{db: db} }

func (a *asrServer) Create(ctx context.Context, object *model.AsrServer) (*model.AsrServer, error) {
	now := time.Now().UnixMilli()
	object.CreatedAt = now
	object.UpdatedAt = now
	if err := a.db.WithContext(ctx).Create(object).Error; err != nil {
		return nil, err
	}
	return object, nil
}

func (a *asrServer) Update(ctx context.Context, id int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	updates["updated_at"] = time.Now().UnixMilli()
	return a.db.WithContext(ctx).Model(&model.AsrServer{}).Where("id = ?", id).Updates(updates).Error
}

func (a *asrServer) Delete(ctx context.Context, id int64) error {
	return a.db.WithContext(ctx).Where("id = ?", id).Delete(&model.AsrServer{}).Error
}

func (a *asrServer) GetById(ctx context.Context, id int64) (*model.AsrServer, error) {
	var obj model.AsrServer
	if err := a.db.WithContext(ctx).Where("id = ?", id).First(&obj).Error; err != nil {
		return nil, err
	}
	return &obj, nil
}

func (a *asrServer) List(ctx context.Context, offset, limit int) ([]*model.AsrServer, error) {
	var items []*model.AsrServer
	err := a.db.WithContext(ctx).Offset(offset).Limit(limit).Order("id desc").Find(&items).Error
	return items, err
}

func (a *asrServer) ListAll(ctx context.Context) ([]*model.AsrServer, error) {
	var items []*model.AsrServer
	err := a.db.WithContext(ctx).Order("id asc").Find(&items).Error
	return items, err
}

func (a *asrServer) Count(ctx context.Context) (int64, error) {
	var count int64
	err := a.db.WithContext(ctx).Model(&model.AsrServer{}).Count(&count).Error
	return count, err
}

func (a *asrServer) CountDevices(ctx context.Context, id int64) (int64, error) {
	var count int64
	err := a.db.WithContext(ctx).Model(&model.Device{}).Where("asr_server_id = ?", id).Count(&count).Error
	return count, err
}
