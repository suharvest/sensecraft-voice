package db

import (
	"context"
	"time"

	"gorm.io/gorm"
	"k8s.io/klog/v2"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
)

type SystemPromptInterface interface {
	Create(ctx context.Context, sp *model.SystemPrompt) error
	Update(ctx context.Context, sp *model.SystemPrompt) error
	UpdateWithActive(ctx context.Context, sp *model.SystemPrompt, updateActive bool) error
	UpdateStatus(ctx context.Context, id int64, isActive bool) error
	SetAsDefault(ctx context.Context, id int64) error
	Delete(ctx context.Context, id int64) error
	BatchDelete(ctx context.Context, ids []int64) error
	Get(ctx context.Context, id int64) (*model.SystemPrompt, error)
	GetByName(ctx context.Context, name string) (*model.SystemPrompt, error)
	SearchByName(ctx context.Context, name string, limit int) ([]*model.SystemPrompt, error)
	List(ctx context.Context, filter *SystemPromptFilter) ([]*model.SystemPrompt, int64, error)
	GetActiveDefault(ctx context.Context) (*model.SystemPrompt, error)
}

type SystemPromptFilter struct {
	Name   string
	Role   string
	Active *bool
	Offset int
	Limit  int
}

type systemPromptDao struct{ db *gorm.DB }

func newSystemPrompt(db *gorm.DB) SystemPromptInterface { return &systemPromptDao{db: db} }

func (d *systemPromptDao) Create(ctx context.Context, sp *model.SystemPrompt) error {
	now := time.Now().UnixMilli()
	sp.CreatedAt = now
	sp.UpdatedAt = now
	klog.Infof("Create: 准备创建系统提示词，ID=%d, Name=%s, IsActive=%v, IsDefault=%v", sp.ID, sp.Name, sp.IsActive, sp.IsDefault)

	// 如果设置为默认，需要先清除其他默认记录
	if sp.IsDefault {
		if err := d.db.WithContext(ctx).Model(&model.SystemPrompt{}).Where("is_default = ?", true).Updates(map[string]interface{}{
			"is_default": false,
			"updated_at": now,
		}).Error; err != nil {
			klog.Errorf("Create: 清除其他默认记录失败: %v", err)
			return err
		}
	}

	err := d.db.WithContext(ctx).Create(sp).Error
	if err != nil {
		klog.Errorf("Create: 创建失败: %v", err)
	} else {
		klog.Infof("Create: 创建成功，ID=%d, IsActive=%v, IsDefault=%v", sp.ID, sp.IsActive, sp.IsDefault)
	}
	return err
}

func (d *systemPromptDao) Update(ctx context.Context, sp *model.SystemPrompt) error {
	now := time.Now().UnixMilli()
	sp.UpdatedAt = now

	// 如果设置为默认，需要先清除其他默认记录
	if sp.IsDefault {
		if err := d.db.WithContext(ctx).Model(&model.SystemPrompt{}).Where("is_default = ? AND id != ?", true, sp.ID).Updates(map[string]interface{}{
			"is_default": false,
			"updated_at": now,
		}).Error; err != nil {
			klog.Errorf("Update: 清除其他默认记录失败: %v", err)
			return err
		}
	}

	return d.db.WithContext(ctx).Model(&model.SystemPrompt{}).Where("id = ?", sp.ID).Updates(sp).Error
}

func (d *systemPromptDao) UpdateWithActive(ctx context.Context, sp *model.SystemPrompt, updateActive bool) error {
	now := time.Now().UnixMilli()
	updates := map[string]interface{}{
		"updated_at": now,
	}

	if sp.Name != "" {
		updates["name"] = sp.Name
	}
	if sp.Role != "" {
		updates["role"] = sp.Role
	}
	if sp.Content != "" {
		updates["content"] = sp.Content
	}
	if sp.Tags != "" {
		updates["tags"] = sp.Tags
	}
	if updateActive {
		updates["is_active"] = sp.IsActive
		updates["is_default"] = sp.IsDefault
		klog.Infof("UpdateWithActive: id = %d, updateActive = %v, isActive = %v, isDefault = %v, updates = %+v", sp.ID, updateActive, sp.IsActive, sp.IsDefault, updates)

		// 如果设置为默认，需要先清除其他默认记录
		if sp.IsDefault {
			if err := d.db.WithContext(ctx).Model(&model.SystemPrompt{}).Where("is_default = ? AND id != ?", true, sp.ID).Updates(map[string]interface{}{
				"is_default": false,
				"updated_at": now,
			}).Error; err != nil {
				klog.Errorf("UpdateWithActive: 清除其他默认记录失败: %v", err)
				return err
			}
		}
	}

	return d.db.WithContext(ctx).Model(&model.SystemPrompt{}).Where("id = ?", sp.ID).Updates(updates).Error
}

