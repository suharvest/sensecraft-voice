package asr_server

import (
	"context"
	"strings"
	"time"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/errors"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/config"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/service"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/util/crypto"
	"k8s.io/klog/v2"
)

type AsrServerGetter interface {
	AsrServer() Interface
}

type Interface interface {
	Create(ctx context.Context, in CreateRequest) (*model.AsrServer, error)
	Update(ctx context.Context, id int64, in UpdateRequest) (*model.AsrServer, error)
	Delete(ctx context.Context, id int64) error
	GetById(ctx context.Context, id int64) (*model.AsrServer, error)
	List(ctx context.Context, in ListRequest) (*ListResponse, error)
	Probe(ctx context.Context, id int64) (*model.AsrServer, error)
}

type CreateRequest struct {
	Name       string `json:"name"`
	BaseUrl    string `json:"base_url" binding:"required"`
	Platform   string `json:"platform"`
	ApiKey     string `json:"api_key"`
	LocationId int64  `json:"location_id"`
}

type UpdateRequest struct {
	Name       *string `json:"name,omitempty"`
	BaseUrl    *string `json:"base_url,omitempty"`
	Platform   *string `json:"platform,omitempty"`
	ApiKey     *string `json:"api_key,omitempty"` // 传空字符串表示清空
	LocationId *int64  `json:"location_id,omitempty"`
}

type ListRequest struct {
	Offset int `form:"offset" binding:"min=0"`
	Limit  int `form:"limit" binding:"min=0,max=100"`
}

type ListResponse struct {
	Total int64              `json:"total"`
	Items []*model.AsrServer `json:"items"`
}

type asrServer struct {
	cc      config.Config
	factory db.ShareDaoFactory
}

func NewAsrServer(cfg config.Config, f db.ShareDaoFactory) *asrServer {
	return &asrServer{cc: cfg, factory: f}
}

func (a *asrServer) cipher() (*crypto.Cipher, error) {
	return crypto.NewCipher(a.cc.Default.CryptoMasterKey)
}

func (a *asrServer) probeTimeout() time.Duration {
	sec := a.cc.ASR.TimeoutSeconds
	if sec <= 0 {
		sec = 5
	}
	return time.Duration(sec) * time.Second
}

// decryptApiKey 取出服务器的明文 api_key（仅服务端内部使用，不出 API）
func (a *asrServer) decryptApiKey(server *model.AsrServer) (string, error) {
	if server.ApiKeyCipher == "" {
		return "", nil
	}
	c, err := a.cipher()
	if err != nil {
		return "", err
	}
	return c.Decrypt(server.ApiKeyCipher)
}

func normalizeBaseUrl(in string) string {
	return strings.TrimRight(strings.TrimSpace(in), "/")
}

func (a *asrServer) Create(ctx context.Context, in CreateRequest) (*model.AsrServer, error) {
	baseUrl := normalizeBaseUrl(in.BaseUrl)
	if baseUrl == "" {
		return nil, errors.ErrInvalidRequest
	}

	obj := &model.AsrServer{
		Name:       strings.TrimSpace(in.Name),
		BaseUrl:    baseUrl,
		Platform:   strings.TrimSpace(in.Platform),
		LocationId: in.LocationId,
		Status:     model.AsrServerStatusUnknown,
	}

	if in.ApiKey != "" {
		c, err := a.cipher()
		if err != nil {
			klog.Errorf("Failed to init cipher: %v", err)
			return nil, errors.ErrServerInternal
		}
		enc, err := c.Encrypt(in.ApiKey)
		if err != nil {
			klog.Errorf("Failed to encrypt api key: %v", err)
			return nil, errors.ErrServerInternal
		}
		obj.ApiKeyCipher = enc
	}

	saved, err := a.factory.AsrServer().Create(ctx, obj)
	if err != nil {
		klog.Errorf("Failed to create asr server: %v", err)
		return nil, errors.ErrServerInternal
	}

	// 保存后立即探测一次，回填 backend / capabilities / status
	if probed, err := a.Probe(ctx, saved.Id); err == nil {
		return probed, nil
	}
	return a.decorate(ctx, saved), nil
}

