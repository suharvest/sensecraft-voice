package device

import (
	"context"
	"strings"
	"time"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/errors"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/config"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/util/token"
	"k8s.io/klog/v2"
)

type DeviceGetter interface {
	Device() Interface
}

type Interface interface {
	Register(ctx context.Context, in RegisterRequest, auth RegisterAuth) (*RegisterResponse, error)
	AssignToLocation(ctx context.Context, deviceId, locationId int64) error
	UpdateName(ctx context.Context, deviceId int64, name string) (*model.Device, error)
	ListByLocation(ctx context.Context, locationId int64, in ListRequest) (*ListResponse, error)
	ListByStore(ctx context.Context, storeId int64, in ListRequest) (*ListResponse, error)
	List(ctx context.Context, in ListRequest) (*ListResponse, error)
	GetByMac(ctx context.Context, mac string) (*model.Device, error)
	GetById(ctx context.Context, id int64) (*model.Device, error)
	Update(ctx context.Context, deviceId int64, in UpdateRequest) (*model.Device, error)
	AssignAsrServer(ctx context.Context, deviceId, asrServerId int64) (*model.Device, error)
	GetAsrConfig(ctx context.Context, device *model.Device) (*AsrConfigResponse, error)
}

type RegisterRequest struct {
	MacAddress       string  `json:"mac_address"`
	Name             string  `json:"name,omitempty"`
	IpAddress        string  `json:"ip_address,omitempty"`
	Version          string  `json:"version"`
	CpuUsagePercent  float64 `json:"cpu_usage_percent"`
	MemoryUsedBytes  int64   `json:"memory_used_bytes"`
	MemoryTotalBytes int64   `json:"memory_total_bytes"`
	DiskUsedBytes    int64   `json:"disk_used_bytes"`
	DiskTotalBytes   int64   `json:"disk_total_bytes"`
	SwapUsedBytes    int64   `json:"swap_used_bytes"`
	SwapTotalBytes   int64   `json:"swap_total_bytes"`

	// 设备上报的 ASR 配置生效结果（随心跳回传）
	AsrConfigVersion *int    `json:"asr_config_version,omitempty"`
	AsrConfigError   *string `json:"asr_config_error,omitempty"`
}

// RegisterAuth 由 middleware 注入的认证结果
type RegisterAuth struct {
	// Device 已通过 device token 认证的设备（nil 表示未用 token 认证）
	Device *model.Device
	// Enrollment 本次请求通过 enrollment key（或过渡期放行）认证
	Enrollment bool
}

// RegisterResponse 注册/心跳响应
type RegisterResponse struct {
	// Token 仅首次注册（或首次补发凭证）时返回，设备需持久化
	Token            string        `json:"token,omitempty"`
	AsrConfigVersion int           `json:"asr_config_version"`
	ServerTime       int64         `json:"server_time"`
	Device           *model.Device `json:"device,omitempty"`
}

// UpdateRequest 管理端更新设备
type UpdateRequest struct {
	Name       *string `json:"name,omitempty"`
	LocationId *int64  `json:"location_id,omitempty"`
	StoreId    *int64  `json:"store_id,omitempty"`
}

// AsrConfigResponse 下发给设备的 ASR 配置，结构与 respeaker-service 的
// POST /api/v1/asr/config 请求体一致，设备原样转发即可
type AsrConfigResponse struct {
	Code       string        `json:"code"`
	ConfigJson AsrConfigJson `json:"config_json"`
	Version    int           `json:"version"`
	ServerId   int64         `json:"asr_server_id"`
	ServerName string        `json:"asr_server_name,omitempty"`
}

type AsrConfigJson struct {
	BaseURL string `json:"base_url"`
	ApiKey  string `json:"api_key"`
}

type ListRequest struct {
	Offset     int    `form:"offset" binding:"min=0"`
	Limit      int    `form:"limit" binding:"min=1,max=100"`
	Name       string `form:"name"`        // 设备名称，支持模糊搜索
	MacAddress string `form:"mac_address"` // MAC地址，支持模糊搜索
	Online     *bool  `form:"online"`      // 在线状态过滤，不传表示不过滤
}

type ListResponse struct {
	Total int64           `json:"total"`
	Items []*model.Device `json:"items"`
}

