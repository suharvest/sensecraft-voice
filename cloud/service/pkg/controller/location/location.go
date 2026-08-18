package location

import (
	"context"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/errors"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/config"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
	"k8s.io/klog/v2"
)

type LocationGetter interface {
	Location() Interface
}

type Interface interface {
	Create(ctx context.Context, in CreateRequest) (*model.Location, error)
	GetById(ctx context.Context, id int64) (*model.Location, error)
	List(ctx context.Context, in ListRequest) (*ListResponse, error)
	ListByStoreId(ctx context.Context, storeId int64, in ListRequest) (*ListResponse, error)
	Update(ctx context.Context, id int64, in UpdateRequest) (*model.Location, error)
	Delete(ctx context.Context, id int64) error
}

type CreateRequest struct {
	StoreId     int64  `json:"store_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	Status      int8   `json:"status"`
}

type UpdateRequest struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Status      int8   `json:"status"`
}

type ListRequest struct {
	Offset  int    `form:"offset" binding:"min=0"`
	Limit   int    `form:"limit" binding:"min=1,max=100"`
	Name    string `form:"name"`     // 点位名称，支持模糊搜索
	StoreId int64  `form:"store_id"` // 门店ID，支持按门店过滤
}

type ListResponse struct {
	Total int64             `json:"total"`
	Items []*model.Location `json:"items"`
}

type location struct {
	cc      config.Config
	factory db.ShareDaoFactory
}

func (l *location) Create(ctx context.Context, in CreateRequest) (*model.Location, error) {
	// 检查门店是否存在
	if _, err := l.factory.Store().GetById(ctx, in.StoreId); err != nil {
		klog.Errorf("Store %d not found", in.StoreId)
		return nil, errors.ErrInvalidRequest
	}

	obj := &model.Location{
		StoreId:     in.StoreId,
		Name:        in.Name,
		Code:        in.Code,
		Description: in.Description,
		Status:      in.Status,
	}

	if err := l.factory.Location().Create(ctx, obj); err != nil {
		klog.Errorf("Failed to create location: %v", err)
		return nil, errors.ErrServerInternal
	}
	return obj, nil
}

func (l *location) GetById(ctx context.Context, id int64) (*model.Location, error) {
	obj, err := l.factory.Location().GetById(ctx, id)
	if err != nil {
		klog.Errorf("Failed to get location by id %d: %v", id, err)
		return nil, errors.ErrServerInternal
	}
	return obj, nil
}

func (l *location) List(ctx context.Context, in ListRequest) (*ListResponse, error) {
	if in.Limit == 0 {
		in.Limit = 20
	}

	var total int64
	var items []*model.Location
	var err error

	// 根据参数组合决定使用哪个查询方法
	if in.StoreId > 0 && in.Name != "" {
		// 同时有store_id和name参数
		total, err = l.factory.Location().CountByStoreIdAndName(ctx, in.StoreId, in.Name)
		if err != nil {
			klog.Errorf("Failed to count locations by store_id and name: %v", err)
			return nil, errors.ErrServerInternal
		}

		items, err = l.factory.Location().ListByStoreIdAndName(ctx, in.StoreId, in.Name, in.Offset, in.Limit)
		if err != nil {
			klog.Errorf("Failed to list locations by store_id and name: %v", err)
			return nil, errors.ErrServerInternal
		}
	} else if in.StoreId > 0 {
		// 只有store_id参数
		total, err = l.factory.Location().CountByStoreId(ctx, in.StoreId)
		if err != nil {
			klog.Errorf("Failed to count locations by store_id: %v", err)
			return nil, errors.ErrServerInternal
		}

		items, err = l.factory.Location().ListByStoreId(ctx, in.StoreId, in.Offset, in.Limit)
		if err != nil {
			klog.Errorf("Failed to list locations by store_id: %v", err)
			return nil, errors.ErrServerInternal
		}
	} else if in.Name != "" {
		// 只有name参数
		total, err = l.factory.Location().CountByName(ctx, in.Name)
		if err != nil {
			klog.Errorf("Failed to count locations by name: %v", err)
			return nil, errors.ErrServerInternal
		}

		items, err = l.factory.Location().ListByName(ctx, in.Name, in.Offset, in.Limit)
		if err != nil {
			klog.Errorf("Failed to list locations by name: %v", err)
			return nil, errors.ErrServerInternal
		}
	} else {
		// 没有任何过滤参数，查询所有
		total, err = l.factory.Location().Count(ctx)
		if err != nil {
			klog.Errorf("Failed to count locations: %v", err)
			return nil, errors.ErrServerInternal
		}

		items, err = l.factory.Location().List(ctx, in.Offset, in.Limit)
		if err != nil {
			klog.Errorf("Failed to list locations: %v", err)
			return nil, errors.ErrServerInternal
		}
	}

	return &ListResponse{
		Total: total,
		Items: items,
	}, nil
}

func (l *location) ListByStoreId(ctx context.Context, storeId int64, in ListRequest) (*ListResponse, error) {
	if in.Limit == 0 {
		in.Limit = 20
	}

	total, err := l.factory.Location().CountByStoreId(ctx, storeId)
	if err != nil {
		klog.Errorf("Failed to count locations for store %d: %v", storeId, err)
		return nil, errors.ErrServerInternal
	}

	items, err := l.factory.Location().ListByStoreId(ctx, storeId, in.Offset, in.Limit)
	if err != nil {
		klog.Errorf("Failed to list locations for store %d: %v", storeId, err)
		return nil, errors.ErrServerInternal
	}

	return &ListResponse{
		Total: total,
		Items: items,
	}, nil
}

func (l *location) Update(ctx context.Context, id int64, in UpdateRequest) (*model.Location, error) {
	obj, err := l.factory.Location().GetById(ctx, id)
	if err != nil {
		klog.Errorf("Failed to get location by id %d: %v", id, err)
		return nil, errors.ErrServerInternal
	}

	if in.Name != "" {
		obj.Name = in.Name
	}
	if in.Code != "" {
		obj.Code = in.Code
	}
	if in.Description != "" {
		obj.Description = in.Description
	}
	if in.Status >= 0 {
		obj.Status = in.Status
	}

	if err := l.factory.Location().Update(ctx, obj); err != nil {
		klog.Errorf("Failed to update location: %v", err)
		return nil, errors.ErrServerInternal
	}
	return obj, nil
}

func (l *location) Delete(ctx context.Context, id int64) error {
	// 检查是否有设备关联
	// TODO: 实现设备关联检查
	if err := l.factory.Location().Delete(ctx, id); err != nil {
		klog.Errorf("Failed to delete location: %v", err)
		return errors.ErrServerInternal
	}
	return nil
}

func NewLocation(cfg config.Config, f db.ShareDaoFactory) *location {
	return &location{cc: cfg, factory: f}
}
