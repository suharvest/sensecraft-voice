package keywords

import (
	"context"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/errors"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/config"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/types"
	"k8s.io/klog/v2"
)

// Controller 关键词控制器实现
type Controller struct {
	cc      config.Config
	factory db.ShareDaoFactory
}

// NewController 创建关键词控制器
func NewController(cfg config.Config, f db.ShareDaoFactory) Interface {
	return &Controller{cc: cfg, factory: f}
}

// Create 创建关键词
func (c *Controller) Create(ctx context.Context, req *types.CreateKeywordRequest) (*model.Keyword, error) {
	// 检查关键词是否已存在
	if _, err := c.factory.Keyword().GetByKeyword(ctx, req.Keyword); err == nil {
		klog.Errorf("Keyword %s already exists", req.Keyword)
		return nil, errors.ErrInvalidRequest
	}

	// 创建新关键词
	keyword := &model.Keyword{
		Keyword:   req.Keyword,
		Synonyms:  req.Synonyms,
		MarkColor: req.MarkColor,
	}

	if err := c.factory.Keyword().Create(ctx, keyword); err != nil {
		klog.Errorf("Failed to create keyword: %v", err)
		return nil, errors.ErrServerInternal
	}

	return keyword, nil
}

// GetById 根据ID获取关键词
func (c *Controller) GetById(ctx context.Context, id int64) (*model.Keyword, error) {
	keyword, err := c.factory.Keyword().GetById(ctx, id)
	if err != nil {
		klog.Errorf("Failed to get keyword by id %d: %v", id, err)
		return nil, errors.ErrServerInternal
	}
	return keyword, nil
}

// List 获取关键词列表
func (c *Controller) List(ctx context.Context, req *types.ListKeywordsRequest) (*types.ListKeywordsResponse, error) {
	if req.Limit == 0 {
		req.Limit = 10
	}

	keywords, total, err := c.factory.Keyword().List(ctx, req)
	if err != nil {
		klog.Errorf("Failed to list keywords: %v", err)
		return nil, errors.ErrServerInternal
	}

	return &types.ListKeywordsResponse{
		Total:  total,
		Items:  keywords,
		Limit:  req.Limit,
		Offset: req.Offset,
	}, nil
}

// Update 更新关键词
func (c *Controller) Update(ctx context.Context, id int64, req *types.UpdateKeywordRequest) (*model.Keyword, error) {
	// 获取现有关键词
	keyword, err := c.factory.Keyword().GetById(ctx, id)
	if err != nil {
		klog.Errorf("Failed to get keyword by id %d: %v", id, err)
		return nil, errors.ErrServerInternal
	}

	// 如果更新了关键词名称，检查是否与其他关键词重复
	if req.Keyword != "" && req.Keyword != keyword.Keyword {
		if _, err := c.factory.Keyword().GetByKeyword(ctx, req.Keyword); err == nil {
			klog.Errorf("Keyword %s already exists", req.Keyword)
			return nil, errors.ErrInvalidRequest
		}
		keyword.Keyword = req.Keyword
	}

	// 更新其他字段
	if req.Synonyms != "" {
		keyword.Synonyms = req.Synonyms
	}
	if req.MarkColor != "" {
		keyword.MarkColor = req.MarkColor
	}

	if err := c.factory.Keyword().Update(ctx, keyword); err != nil {
		klog.Errorf("Failed to update keyword: %v", err)
		return nil, errors.ErrServerInternal
	}

	return keyword, nil
}

// Delete 删除关键词
func (c *Controller) Delete(ctx context.Context, id int64) error {
	if err := c.factory.Keyword().Delete(ctx, id); err != nil {
		klog.Errorf("Failed to delete keyword: %v", err)
		return errors.ErrServerInternal
	}
	return nil
}

// BatchDelete 批量删除关键词
func (c *Controller) BatchDelete(ctx context.Context, req *types.BatchDeleteKeywordsRequest) (*types.BatchDeleteKeywordsResponse, error) {
	deletedCount, deletedIDs, err := c.factory.Keyword().BatchDelete(ctx, req.IDs)
	if err != nil {
		klog.Errorf("Failed to batch delete keywords: %v", err)
		return nil, errors.ErrServerInternal
	}

	return &types.BatchDeleteKeywordsResponse{
		DeletedCount: deletedCount,
		DeletedIDs:   deletedIDs,
	}, nil
}
