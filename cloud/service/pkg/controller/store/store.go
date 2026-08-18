package store

import (
	"context"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/errors"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/config"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
	"k8s.io/klog/v2"
)

type StoreGetter interface {
	Store() Interface
}

type Interface interface {
	Create(ctx context.Context, in CreateRequest) (*model.Store, error)
	GetById(ctx context.Context, id int64) (*model.Store, error)
	List(ctx context.Context, in ListRequest) (*ListResponse, error)
	Update(ctx context.Context, id int64, in UpdateRequest) (*model.Store, error)
	Delete(ctx context.Context, id int64) error
}

type CreateRequest struct {
	Name    string `json:"name" binding:"required"`
	Code    string `json:"code" binding:"required"`
	Address string `json:"address"`
	Contact string `json:"contact"`
	Phone   string `json:"phone"`
	Status  int8   `json:"status"`
}

type UpdateRequest struct {
	Name    string `json:"name"`
	Code    string `json:"code"`
	Address string `json:"address"`
	Contact string `json:"contact"`
	Phone   string `json:"phone"`
	Status  int8   `json:"status"`
}

type ListRequest struct {
	Offset int    `form:"offset" binding:"min=0"`
	Limit  int    `form:"limit" binding:"min=1,max=100"`
	Name   string `form:"name"` // 门店名称，支持模糊搜索
}

type ListResponse struct {
	Total int64          `json:"total"`
	Items []*model.Store `json:"items"`
}

type store struct {
	cc      config.Config
	factory db.ShareDaoFactory
}

func (s *store) Create(ctx context.Context, in CreateRequest) (*model.Store, error) {
	// 检查门店代码是否已存在
	if _, err := s.factory.Store().GetByCode(ctx, in.Code); err == nil {
		klog.Errorf("Store code %s already exists", in.Code)
		return nil, errors.ErrInvalidRequest
	}

	obj := &model.Store{
		Name:    in.Name,
		Code:    in.Code,
		Address: in.Address,
		Contact: in.Contact,
		Phone:   in.Phone,
		Status:  in.Status,
	}

	if err := s.factory.Store().Create(ctx, obj); err != nil {
		klog.Errorf("Failed to create store: %v", err)
		return nil, errors.ErrServerInternal
	}
	return obj, nil
}

func (s *store) GetById(ctx context.Context, id int64) (*model.Store, error) {
	obj, err := s.factory.Store().GetById(ctx, id)
	if err != nil {
		klog.Errorf("Failed to get store by id %d: %v", id, err)
		return nil, errors.ErrServerInternal
	}
	return obj, nil
}

func (s *store) List(ctx context.Context, in ListRequest) (*ListResponse, error) {
	if in.Limit == 0 {
		in.Limit = 20
	}

	// 根据是否有name参数决定使用哪个查询方法
	var total int64
	var items []*model.Store
	var err error

	if in.Name != "" {
		// 使用带name过滤的查询
		total, err = s.factory.Store().CountByName(ctx, in.Name)
		if err != nil {
			klog.Errorf("Failed to count stores by name: %v", err)
			return nil, errors.ErrServerInternal
		}

		items, err = s.factory.Store().ListByName(ctx, in.Name, in.Offset, in.Limit)
		if err != nil {
			klog.Errorf("Failed to list stores by name: %v", err)
			return nil, errors.ErrServerInternal
		}
	} else {
		// 使用原有的查询所有的方法
		total, err = s.factory.Store().Count(ctx)
		if err != nil {
			klog.Errorf("Failed to count stores: %v", err)
			return nil, errors.ErrServerInternal
		}

		items, err = s.factory.Store().List(ctx, in.Offset, in.Limit)
		if err != nil {
			klog.Errorf("Failed to list stores: %v", err)
			return nil, errors.ErrServerInternal
		}
	}

	return &ListResponse{
		Total: total,
		Items: items,
	}, nil
}

func (s *store) Update(ctx context.Context, id int64, in UpdateRequest) (*model.Store, error) {
	obj, err := s.factory.Store().GetById(ctx, id)
	if err != nil {
		klog.Errorf("Failed to get store by id %d: %v", id, err)
		return nil, errors.ErrServerInternal
	}

	// 如果更新了代码，检查是否与其他门店重复
	if in.Code != "" && in.Code != obj.Code {
		if _, err := s.factory.Store().GetByCode(ctx, in.Code); err == nil {
			klog.Errorf("Store code %s already exists", in.Code)
			return nil, errors.ErrInvalidRequest
		}
		obj.Code = in.Code
	}

	if in.Name != "" {
		obj.Name = in.Name
	}
	if in.Address != "" {
		obj.Address = in.Address
	}
	if in.Contact != "" {
		obj.Contact = in.Contact
	}
	if in.Phone != "" {
		obj.Phone = in.Phone
	}
	if in.Status >= 0 {
		obj.Status = in.Status
	}

	if err := s.factory.Store().Update(ctx, obj); err != nil {
		klog.Errorf("Failed to update store: %v", err)
		return nil, errors.ErrServerInternal
	}
	return obj, nil
}

func (s *store) Delete(ctx context.Context, id int64) error {
	// 检查是否有点位关联
	count, err := s.factory.Location().CountByStoreId(ctx, id)
	if err != nil {
		klog.Errorf("Failed to count locations for store %d: %v", id, err)
		return errors.ErrServerInternal
	}
	if count > 0 {
		klog.Errorf("Cannot delete store %d, it has %d locations", id, count)
		return errors.ErrInvalidRequest
	}

	if err := s.factory.Store().Delete(ctx, id); err != nil {
		klog.Errorf("Failed to delete store: %v", err)
		return errors.ErrServerInternal
	}
	return nil
}

func NewStore(cfg config.Config, f db.ShareDaoFactory) *store {
	return &store{cc: cfg, factory: f}
}
