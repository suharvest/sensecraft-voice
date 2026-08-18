package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
)

// ReasonSessionsFull /readyz 的 reason：限流器无空余容量。
// RK / orin-nano 的并发上限是 1，一次转写进行中 /readyz 就是 503 sessions_full，
// 这是「忙」不是「坏」，不能计入失败次数（orin-nano 真机实测确认）。
const ReasonSessionsFull = "sessions_full"

// AsrProbeResult 一次探测的结果
type AsrProbeResult struct {
	// Ready 服务器可用（HTTP 200，或 503 但仅因 sessions_full）
	Ready bool
	// Busy 503 且 reasons 只含 sessions_full：正在解码，视为健康
	Busy bool
	// Reasons /readyz 返回的原因列表（503 时有值）
	Reasons      []string
	Backend      string
	Capabilities []string
	SampleRate   int
	// Err 探测失败的原因（Ready=false 时有值；Ready=true 时可能是 capabilities 取不到）
	Err error
}

// asrReadyzResponse /readyz 的响应体
// 形如 {"status":"not_ready","reasons":["sessions_full"]}
type asrReadyzResponse struct {
	Status  string   `json:"status"`
	Reasons []string `json:"reasons"`
}

// onlySessionsFull reasons 是否只包含 sessions_full
func onlySessionsFull(reasons []string) bool {
	if len(reasons) == 0 {
		return false
	}
	for _, r := range reasons {
		if strings.TrimSpace(r) != ReasonSessionsFull {
			return false
		}
	}
	return true
}

// asrCapabilitiesResponse GET /asr/capabilities 的响应
// 形如 {"backend":"rk.asr","capabilities":["offline","multi_language"],"sample_rate":16000}
type asrCapabilitiesResponse struct {
	Backend      string   `json:"backend"`
	Capabilities []string `json:"capabilities"`
	SampleRate   int      `json:"sample_rate"`
}

// AsrProber 探测 ASR 服务器（OVS）的就绪状态与能力
type AsrProber struct {
	client *http.Client
}

func NewAsrProber(timeout time.Duration) *AsrProber {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &AsrProber{client: &http.Client{Timeout: timeout}}
}

// Probe 探测 /readyz，可用后再取 /asr/capabilities。
// capabilities 端点需要 api key，取不到不影响可用性判定（backend 未就绪时它也返回 503）。
func (p *AsrProber) Probe(ctx context.Context, baseURL, apiKey string) AsrProbeResult {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		return AsrProbeResult{Err: fmt.Errorf("base_url 为空")}
	}

	res := p.probeReady(ctx, base)
	if !res.Ready {
		return res
	}

	caps, err := p.fetchCapabilities(ctx, base, apiKey)
	if err != nil {
		res.Err = err
		return res
	}
	res.Backend = caps.Backend
	res.Capabilities = caps.Capabilities
	res.SampleRate = caps.SampleRate
	return res
}

// probeReady 打 /readyz 并按 reasons 区分「忙」与「坏」：
//   - 200                                  -> Ready
//   - 503 且 reasons 只含 sessions_full     -> Ready + Busy（限流器占满，解码中）
//   - 503 含 backend_not_ready 等其他原因   -> 失败
//   - 连接超时/拒绝                          -> 失败
func (p *AsrProber) probeReady(ctx context.Context, base string) AsrProbeResult {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/readyz", nil)
	if err != nil {
		return AsrProbeResult{Err: err}
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return AsrProbeResult{Err: err}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<16))

	if resp.StatusCode == http.StatusOK {
		return AsrProbeResult{Ready: true}
	}

	var parsed asrReadyzResponse
	_ = json.Unmarshal(body, &parsed)

	if resp.StatusCode == http.StatusServiceUnavailable && onlySessionsFull(parsed.Reasons) {
		// 忙：不计失败
		return AsrProbeResult{Ready: true, Busy: true, Reasons: parsed.Reasons}
	}

	reasons := parsed.Reasons
	detail := strings.Join(reasons, ",")
	if detail == "" {
		detail = strings.TrimSpace(string(body))
	}
	return AsrProbeResult{
		Reasons: reasons,
		Err:     fmt.Errorf("/readyz 返回 %d: %s", resp.StatusCode, truncate(detail, 200)),
	}
}

func (p *AsrProber) fetchCapabilities(ctx context.Context, base, apiKey string) (*asrCapabilitiesResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/asr/capabilities", nil)
	if err != nil {
		return nil, err
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("/asr/capabilities 返回 %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var out asrCapabilitiesResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("解析 /asr/capabilities 响应失败: %v", err)
	}
	return &out, nil
}

// BuildProbeUpdates 把探测结果折算成 DB 更新字段。
// 连续失败次数未达阈值时保留原 status，避免转写高峰期的探测超时误判 down
// （OVS 推理同步阻塞 event loop，见设计文档 §1.5）。
func BuildProbeUpdates(server *model.AsrServer, result AsrProbeResult, failThreshold int) map[string]interface{} {
	if failThreshold <= 0 {
		failThreshold = 5
	}

	updates := map[string]interface{}{
		"last_probe_at": time.Now().UnixMilli(),
	}

	if result.Ready {
		if result.Busy {
			// 忙：限流器占满（并发上限 rk/orin-nano = 1），视为健康
			updates["status"] = model.AsrServerStatusBusy
		} else {
			updates["status"] = model.AsrServerStatusUp
		}
		updates["fail_count"] = 0
		if result.Err != nil {
			updates["last_error"] = truncate(result.Err.Error(), 500)
		} else {
			updates["last_error"] = ""
		}
		if result.Backend != "" {
			updates["backend"] = result.Backend
		}
		if len(result.Capabilities) > 0 {
			updates["capabilities"] = strings.Join(result.Capabilities, ",")
		}
		if result.SampleRate > 0 {
			updates["sample_rate"] = result.SampleRate
		}
		return updates
	}

	fails := server.FailCount + 1
	updates["fail_count"] = fails
	if result.Err != nil {
		updates["last_error"] = truncate(result.Err.Error(), 500)
	}
	if fails >= failThreshold {
		updates["status"] = model.AsrServerStatusDown
	}
	return updates
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
