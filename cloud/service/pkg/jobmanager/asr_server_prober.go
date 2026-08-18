package jobmanager

import (
	"context"
	"time"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/service"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/util/crypto"
)

// AsrServerProber 定时探测所有 ASR 服务器的 /readyz + /asr/capabilities。
// 连续失败次数达到阈值（默认 5 次）才置 status=down：OVS 推理同步阻塞 event loop，
// 转写高峰期 /readyz 超时属正常现象（设计文档 §1.5）。
type AsrServerProber struct {
	cfg     AsrProbeOptions
	factory db.ShareDaoFactory
}

// AsrProbeOptions 探测参数（config 侧以 asr: 段落映射，见 cmd/app/config/config.go）
type AsrProbeOptions struct {
	Schedule       string `yaml:"probe_schedule"`
	TimeoutSeconds int    `yaml:"probe_timeout_seconds"`
	FailThreshold  int    `yaml:"probe_fail_threshold"`
	// MasterKey 解密 asr_servers.api_key_cipher 的主密钥，由 options 注入
	MasterKey string `yaml:"-"`
}

func NewAsrServerProber(cfg AsrProbeOptions, f db.ShareDaoFactory) *AsrServerProber {
	return &AsrServerProber{cfg: cfg, factory: f}
}

func (j *AsrServerProber) Name() string { return "asr-server-prober" }

func (j *AsrServerProber) CronSpec() string {
	if j.cfg.Schedule == "" {
		return "*/1 * * * *"
	}
	return j.cfg.Schedule
}

func (j *AsrServerProber) Do(ctx *JobContext) error {
	servers, err := j.factory.AsrServer().ListAll(ctx)
	if err != nil {
		return err
	}
	if len(servers) == 0 {
		ctx.WithLogFields(map[string]interface{}{"servers": 0})
		return nil
	}

	timeout := time.Duration(j.cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	prober := service.NewAsrProber(timeout)

	var cipher *crypto.Cipher
	if c, err := crypto.NewCipher(j.cfg.MasterKey); err == nil {
		cipher = c
	}

	var up, busy, down int
	for _, server := range servers {
		apiKey := ""
		if cipher != nil && server.ApiKeyCipher != "" {
			if plain, err := cipher.Decrypt(server.ApiKeyCipher); err == nil {
				apiKey = plain
			}
		}

		probeCtx, cancel := context.WithTimeout(ctx, timeout+time.Second)
		result := prober.Probe(probeCtx, server.BaseUrl, apiKey)
		cancel()

		switch {
		case result.Busy:
			// /readyz 503 sessions_full：限流器占满，服务器健康（并发上限 rk/orin-nano = 1）
			busy++
		case result.Ready:
			up++
		default:
			down++
		}

		updates := service.BuildProbeUpdates(server, result, j.cfg.FailThreshold)
		if err := j.factory.AsrServer().Update(ctx, server.Id, updates); err != nil {
			ctx.WithLogFields(map[string]interface{}{
				"server_id":    server.Id,
				"update_error": err.Error(),
			})
		}
		if status, ok := updates["status"]; ok && status == model.AsrServerStatusDown {
			ctx.WithLogFields(map[string]interface{}{
				"server_down": server.BaseUrl,
				"fail_count":  updates["fail_count"],
			})
		}
	}

	ctx.WithLogFields(map[string]interface{}{
		"servers":     len(servers),
		"probe_ok":    up,
		"probe_busy":  busy,
		"probe_fail":  down,
		"fail_thresh": j.cfg.FailThreshold,
	})
	return nil
}
