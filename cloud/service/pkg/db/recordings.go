package db

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
)

type RecordingInterface interface {
	Create(ctx context.Context, object *model.Recording) (*model.Recording, error)
	List(ctx context.Context, in ListRequest) ([]*model.Recording, error)
	Count(ctx context.Context, in ListRequest) (int64, error)
	Query(ctx context.Context, in QueryRequest) ([]*model.Recording, error)
}

type ListRequest struct {
	// 分页参数
	Offset int
	Limit  int

	// 时间范围
	StartTime int64
	EndTime   int64

	// 设备时间范围
	DeviceStartTime int64
	DeviceEndTime   int64

	// 业务参数
	StoreID    int
	LocationID int
	MacAddress []string // 支持多个MAC地址

	// 状态过滤
	Status *int8 // 支持状态过滤，使用指针类型表示可选
}

type QueryRequest struct {
	DeviceId       string
	StartTimestamp int64
	EndTimestamp   int64
	Sid            string
}

type recording struct {
	db *gorm.DB
}

func newRecording(db *gorm.DB) RecordingInterface { return &recording{db: db} }

func (r *recording) Create(ctx context.Context, object *model.Recording) (*model.Recording, error) {
	// 设置创建时间为当前毫秒时间戳
	object.CreatedAtMs = time.Now().UnixMilli()
	if err := r.db.WithContext(ctx).Create(object).Error; err != nil {
		return nil, err
	}
	return object, nil
}

func (r *recording) List(ctx context.Context, in ListRequest) ([]*model.Recording, error) {
	// 使用LEFT JOIN查询，确保未分配设备的录音数据也能被查询到
	query := r.db.WithContext(ctx).Table("recordings").
		Select("recordings.*").
		Joins("LEFT JOIN devices ON recordings.mac_address = devices.mac_address").
		Joins("LEFT JOIN locations ON devices.location_id = locations.id")

	// 时间范围查询
	if in.StartTime > 0 {
		query = query.Where("recordings.created_at >= ?", in.StartTime)
	}

	if in.EndTime > 0 {
		query = query.Where("recordings.created_at <= ?", in.EndTime)
	}

	// 设备时间范围查询
	if in.DeviceStartTime > 0 {
		query = query.Where("recordings.device_time >= ?", in.DeviceStartTime)
	}

	if in.DeviceEndTime > 0 {
		query = query.Where("recordings.device_time <= ?", in.DeviceEndTime)
	}

	// 门店查询
	if in.StoreID > 0 {
		query = query.Where("locations.store_id = ?", in.StoreID)
	}

	// 点位查询
	if in.LocationID > 0 {
		query = query.Where("locations.id = ?", in.LocationID)
	}

	// MAC地址查询 - 支持多个
	if len(in.MacAddress) > 0 {
		// 过滤掉空字符串并转换为小写
		var validMacs []string
		for _, mac := range in.MacAddress {
			if mac != "" {
				validMacs = append(validMacs, strings.ToLower(mac))
			}
		}

		if len(validMacs) > 0 {
			query = query.Where("recordings.mac_address IN ?", validMacs)
		}
	}

	// 状态过滤
	if in.Status != nil {
		query = query.Where("recordings.status = ?", *in.Status)
	}

	// 分页查询
	var recordings []*model.Recording
	err := query.Offset(in.Offset).Limit(in.Limit).Order("recordings.device_time desc").Find(&recordings).Error
	return recordings, err
}

func (r *recording) Count(ctx context.Context, in ListRequest) (int64, error) {
	// 使用LEFT JOIN查询，与List方法保持一致的查询条件
	query := r.db.WithContext(ctx).Table("recordings").
		Joins("LEFT JOIN devices ON recordings.mac_address = devices.mac_address").
		Joins("LEFT JOIN locations ON devices.location_id = locations.id")

	// 时间范围查询
	if in.StartTime > 0 {
		query = query.Where("recordings.created_at >= ?", in.StartTime)
	}

	if in.EndTime > 0 {
		query = query.Where("recordings.created_at <= ?", in.EndTime)
	}

	// 设备时间范围查询
	if in.DeviceStartTime > 0 {
		query = query.Where("recordings.device_time >= ?", in.DeviceStartTime)
	}

	if in.DeviceEndTime > 0 {
		query = query.Where("recordings.device_time <= ?", in.DeviceEndTime)
	}

	// 门店查询
	if in.StoreID > 0 {
		query = query.Where("locations.store_id = ?", in.StoreID)
	}

	// 点位查询
	if in.LocationID > 0 {
		query = query.Where("locations.id = ?", in.LocationID)
	}

	// MAC地址查询 - 支持多个
	if len(in.MacAddress) > 0 {
		// 过滤掉空字符串并转换为小写
		var validMacs []string
		for _, mac := range in.MacAddress {
			if mac != "" {
				validMacs = append(validMacs, strings.ToLower(mac))
			}
		}

		if len(validMacs) > 0 {
			query = query.Where("recordings.mac_address IN ?", validMacs)
		}
	}

	// 状态过滤
	if in.Status != nil {
		query = query.Where("recordings.status = ?", *in.Status)
	}

	var count int64
	err := query.Count(&count).Error
	return count, err
}

func (r *recording) Query(ctx context.Context, in QueryRequest) ([]*model.Recording, error) {
	query := r.db.WithContext(ctx)

	// 按设备ID（MAC地址）过滤
	if in.DeviceId != "" {
		query = query.Where("mac_address = ?", strings.ToLower(in.DeviceId))
	}

	// 按时间范围过滤
	if in.StartTimestamp > 0 {
		query = query.Where("device_time >= ?", in.StartTimestamp)
	}
	if in.EndTimestamp > 0 {
		query = query.Where("device_time <= ?", in.EndTimestamp)
	}

	// 按会话ID过滤（如果模型中有session_id字段）
	// 注意：当前Recording模型中没有session_id字段，这里先注释掉
	// if in.Sid != "" {
	//     query = query.Where("session_id = ?", in.Sid)
	// }

	// 按时间倒序排列
	var recordings []*model.Recording
	err := query.Order("device_time desc").Find(&recordings).Error
	return recordings, err
}
