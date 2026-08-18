package db

import (
	"context"
	"time"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
	"gorm.io/gorm"
)

type UserInterface interface {
	Create(ctx context.Context, user *model.User) error
	GetById(ctx context.Context, id int64) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	List(ctx context.Context, offset, limit int) ([]*model.User, error)
	ListWithFilter(ctx context.Context, username string, offset, limit int) ([]*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id int64) error
	Count(ctx context.Context) (int64, error)
	CountWithFilter(ctx context.Context, username string) (int64, error)
}

type user struct {
	db *gorm.DB
}

func newUser(db *gorm.DB) UserInterface {
	return &user{db: db}
}

func (u *user) Create(ctx context.Context, user *model.User) error {
	now := time.Now().UnixMilli()
	user.CreatedAt = now
	user.UpdatedAt = now
	return u.db.WithContext(ctx).Create(user).Error
}

func (u *user) GetById(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	err := u.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *user) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := u.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *user) List(ctx context.Context, offset, limit int) ([]*model.User, error) {
	var users []*model.User
	err := u.db.WithContext(ctx).Offset(offset).Limit(limit).Order("id desc").Find(&users).Error
	return users, err
}

func (u *user) Update(ctx context.Context, user *model.User) error {
	user.UpdatedAt = time.Now().UnixMilli()
	err := u.db.WithContext(ctx).Model(user).Updates(map[string]interface{}{
		"username":   user.Username,
		"updated_at": user.UpdatedAt,
	}).Error

	return err
}

func (u *user) Delete(ctx context.Context, id int64) error {
	return u.db.WithContext(ctx).Delete(&model.User{}, id).Error
}

func (u *user) Count(ctx context.Context) (int64, error) {
	var count int64
	err := u.db.WithContext(ctx).Model(&model.User{}).Count(&count).Error
	return count, err
}

func (u *user) ListWithFilter(ctx context.Context, username string, offset, limit int) ([]*model.User, error) {
	var users []*model.User
	query := u.db.WithContext(ctx).Model(&model.User{})

	// 添加用户名模糊搜索条件
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}

	err := query.Offset(offset).Limit(limit).Order("id desc").Find(&users).Error
	return users, err
}

func (u *user) CountWithFilter(ctx context.Context, username string) (int64, error) {
	var count int64
	query := u.db.WithContext(ctx).Model(&model.User{})

	// 添加用户名模糊搜索条件
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}

	err := query.Count(&count).Error
	return count, err
}
