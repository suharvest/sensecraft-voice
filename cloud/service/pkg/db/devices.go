package db

import (
	"context"
	"time"

	"gorm.io/gorm"
	"k8s.io/klog/v2"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
)

type DeviceInterface interface {
	Upsert(ctx context.Context, object *model.Device) (*model.Device, error)
	GetById(ctx context.Context, id int64) (*model.Device, error)
	GetByMac(ctx context.Context, mac string) (*model.Device, error)
	List(ctx context.Context, offset, limit int) ([]*model.Device, error)
	ListByFilters(ctx context.Context, name, macAddress string, offset, limit int) ([]*model.Device, error)
	Count(ctx context.Context) (int64, error)
	CountByFilters(ctx context.Context, name, macAddress string) (int64, error)
	ListByLocationId(ctx context.Context, locationId int64, offset, limit int) ([]*model.Device, error)
	ListByStoreId(ctx context.Context, storeId int64, offset, limit int) ([]*model.Device, error)
	CountByLocationId(ctx context.Context, locationId int64) (int64, error)
	CountByStoreId(ctx context.Context, storeId int64) (int64, error)
	AssignToLocation(ctx context.Context, deviceId, locationId int64) error
	UpdateName(ctx context.Context, deviceId int64, name string) (*model.Device, error)

	// 设备认证与在线状态
	GetByTokenHash(ctx context.Context, tokenHash string) (*model.Device, error)
	SetTokenHash(ctx context.Context, deviceId int64, tokenHash string) error
	TouchLastSeen(ctx context.Context, deviceId int64, at int64) error
	ListAdvanced(ctx context.Context, f DeviceFilter, offset, limit int) ([]*model.Device, error)
	CountAdvanced(ctx context.Context, f DeviceFilter) (int64, error)
	Update(ctx context.Context, deviceId int64, updates map[string]interface{}) (*model.Device, error)

	// ASR 配置下发
	AssignAsrServer(ctx context.Context, deviceId, asrServerId int64) (*model.Device, error)
	ReportAsrConfig(ctx context.Context, deviceId int64, errMsg string) error
	ClearAsrServer(ctx context.Context, asrServerId int64) error
}

// DeviceFilter 设备列表过滤条件
type DeviceFilter struct {
	Name       string
	MacAddress string
	LocationId int64
	StoreId    int64
	// Online 为 nil 表示不按在线状态过滤；非 nil 时与 OnlineSince 一起生效
	Online      *bool
	OnlineSince int64
}

type device struct {
	db *gorm.DB
}

func newDevice(db *gorm.DB) DeviceInterface { return &device{db: db} }

func (d *device) Upsert(ctx context.Context, object *model.Device) (*model.Device, error) {
	// 按 mac_address 幂等更新
	now := time.Now().UnixMilli()
	var exist model.Device
	tx := d.db.WithContext(ctx)
	if err := tx.Where("mac_address = ?", object.MacAddress).First(&exist).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新设备时设置时间字段
			object.CreatedAt = now
			object.UpdatedAt = now
			if err := tx.Create(object).Error; err != nil {
				klog.Errorf("Create device error: %v", err)
				return nil, err
			}
			klog.Infof("Create device success: %v", object)
			return object, nil
		}
		klog.Errorf("Upsert device error: %v", err)
		return nil, err
	}

	updates := map[string]interface{}{
		"ip_address":         object.IpAddress,
		"version":            object.Version,
		"cpu_usage_percent":  object.CpuUsagePercent,
		"memory_used_bytes":  object.MemoryUsedBytes,
		"memory_total_bytes": object.MemoryTotalBytes,
		"disk_used_bytes":    object.DiskUsedBytes,
		"disk_total_bytes":   object.DiskTotalBytes,
		"swap_used_bytes":    object.SwapUsedBytes,
		"swap_total_bytes":   object.SwapTotalBytes,
		"updated_at":         now,
	}
	if err := tx.Model(&exist).Updates(updates).Error; err != nil {
		klog.Errorf("Update device error: %v", err)
		return nil, err
	}
	// 更新返回对象的字段
	exist.Name = object.Name
	exist.IpAddress = object.IpAddress
	exist.Version = object.Version
	exist.CpuUsagePercent = object.CpuUsagePercent
	exist.MemoryUsedBytes = object.MemoryUsedBytes
	exist.MemoryTotalBytes = object.MemoryTotalBytes
	exist.DiskUsedBytes = object.DiskUsedBytes
	exist.DiskTotalBytes = object.DiskTotalBytes
	exist.SwapUsedBytes = object.SwapUsedBytes
	exist.SwapTotalBytes = object.SwapTotalBytes
	exist.UpdatedAt = now
	// 回填主键到入参，避免调用方拿到 Id=0 的对象（历史坑）
	object.Id = exist.Id
	object.CreatedAt = exist.CreatedAt
	return &exist, nil
}

