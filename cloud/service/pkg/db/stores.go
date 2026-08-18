package db

import (
	"context"
	"time"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
	"gorm.io/gorm"
)

type StoreInterface interface {
	Create(ctx context.Context, store *model.Store) error
	GetById(ctx context.Context, id int64) (*model.Store, error)
	GetByCode(ctx context.Context, code string) (*model.Store, error)
	List(ctx context.Context, offset, limit int) ([]*model.Store, error)
	ListByName(ctx context.Context, name string, offset, limit int) ([]*model.Store, error)
	Update(ctx context.Context, store *model.Store) error
	Delete(ctx context.Context, id int64) error
	Count(ctx context.Context) (int64, error)
	CountByName(ctx context.Context, name string) (int64, error)
}

type store struct {
	db *gorm.DB
}

func newStore(db *gorm.DB) StoreInterface {
	return &store{db: db}
}

func (s *store) Create(ctx context.Context, store *model.Store) error {
	now := time.Now().UnixMilli()
	store.CreatedAt = now
	store.UpdatedAt = now
	return s.db.WithContext(ctx).Create(store).Error
}

func (s *store) GetById(ctx context.Context, id int64) (*model.Store, error) {
	var store model.Store
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&store).Error
	if err != nil {
		return nil, err
	}
	return &store, nil
}

func (s *store) GetByCode(ctx context.Context, code string) (*model.Store, error) {
	var store model.Store
	err := s.db.WithContext(ctx).Where("code = ?", code).First(&store).Error
	if err != nil {
		return nil, err
	}
	return &store, nil
}

func (s *store) List(ctx context.Context, offset, limit int) ([]*model.Store, error) {
	var stores []*model.Store
	err := s.db.WithContext(ctx).Offset(offset).Limit(limit).Order("id desc").Find(&stores).Error
	return stores, err
}

func (s *store) Update(ctx context.Context, store *model.Store) error {
	store.UpdatedAt = time.Now().UnixMilli()
	return s.db.WithContext(ctx).Save(store).Error
}

func (s *store) Delete(ctx context.Context, id int64) error {
	return s.db.WithContext(ctx).Delete(&model.Store{}, id).Error
}

func (s *store) Count(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&model.Store{}).Count(&count).Error
	return count, err
}

func (s *store) ListByName(ctx context.Context, name string, offset, limit int) ([]*model.Store, error) {
	var stores []*model.Store
	err := s.db.WithContext(ctx).
		Where("name LIKE ?", "%"+name+"%").
		Offset(offset).
		Limit(limit).
		Order("id desc").
		Find(&stores).Error
	return stores, err
}

func (s *store) CountByName(ctx context.Context, name string) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).
		Model(&model.Store{}).
		Where("name LIKE ?", "%"+name+"%").
		Count(&count).Error
	return count, err
}
