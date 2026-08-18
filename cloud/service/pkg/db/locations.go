package db

import (
	"context"
	"time"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
	"gorm.io/gorm"
)

type LocationInterface interface {
	Create(ctx context.Context, location *model.Location) error
	GetById(ctx context.Context, id int64) (*model.Location, error)
	ListByStoreId(ctx context.Context, storeId int64, offset, limit int) ([]*model.Location, error)
	List(ctx context.Context, offset, limit int) ([]*model.Location, error)
	ListByName(ctx context.Context, name string, offset, limit int) ([]*model.Location, error)
	ListByStoreIdAndName(ctx context.Context, storeId int64, name string, offset, limit int) ([]*model.Location, error)
	Update(ctx context.Context, location *model.Location) error
	Delete(ctx context.Context, id int64) error
	Count(ctx context.Context) (int64, error)
	CountByStoreId(ctx context.Context, storeId int64) (int64, error)
	CountByName(ctx context.Context, name string) (int64, error)
	CountByStoreIdAndName(ctx context.Context, storeId int64, name string) (int64, error)
}

type location struct {
	db *gorm.DB
}

func newLocation(db *gorm.DB) LocationInterface {
	return &location{db: db}
}

func (l *location) Create(ctx context.Context, location *model.Location) error {
	now := time.Now().UnixMilli()
	location.CreatedAt = now
	location.UpdatedAt = now
	return l.db.WithContext(ctx).Create(location).Error
}

func (l *location) GetById(ctx context.Context, id int64) (*model.Location, error) {
	var location model.Location
	err := l.db.WithContext(ctx).Where("id = ?", id).First(&location).Error
	if err != nil {
		return nil, err
	}
	return &location, nil
}

func (l *location) ListByStoreId(ctx context.Context, storeId int64, offset, limit int) ([]*model.Location, error) {
	var locations []*model.Location
	err := l.db.WithContext(ctx).Where("store_id = ?", storeId).Offset(offset).Limit(limit).Order("id desc").Find(&locations).Error
	return locations, err
}

func (l *location) List(ctx context.Context, offset, limit int) ([]*model.Location, error) {
	var locations []*model.Location
	err := l.db.WithContext(ctx).Offset(offset).Limit(limit).Order("id desc").Find(&locations).Error
	return locations, err
}

func (l *location) Update(ctx context.Context, location *model.Location) error {
	location.UpdatedAt = time.Now().UnixMilli()
	return l.db.WithContext(ctx).Save(location).Error
}

func (l *location) Delete(ctx context.Context, id int64) error {
	return l.db.WithContext(ctx).Delete(&model.Location{}, id).Error
}

func (l *location) Count(ctx context.Context) (int64, error) {
	var count int64
	err := l.db.WithContext(ctx).Model(&model.Location{}).Count(&count).Error
	return count, err
}

func (l *location) CountByStoreId(ctx context.Context, storeId int64) (int64, error) {
	var count int64
	err := l.db.WithContext(ctx).Model(&model.Location{}).Where("store_id = ?", storeId).Count(&count).Error
	return count, err
}

func (l *location) ListByName(ctx context.Context, name string, offset, limit int) ([]*model.Location, error) {
	var locations []*model.Location
	err := l.db.WithContext(ctx).
		Where("name LIKE ?", "%"+name+"%").
		Offset(offset).
		Limit(limit).
		Order("id desc").
		Find(&locations).Error
	return locations, err
}

func (l *location) CountByName(ctx context.Context, name string) (int64, error) {
	var count int64
	err := l.db.WithContext(ctx).
		Model(&model.Location{}).
		Where("name LIKE ?", "%"+name+"%").
		Count(&count).Error
	return count, err
}

func (l *location) ListByStoreIdAndName(ctx context.Context, storeId int64, name string, offset, limit int) ([]*model.Location, error) {
	var locations []*model.Location
	err := l.db.WithContext(ctx).
		Where("store_id = ? AND name LIKE ?", storeId, "%"+name+"%").
		Offset(offset).
		Limit(limit).
		Order("id desc").
		Find(&locations).Error
	return locations, err
}

func (l *location) CountByStoreIdAndName(ctx context.Context, storeId int64, name string) (int64, error) {
	var count int64
	err := l.db.WithContext(ctx).
		Model(&model.Location{}).
		Where("store_id = ? AND name LIKE ?", storeId, "%"+name+"%").
		Count(&count).Error
	return count, err
}