func (d *device) GetByTokenHash(ctx context.Context, tokenHash string) (*model.Device, error) {
	var obj model.Device
	if tokenHash == "" {
		return nil, gorm.ErrRecordNotFound
	}
	err := d.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&obj).Error
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func (d *device) SetTokenHash(ctx context.Context, deviceId int64, tokenHash string) error {
	return d.db.WithContext(ctx).Model(&model.Device{}).Where("id = ?", deviceId).
		Updates(map[string]interface{}{
			"token_hash": tokenHash,
			"updated_at": time.Now().UnixMilli(),
		}).Error
}

func (d *device) TouchLastSeen(ctx context.Context, deviceId int64, at int64) error {
	return d.db.WithContext(ctx).Model(&model.Device{}).Where("id = ?", deviceId).
		Update("last_seen_at", at).Error
}

func (d *device) Update(ctx context.Context, deviceId int64, updates map[string]interface{}) (*model.Device, error) {
	if len(updates) == 0 {
		return d.GetById(ctx, deviceId)
	}
	updates["updated_at"] = time.Now().UnixMilli()
	if err := d.db.WithContext(ctx).Model(&model.Device{}).Where("id = ?", deviceId).
		Updates(updates).Error; err != nil {
		return nil, err
	}
	return d.GetById(ctx, deviceId)
}

// applyDeviceFilter 组装列表/计数共用的过滤条件
func applyDeviceFilter(query *gorm.DB, f DeviceFilter) *gorm.DB {
	if f.Name != "" {
		query = query.Where("devices.name LIKE ?", "%"+f.Name+"%")
	}
	if f.MacAddress != "" {
		query = query.Where("devices.mac_address LIKE ?", "%"+f.MacAddress+"%")
	}
	if f.LocationId > 0 {
		query = query.Where("devices.location_id = ?", f.LocationId)
	}
	if f.StoreId > 0 {
		query = query.Where("devices.store_id = ?", f.StoreId)
	}
	if f.Online != nil {
		if *f.Online {
			query = query.Where("devices.last_seen_at > ?", f.OnlineSince)
		} else {
			query = query.Where("devices.last_seen_at <= ?", f.OnlineSince)
		}
	}
	return query
}

func (d *device) ListAdvanced(ctx context.Context, f DeviceFilter, offset, limit int) ([]*model.Device, error) {
	var devices []*model.Device
	query := d.db.WithContext(ctx).Model(&model.Device{}).
		Select(`devices.*,
			locations.name as location_name,
			stores.name as store_name,
			(SELECT MAX(created_at) FROM recordings WHERE mac_address = devices.mac_address) as latest_recording_time`).
		Joins("LEFT JOIN locations ON devices.location_id = locations.id").
		Joins("LEFT JOIN stores ON devices.store_id = stores.id")

	err := applyDeviceFilter(query, f).
		Offset(offset).Limit(limit).Order("devices.id desc").Find(&devices).Error
	return devices, err
}