type device struct {
	cc      config.Config
	factory db.ShareDaoFactory
}

func (d *device) Register(ctx context.Context, in RegisterRequest, auth RegisterAuth) (*RegisterResponse, error) {
	if in.MacAddress == "" {
		klog.Errorf("MacAddress is empty")
		return nil, errors.ErrInvalidRequest
	}
	mac := strings.ToLower(in.MacAddress)

	// 用 token 认证的请求，mac 必须与凭证归属设备一致，防止串号改写他人记录
	if auth.Device != nil && auth.Device.MacAddress != mac {
		klog.Errorf("device token mac mismatch: token=%s request=%s", auth.Device.MacAddress, mac)
		return nil, errors.ErrUnauthorized
	}

	obj := &model.Device{
		MacAddress:       mac,
		Name:             in.Name,
		IpAddress:        in.IpAddress,
		Version:          in.Version,
		CpuUsagePercent:  in.CpuUsagePercent,
		MemoryUsedBytes:  in.MemoryUsedBytes,
		MemoryTotalBytes: in.MemoryTotalBytes,
		DiskUsedBytes:    in.DiskUsedBytes,
		DiskTotalBytes:   in.DiskTotalBytes,
		SwapUsedBytes:    in.SwapUsedBytes,
		SwapTotalBytes:   in.SwapTotalBytes,
	}

	saved, err := d.factory.Device().Upsert(ctx, obj)
	if err != nil {
		klog.Errorf("Failed to upsert device %s: %v", mac, err)
		return nil, errors.ErrServerInternal
	}
	// Upsert 的 update 分支历史上不回填 Id，这里兜底再查一次
	if saved == nil || saved.Id == 0 {
		saved, err = d.factory.Device().GetByMac(ctx, mac)
		if err != nil {
			klog.Errorf("Failed to reload device %s after upsert: %v", mac, err)
			return nil, errors.ErrServerInternal
		}
	}

	now := time.Now().UnixMilli()

	// 心跳：刷新 last_seen_at
	if err := d.factory.Device().TouchLastSeen(ctx, saved.Id, now); err != nil {
		klog.Errorf("Failed to touch last_seen_at for device %d: %v", saved.Id, err)
	}
	saved.LastSeenAt = now
	saved.Online = true

	// 设备上报 ASR 配置生效结果
	if in.AsrConfigVersion != nil {
		errMsg := ""
		if in.AsrConfigError != nil {
			errMsg = *in.AsrConfigError
		}
		if len(errMsg) > 500 {
			errMsg = errMsg[:500]
		}
		if *in.AsrConfigVersion == saved.AsrConfigVersion {
			if err := d.factory.Device().ReportAsrConfig(ctx, saved.Id, errMsg); err != nil {
				klog.Errorf("Failed to record asr config report for device %d: %v", saved.Id, err)
			} else {
				saved.AsrConfigAppliedAt = now
				saved.AsrConfigError = errMsg
			}
		} else {
			klog.V(4).Infof("device %d reported stale asr_config_version %d (current %d)",
				saved.Id, *in.AsrConfigVersion, saved.AsrConfigVersion)
		}
	}

	resp := &RegisterResponse{
		AsrConfigVersion: saved.AsrConfigVersion,
		ServerTime:       now,
		Device:           saved,
	}

	// 首次注册（或凭证缺失需要补发）时签发不透明 token
	if saved.TokenHash == "" && (auth.Enrollment || auth.Device != nil) {
		plain, err := token.GenerateDeviceToken()
		if err != nil {
			klog.Errorf("Failed to generate device token: %v", err)
			return nil, errors.ErrServerInternal
		}
		if err := d.factory.Device().SetTokenHash(ctx, saved.Id, token.HashDeviceToken(plain)); err != nil {
			klog.Errorf("Failed to persist device token for %d: %v", saved.Id, err)
			return nil, errors.ErrServerInternal
		}
		resp.Token = plain
		klog.Infof("Issued device token for device %d (%s)", saved.Id, mac)
	}

	return resp, nil
}

func (d *device) AssignToLocation(ctx context.Context, deviceId, locationId int64) error {
	if err := d.factory.Device().AssignToLocation(ctx, deviceId, locationId); err != nil {
		klog.Errorf("Failed to assign device %d to location %d: %v", deviceId, locationId, err)
		return errors.ErrServerInternal
	}
	return nil
}

