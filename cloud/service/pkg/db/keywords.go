package db

import (
	"context"
	"time"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/types"
	"gorm.io/gorm"
)

type KeywordInterface interface {
	Create(ctx context.Context, keyword *model.Keyword) error
	GetById(ctx context.Context, id int64) (*model.Keyword, error)
	GetByKeyword(ctx context.Context, keyword string) (*model.Keyword, error)
	List(ctx context.Context, req *types.ListKeywordsRequest) ([]*model.Keyword, int64, error)
	Update(ctx context.Context, keyword *model.Keyword) error
	Delete(ctx context.Context, id int64) error
	BatchDelete(ctx context.Context, ids []int64) (int64, []int64, error)
}

type keyword struct {
	db *gorm.DB
}

func newKeyword(db *gorm.DB) KeywordInterface {
	return &keyword{db: db}
}

func (k *keyword) Create(ctx context.Context, keyword *model.Keyword) error {
	now := time.Now().UnixMilli()
	keyword.CreatedAt = now
	keyword.UpdatedAt = now
	return k.db.WithContext(ctx).Create(keyword).Error
}

func (k *keyword) GetById(ctx context.Context, id int64) (*model.Keyword, error) {
	var keyword model.Keyword
	err := k.db.WithContext(ctx).Where("id = ?", id).First(&keyword).Error
	if err != nil {
		return nil, err
	}
	return &keyword, nil
}

func (k *keyword) GetByKeyword(ctx context.Context, keyword string) (*model.Keyword, error) {
	var kw model.Keyword
	err := k.db.WithContext(ctx).Where("keyword = ?", keyword).First(&kw).Error
	if err != nil {
		return nil, err
	}
	return &kw, nil
}

func (k *keyword) List(ctx context.Context, req *types.ListKeywordsRequest) ([]*model.Keyword, int64, error) {
	var keywords []*model.Keyword
	var total int64

	query := k.db.WithContext(ctx).Model(&model.Keyword{})

	// 关键词搜索
	if req.Keyword != "" {
		query = query.Where("keyword LIKE ?", "%"+req.Keyword+"%")
	}

	// 颜色筛选
	if req.MarkColor != "" {
		query = query.Where("mark_color = ?", req.MarkColor)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if req.Limit == 0 {
		req.Limit = 10
	}

	err := query.Offset(req.Offset).Limit(req.Limit).Order("id desc").Find(&keywords).Error
	return keywords, total, err
}

func (k *keyword) Update(ctx context.Context, keyword *model.Keyword) error {
	keyword.UpdatedAt = time.Now().UnixMilli()
	err := k.db.WithContext(ctx).Model(keyword).Updates(map[string]interface{}{
		"keyword":    keyword.Keyword,
		"synonyms":   keyword.Synonyms,
		"mark_color": keyword.MarkColor,
		"updated_at": keyword.UpdatedAt,
	}).Error

	return err
}

func (k *keyword) Delete(ctx context.Context, id int64) error {
	return k.db.WithContext(ctx).Delete(&model.Keyword{}, id).Error
}

func (k *keyword) BatchDelete(ctx context.Context, ids []int64) (int64, []int64, error) {
	var deletedIDs []int64
	var deletedCount int64

	// 先查询存在的ID
	var existingKeywords []*model.Keyword
	err := k.db.WithContext(ctx).Where("id IN ?", ids).Find(&existingKeywords).Error
	if err != nil {
		return 0, nil, err
	}

	// 提取存在的ID
	for _, kw := range existingKeywords {
		deletedIDs = append(deletedIDs, kw.ID)
	}

	if len(deletedIDs) == 0 {
		return 0, []int64{}, nil
	}

	// 批量删除
	result := k.db.WithContext(ctx).Where("id IN ?", deletedIDs).Delete(&model.Keyword{})
	if result.Error != nil {
		return 0, nil, result.Error
	}

	deletedCount = result.RowsAffected
	return deletedCount, deletedIDs, nil
}