func (d *device) CountAdvanced(ctx context.Context, f DeviceFilter) (int64, error) {
	var count int64
	query := d.db.WithContext(ctx).Model(&model.Device{})
	err := applyDeviceFilter(query, f).Count(&count).Error
	return count, err
}

// AssignAsrServer 分配设备到 ASR 服务器；服务器变更时 asr_config_version++ 并清空上次生效状态
func (d *device) AssignAsrServer(ctx context.Context, deviceId, asrServerId int64) (*model.Device, error) {
	var exist model.Device
	if err := d.db.WithContext(ctx).Where("id = ?", deviceId).First(&exist).Error; err != nil {
		return nil, err
	}
	if exist.AsrServerId == asrServerId {
		// 未变更：不动版本号，保持幂等
		return &exist, nil
	}

	updates := map[string]interface{}{
		"asr_server_id":         asrServerId,
		"asr_config_version":    exist.AsrConfigVersion + 1,
		"asr_config_applied_at": 0,
		"asr_config_error":      "",
		"updated_at":            time.Now().UnixMilli(),
	}
	if err := d.db.WithContext(ctx).Model(&model.Device{}).Where("id = ?", deviceId).
		Updates(updates).Error; err != nil {
		return nil, err
	}
	return d.GetById(ctx, deviceId)
}

// ReportAsrConfig 记录设备上报的 ASR 配置生效结果（版本一致性由调用方判定）
func (d *device) ReportAsrConfig(ctx context.Context, deviceId int64, errMsg string) error {
	updates := map[string]interface{}{
		"asr_config_applied_at": time.Now().UnixMilli(),
		"asr_config_error":      errMsg,
		"updated_at":            time.Now().UnixMilli(),
	}
	return d.db.WithContext(ctx).Model(&model.Device{}).Where("id = ?", deviceId).Updates(updates).Error
}

// ClearAsrServer 服务器被删除时解除所有设备的分配关系，并 bump 版本号触发重新下发
func (d *device) ClearAsrServer(ctx context.Context, asrServerId int64) error {
	return d.db.WithContext(ctx).Model(&model.Device{}).
		Where("asr_server_id = ?", asrServerId).
		Updates(map[string]interface{}{
			"asr_server_id":      0,
			"asr_config_version": gorm.Expr("asr_config_version + 1"),
			"asr_config_error":   "",
			"updated_at":         time.Now().UnixMilli(),
		}).Error
}

func (d *device) GetById(ctx context.Context, id int64) (*model.Device, error) {
	var device model.Device
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (d *device) GetByMac(ctx context.Context, mac string) (*model.Device, error) {
	var device model.Device
	err := d.db.WithContext(ctx).Where("mac_address = ?", mac).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (d *device) List(ctx context.Context, offset, limit int) ([]*model.Device, error) {
	var devices []*model.Device
	err := d.db.WithContext(ctx).
		Select(`devices.*, 
			locations.name as location_name, 
			stores.name as store_name,
			(SELECT MAX(created_at) FROM recordings WHERE mac_address = devices.mac_address) as latest_recording_time`).
		Joins("LEFT JOIN locations ON devices.location_id = locations.id").
		Joins("LEFT JOIN stores ON devices.store_id = stores.id").
		Offset(offset).Limit(limit).Order("devices.id desc").Find(&devices).Error
	return devices, err
}

func (d *device) Count(ctx context.Context) (int64, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&model.Device{}).Count(&count).Error
	return count, err
}

func (d *device) ListByLocationId(ctx context.Context, locationId int64, offset, limit int) ([]*model.Device, error) {
	var devices []*model.Device
	err := d.db.WithContext(ctx).
		Select(`devices.*, 
			locations.name as location_name, 
			stores.name as store_name,
			(SELECT MAX(created_at) FROM recordings WHERE mac_address = devices.mac_address) as latest_recording_time`).
		Joins("LEFT JOIN locations ON devices.location_id = locations.id").
		Joins("LEFT JOIN stores ON devices.store_id = stores.id").
		Where("devices.location_id = ?", locationId).
		Offset(offset).Limit(limit).Order("devices.id desc").Find(&devices).Error
	return devices, err
}

func (d *device) ListByStoreId(ctx context.Context, storeId int64, offset, limit int) ([]*model.Device, error) {
	var devices []*model.Device
	err := d.db.WithContext(ctx).
		Select(`devices.*, 
			locations.name as location_name, 
			stores.name as store_name,
			(SELECT MAX(created_at) FROM recordings WHERE mac_address = devices.mac_address) as latest_recording_time`).
		Joins("LEFT JOIN locations ON devices.location_id = locations.id").
		Joins("LEFT JOIN stores ON devices.store_id = stores.id").
		Where("devices.store_id = ?", storeId).
		Offset(offset).Limit(limit).Order("devices.id desc").Find(&devices).Error
	return devices, err
}

func (d *device) CountByLocationId(ctx context.Context, locationId int64) (int64, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&model.Device{}).Where("location_id = ?", locationId).Count(&count).Error
	return count, err
}