func (d *systemPromptDao) UpdateStatus(ctx context.Context, id int64, isActive bool) error {
	now := time.Now().UnixMilli()
	klog.Infof("UpdateStatus: id = %d, isActive = %v", id, isActive)
	return d.db.WithContext(ctx).Model(&model.SystemPrompt{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_active":  isActive,
		"updated_at": now,
	}).Error
}

func (d *systemPromptDao) SetAsDefault(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()

	// 开始事务
	tx := d.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 先将所有其他记录的 is_default 设为 false
	if err := tx.Model(&model.SystemPrompt{}).Where("id != ?", id).Updates(map[string]interface{}{
		"is_default": false,
		"updated_at": now,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 2. 将指定记录的 is_default 设为 true
	if err := tx.Model(&model.SystemPrompt{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_default": true,
		"updated_at": now,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	klog.Infof("SetAsDefault: 成功设置 ID=%d 为默认系统提示词", id)
	return nil
}

func (d *systemPromptDao) Delete(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).Delete(&model.SystemPrompt{}, id).Error
}

func (d *systemPromptDao) BatchDelete(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	return d.db.WithContext(ctx).Where("id IN ?", ids).Delete(&model.SystemPrompt{}).Error
}

func (d *systemPromptDao) Get(ctx context.Context, id int64) (*model.SystemPrompt, error) {
	var sp model.SystemPrompt
	if err := d.db.WithContext(ctx).First(&sp, id).Error; err != nil {
		return nil, err
	}
	return &sp, nil
}

func (d *systemPromptDao) GetByName(ctx context.Context, name string) (*model.SystemPrompt, error) {
	var sp model.SystemPrompt
	if err := d.db.WithContext(ctx).Where("name = ?", name).First(&sp).Error; err != nil {
		return nil, err
	}
	return &sp, nil
}

func (d *systemPromptDao) SearchByName(ctx context.Context, name string, limit int) ([]*model.SystemPrompt, error) {
	var sps []*model.SystemPrompt
	q := d.db.WithContext(ctx).Model(&model.SystemPrompt{}).Where("name LIKE ?", "%"+name+"%")

	if limit > 0 {
		q = q.Limit(limit)
	}

	if err := q.Order("id DESC").Find(&sps).Error; err != nil {
		return nil, err
	}
	return sps, nil
}

func (d *systemPromptDao) List(ctx context.Context, filter *SystemPromptFilter) ([]*model.SystemPrompt, int64, error) {
	var sps []*model.SystemPrompt
	q := d.db.WithContext(ctx).Model(&model.SystemPrompt{})
	if filter != nil {
		if filter.Name != "" {
			q = q.Where("name LIKE ?", "%"+filter.Name+"%")
		}
		if filter.Role != "" {
			q = q.Where("role = ?", filter.Role)
		}
		if filter.Active != nil {
			q = q.Where("is_active = ?", *filter.Active)
		}
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if filter != nil {
		if filter.Offset > 0 {
			q = q.Offset(filter.Offset)
		}
		if filter.Limit > 0 {
			q = q.Limit(filter.Limit)
		}
	}
	if err := q.Order("id DESC").Find(&sps).Error; err != nil {
		return nil, 0, err
	}
	return sps, total, nil
}

func (d *systemPromptDao) GetActiveDefault(ctx context.Context) (*model.SystemPrompt, error) {
	var sp model.SystemPrompt
	if err := d.db.WithContext(ctx).Where("is_active = ? AND is_default = ?", true, true).First(&sp).Error; err != nil {
		return nil, err
	}
	return &sp, nil
}