func (d *device) UpdateName(ctx context.Context, deviceId int64, name string) (*model.Device, error) {
	if deviceId <= 0 {
		klog.Errorf("Invalid device ID: %d", deviceId)
		return nil, errors.ErrInvalidRequest
	}

	device, err := d.factory.Device().UpdateName(ctx, deviceId, name)
	if err != nil {
		klog.Errorf("Failed to update device %d name to %s: %v", deviceId, name, err)
		return nil, errors.ErrServerInternal
	}
	d.markOnline(device)
	return device, nil
}

func (d *device) ListByLocation(ctx context.Context, locationId int64, in ListRequest) (*ListResponse, error) {
	if in.Limit == 0 {
		in.Limit = 20
	}

	total, err := d.factory.Device().CountByLocationId(ctx, locationId)
	if err != nil {
		klog.Errorf("Failed to count devices for location %d: %v", locationId, err)
		return nil, errors.ErrServerInternal
	}

	items, err := d.factory.Device().ListByLocationId(ctx, locationId, in.Offset, in.Limit)
	if err != nil {
		klog.Errorf("Failed to list devices for location %d: %v", locationId, err)
		return nil, errors.ErrServerInternal
	}
	d.markOnline(items...)

	return &ListResponse{
		Total: total,
		Items: items,
	}, nil
}

func (d *device) ListByStore(ctx context.Context, storeId int64, in ListRequest) (*ListResponse, error) {
	if in.Limit == 0 {
		in.Limit = 20
	}

	total, err := d.factory.Device().CountByStoreId(ctx, storeId)
	if err != nil {
		klog.Errorf("Failed to count devices for store %d: %v", storeId, err)
		return nil, errors.ErrServerInternal
	}

	items, err := d.factory.Device().ListByStoreId(ctx, storeId, in.Offset, in.Limit)
	if err != nil {
		klog.Errorf("Failed to list devices for store %d: %v", storeId, err)
		return nil, errors.ErrServerInternal
	}
	d.markOnline(items...)

	return &ListResponse{
		Total: total,
		Items: items,
	}, nil
}

func (d *device) List(ctx context.Context, in ListRequest) (*ListResponse, error) {
	if in.Limit == 0 {
		in.Limit = 20
	}

	filter := db.DeviceFilter{
		Name:        in.Name,
		MacAddress:  strings.ToLower(in.MacAddress),
		Online:      in.Online,
		OnlineSince: d.onlineSince(),
	}

	total, err := d.factory.Device().CountAdvanced(ctx, filter)
	if err != nil {
		klog.Errorf("Failed to count devices: %v", err)
		return nil, errors.ErrServerInternal
	}
	items, err := d.factory.Device().ListAdvanced(ctx, filter, in.Offset, in.Limit)
	if err != nil {
		klog.Errorf("Failed to list devices: %v", err)
		return nil, errors.ErrServerInternal
	}
	d.markOnline(items...)

	return &ListResponse{Total: total, Items: items}, nil
}

// onlineSince 在线判定的时间下界：now - 3×心跳间隔
func (d *device) onlineSince() int64 {
	return time.Now().UnixMilli() - d.cc.Default.OnlineWindowMs()
}

// markOnline 按 last_seen_at 计算 online 字段（不落库、不需要扫表任务）
func (d *device) markOnline(items ...*model.Device) {
	since := d.onlineSince()
	for _, item := range items {
		if item == nil {
			continue
		}
		item.Online = item.LastSeenAt > since
	}
}

func (d *device) GetByMac(ctx context.Context, mac string) (*model.Device, error) {
	if mac == "" {
		return nil, errors.ErrInvalidRequest
	}
	res, err := d.factory.Device().GetByMac(ctx, strings.ToLower(mac))
	if err != nil {
		klog.Errorf("Failed to get device by mac %s: %v", mac, err)
		return nil, errors.ErrServerInternal
	}
	d.markOnline(res)
	return res, nil
}

func NewDevice(cfg config.Config, f db.ShareDaoFactory) *device {
	return &device{cc: cfg, factory: f}
}