func (a *asrServer) Update(ctx context.Context, id int64, in UpdateRequest) (*model.AsrServer, error) {
	if id <= 0 {
		return nil, errors.ErrInvalidRequest
	}
	if _, err := a.factory.AsrServer().GetById(ctx, id); err != nil {
		klog.Errorf("asr server %d not found: %v", id, err)
		return nil, errors.ErrInvalidRequest
	}

	updates := map[string]interface{}{}
	if in.Name != nil {
		updates["name"] = strings.TrimSpace(*in.Name)
	}
	if in.BaseUrl != nil {
		bu := normalizeBaseUrl(*in.BaseUrl)
		if bu == "" {
			return nil, errors.ErrInvalidRequest
		}
		updates["base_url"] = bu
	}
	if in.Platform != nil {
		updates["platform"] = strings.TrimSpace(*in.Platform)
	}
	if in.LocationId != nil {
		updates["location_id"] = *in.LocationId
	}
	if in.ApiKey != nil {
		if *in.ApiKey == "" {
			updates["api_key_cipher"] = ""
		} else {
			c, err := a.cipher()
			if err != nil {
				klog.Errorf("Failed to init cipher: %v", err)
				return nil, errors.ErrServerInternal
			}
			enc, err := c.Encrypt(*in.ApiKey)
			if err != nil {
				klog.Errorf("Failed to encrypt api key: %v", err)
				return nil, errors.ErrServerInternal
			}
			updates["api_key_cipher"] = enc
		}
	}

	if err := a.factory.AsrServer().Update(ctx, id, updates); err != nil {
		klog.Errorf("Failed to update asr server %d: %v", id, err)
		return nil, errors.ErrServerInternal
	}

	// 地址/密钥可能变了，重新探测回填
	if probed, err := a.Probe(ctx, id); err == nil {
		return probed, nil
	}
	return a.GetById(ctx, id)
}

func (a *asrServer) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.ErrInvalidRequest
	}
	// 解除设备分配关系并 bump 版本，设备下次心跳会察觉配置变化
	if err := a.factory.Device().ClearAsrServer(ctx, id); err != nil {
		klog.Errorf("Failed to clear device assignment of asr server %d: %v", id, err)
		return errors.ErrServerInternal
	}
	if err := a.factory.AsrServer().Delete(ctx, id); err != nil {
		klog.Errorf("Failed to delete asr server %d: %v", id, err)
		return errors.ErrServerInternal
	}
	return nil
}

func (a *asrServer) GetById(ctx context.Context, id int64) (*model.AsrServer, error) {
	if id <= 0 {
		return nil, errors.ErrInvalidRequest
	}
	obj, err := a.factory.AsrServer().GetById(ctx, id)
	if err != nil {
		klog.Errorf("Failed to get asr server %d: %v", id, err)
		return nil, errors.ErrServerInternal
	}
	return a.decorate(ctx, obj), nil
}

func (a *asrServer) List(ctx context.Context, in ListRequest) (*ListResponse, error) {
	if in.Limit == 0 {
		in.Limit = 20
	}
	total, err := a.factory.AsrServer().Count(ctx)
	if err != nil {
		klog.Errorf("Failed to count asr servers: %v", err)
		return nil, errors.ErrServerInternal
	}
	items, err := a.factory.AsrServer().List(ctx, in.Offset, in.Limit)
	if err != nil {
		klog.Errorf("Failed to list asr servers: %v", err)
		return nil, errors.ErrServerInternal
	}
	for _, item := range items {
		a.decorate(ctx, item)
	}
	return &ListResponse{Total: total, Items: items}, nil
}

// decorate 回填不落库的展示字段（是否配置了 key、已分配设备数）
func (a *asrServer) decorate(ctx context.Context, obj *model.AsrServer) *model.AsrServer {
	if obj == nil {
		return nil
	}
	obj.HasApiKey = obj.ApiKeyCipher != ""
	if count, err := a.factory.AsrServer().CountDevices(ctx, obj.Id); err == nil {
		obj.DeviceCount = count
	}
	return obj
}

// Probe 探测服务器 /readyz + /asr/capabilities，回填 status/backend/capabilities。
// 手工保存/编辑时立即调用；定时任务走 jobmanager 的 asr-server-prober。
func (a *asrServer) Probe(ctx context.Context, id int64) (*model.AsrServer, error) {
	server, err := a.factory.AsrServer().GetById(ctx, id)
	if err != nil {
		klog.Errorf("Failed to get asr server %d: %v", id, err)
		return nil, errors.ErrServerInternal
	}

	apiKey, err := a.decryptApiKey(server)
	if err != nil {
		klog.Errorf("Failed to decrypt api key of asr server %d: %v", id, err)
	}

	result := service.NewAsrProber(a.probeTimeout()).Probe(ctx, server.BaseUrl, apiKey)
	updates := service.BuildProbeUpdates(server, result, a.cc.ASR.FailThreshold)
	if err := a.factory.AsrServer().Update(ctx, id, updates); err != nil {
		klog.Errorf("Failed to persist probe result of asr server %d: %v", id, err)
		return nil, errors.ErrServerInternal
	}
	return a.GetById(ctx, id)
}
