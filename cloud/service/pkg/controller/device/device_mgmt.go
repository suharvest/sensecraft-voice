package device

import (
	"context"
	goerrors "errors"
	"net/http"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/errors"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/util/crypto"
	"k8s.io/klog/v2"
)

// ErrNoAsrServerAssigned 设备尚未分配 ASR 服务器
var ErrNoAsrServerAssigned = errors.NewError(goerrors.New("设备未分配 ASR 服务器"), http.StatusNotFound)

// GetById 管理端按数字 id 查询设备详情
func (d *device) GetById(ctx context.Context, id int64) (*model.Device, error) {
	if id <= 0 {
		return nil, errors.ErrInvalidRequest
	}
	res, err := d.factory.Device().GetById(ctx, id)
	if err != nil {
		klog.Errorf("Failed to get device by id %d: %v", id, err)
		return nil, errors.ErrServerInternal
	}
	d.markOnline(res)
	return res, nil
}

// Update 管理端更新设备（名称 / 点位 / 门店）
func (d *device) Update(ctx context.Context, deviceId int64, in UpdateRequest) (*model.Device, error) {
	if deviceId <= 0 {
		return nil, errors.ErrInvalidRequest
	}

	updates := map[string]interface{}{}
	if in.Name != nil {
		updates["name"] = *in.Name
	}
	if in.LocationId != nil {
		updates["location_id"] = *in.LocationId
	}
	if in.StoreId != nil {
		updates["store_id"] = *in.StoreId
	}

	res, err := d.factory.Device().Update(ctx, deviceId, updates)
	if err != nil {
		klog.Errorf("Failed to update device %d: %v", deviceId, err)
		return nil, errors.ErrServerInternal
	}
	d.markOnline(res)
	return res, nil
}

// AssignAsrServer 把设备分配到某台 ASR 服务器；asrServerId=0 表示解除分配。
// 分配关系变更时 asr_config_version++，设备下次心跳发现版本变化即拉新配置。
func (d *device) AssignAsrServer(ctx context.Context, deviceId, asrServerId int64) (*model.Device, error) {
	if deviceId <= 0 {
		return nil, errors.ErrInvalidRequest
	}
	if asrServerId > 0 {
		if _, err := d.factory.AsrServer().GetById(ctx, asrServerId); err != nil {
			klog.Errorf("asr server %d not found: %v", asrServerId, err)
			return nil, errors.ErrInvalidRequest
		}
	}

	res, err := d.factory.Device().AssignAsrServer(ctx, deviceId, asrServerId)
	if err != nil {
		klog.Errorf("Failed to assign device %d to asr server %d: %v", deviceId, asrServerId, err)
		return nil, errors.ErrServerInternal
	}
	d.markOnline(res)
	return res, nil
}

// GetAsrConfig 生成设备侧可直接使用的 ASR 配置（api_key 解密后下发）
func (d *device) GetAsrConfig(ctx context.Context, dev *model.Device) (*AsrConfigResponse, error) {
	if dev == nil {
		return nil, errors.ErrUnauthorized
	}
	if dev.AsrServerId <= 0 {
		klog.V(4).Infof("device %d has no asr server assigned", dev.Id)
		return nil, ErrNoAsrServerAssigned
	}

	server, err := d.factory.AsrServer().GetById(ctx, dev.AsrServerId)
	if err != nil {
		klog.Errorf("Failed to load asr server %d for device %d: %v", dev.AsrServerId, dev.Id, err)
		return nil, errors.ErrServerInternal
	}

	apiKey := ""
	if server.ApiKeyCipher != "" {
		cipher, err := crypto.NewCipher(d.cc.Default.CryptoMasterKey)
		if err != nil {
			klog.Errorf("Failed to init cipher: %v", err)
			return nil, errors.ErrServerInternal
		}
		apiKey, err = cipher.Decrypt(server.ApiKeyCipher)
		if err != nil {
			klog.Errorf("Failed to decrypt api key of asr server %d: %v", server.Id, err)
			return nil, errors.ErrServerInternal
		}
	}

	return &AsrConfigResponse{
		// 一期走 OVS 的 OpenAI 兼容路由（POST /v1/audio/transcriptions），
		// 设备侧沿用现成的 openai_whisper adapter，零改动
		Code: "openai_whisper",
		ConfigJson: AsrConfigJson{
			BaseURL: server.BaseUrl,
			ApiKey:  apiKey,
		},
		Version:    dev.AsrConfigVersion,
		ServerId:   server.Id,
		ServerName: server.Name,
	}, nil
}