func (d *device) CountByStoreId(ctx context.Context, storeId int64) (int64, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&model.Device{}).Where("store_id = ?", storeId).Count(&count).Error
	return count, err
}

func (d *device) AssignToLocation(ctx context.Context, deviceId, locationId int64) error {
	// 获取点位信息以确定门店ID
	var location struct {
		Id      int64 `gorm:"column:id"`
		StoreId int64 `gorm:"column:store_id"`
	}
	err := d.db.WithContext(ctx).Table("locations").Where("id = ?", locationId).First(&location).Error
	if err != nil {
		return err
	}

	// 更新设备的点位和门店关联
	updates := map[string]interface{}{
		"location_id": locationId,
		"store_id":    location.StoreId,
		"updated_at":  time.Now().UnixMilli(),
	}

	return d.db.WithContext(ctx).Model(&model.Device{}).Where("id = ?", deviceId).Updates(updates).Error
}

func (d *device) UpdateName(ctx context.Context, deviceId int64, name string) (*model.Device, error) {
	// 先检查设备是否存在
	var device model.Device
	err := d.db.WithContext(ctx).Where("id = ?", deviceId).First(&device).Error
	if err != nil {
		return nil, err
	}

	// 更新设备名称
	updates := map[string]interface{}{
		"name":       name,
		"updated_at": time.Now().UnixMilli(),
	}

	err = d.db.WithContext(ctx).Model(&device).Updates(updates).Error
	if err != nil {
		return nil, err
	}

	// 更新返回对象的名称字段
	device.Name = name
	device.UpdatedAt = time.Now().UnixMilli()

	return &device, nil
}

func (d *device) ListByFilters(ctx context.Context, name, macAddress string, offset, limit int) ([]*model.Device, error) {
	var devices []*model.Device
	query := d.db.WithContext(ctx).
		Select(`devices.*, 
			locations.name as location_name, 
			stores.name as store_name,
			(SELECT MAX(created_at) FROM recordings WHERE mac_address = devices.mac_address) as latest_recording_time`).
		Joins("LEFT JOIN locations ON devices.location_id = locations.id").
		Joins("LEFT JOIN stores ON devices.store_id = stores.id")

	// 添加过滤条件
	if name != "" {
		query = query.Where("devices.name LIKE ?", "%"+name+"%")
	}
	if macAddress != "" {
		query = query.Where("devices.mac_address LIKE ?", "%"+macAddress+"%")
	}

	err := query.Offset(offset).Limit(limit).Order("devices.id desc").Find(&devices).Error
	return devices, err
}

func (d *device) CountByFilters(ctx context.Context, name, macAddress string) (int64, error) {
	var count int64
	query := d.db.WithContext(ctx).Model(&model.Device{})

	// 添加过滤条件
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if macAddress != "" {
		query = query.Where("mac_address LIKE ?", "%"+macAddress+"%")
	}

	err := query.Count(&count).Error
	return count, err
}
