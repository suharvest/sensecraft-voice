package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
)

// readyzServer 构造一个只实现 /readyz 与 /asr/capabilities 的假 OVS
func readyzServer(t *testing.T, code int, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/readyz":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(code)
			_, _ = w.Write([]byte(body))
		case "/asr/capabilities":
			if r.Header.Get("Authorization") != "Bearer k" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"backend":"fake","capabilities":["offline","multi_language"],"sample_rate":16000}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestProbeReadyOK(t *testing.T) {
	srv := readyzServer(t, http.StatusOK, `{"status":"ready"}`)
	defer srv.Close()

	res := NewAsrProber(2 * time.Second).Probe(context.Background(), srv.URL, "k")
	if !res.Ready || res.Busy {
		t.Fatalf("expected ready & not busy, got %+v", res)
	}
	if res.Backend != "fake" || res.SampleRate != 16000 || len(res.Capabilities) != 2 {
		t.Fatalf("capabilities not filled: %+v", res)
	}
}

// /readyz 503 且 reasons 只含 sessions_full：并发上限 1 的设备解码中，属「忙」不属「坏」
func TestProbeBusySessionsFull(t *testing.T) {
	srv := readyzServer(t, http.StatusServiceUnavailable, `{"status":"not_ready","reasons":["sessions_full"]}`)
	defer srv.Close()

	res := NewAsrProber(2 * time.Second).Probe(context.Background(), srv.URL, "k")
	if !res.Ready || !res.Busy {
		t.Fatalf("expected ready & busy, got %+v", res)
	}

	updates := BuildProbeUpdates(&model.AsrServer{FailCount: 3}, res, 5)
	if updates["status"] != model.AsrServerStatusBusy {
		t.Fatalf("expected status busy, got %v", updates["status"])
	}
	if updates["fail_count"] != 0 {
		t.Fatalf("busy must reset fail_count, got %v", updates["fail_count"])
	}
}

func TestProbeFailOnBackendNotReady(t *testing.T) {
	srv := readyzServer(t, http.StatusServiceUnavailable, `{"status":"not_ready","reasons":["backend_not_ready"]}`)
	defer srv.Close()

	res := NewAsrProber(2 * time.Second).Probe(context.Background(), srv.URL, "k")
	if res.Ready || res.Err == nil {
		t.Fatalf("expected failure, got %+v", res)
	}

	// 连续失败未达阈值时不改 status
	updates := BuildProbeUpdates(&model.AsrServer{FailCount: 1}, res, 5)
	if _, ok := updates["status"]; ok {
		t.Fatalf("status must stay untouched below threshold: %v", updates)
	}
	if updates["fail_count"] != 2 {
		t.Fatalf("expected fail_count 2, got %v", updates["fail_count"])
	}

	// 达到阈值才置 down
	updates = BuildProbeUpdates(&model.AsrServer{FailCount: 4}, res, 5)
	if updates["status"] != model.AsrServerStatusDown {
		t.Fatalf("expected down at threshold, got %v", updates["status"])
	}
}

// sessions_full 与其他原因并存时必须计失败
func TestProbeFailOnMixedReasons(t *testing.T) {
	srv := readyzServer(t, http.StatusServiceUnavailable,
		`{"status":"not_ready","reasons":["sessions_full","gpu_watchdog_failed"]}`)
	defer srv.Close()

	res := NewAsrProber(2 * time.Second).Probe(context.Background(), srv.URL, "k")
	if res.Ready || res.Busy {
		t.Fatalf("mixed reasons must not count as busy: %+v", res)
	}
}

// capabilities 需要 api key；拿不到不影响可用性判定
func TestProbeReadyWithCapabilitiesAuthFailure(t *testing.T) {
	srv := readyzServer(t, http.StatusOK, `{"status":"ready"}`)
	defer srv.Close()

	res := NewAsrProber(2 * time.Second).Probe(context.Background(), srv.URL, "wrong-key")
	if !res.Ready {
		t.Fatalf("capabilities auth failure must not mark server unavailable: %+v", res)
	}
	if res.Err == nil {
		t.Fatal("expected capabilities error to be surfaced")
	}
	updates := BuildProbeUpdates(&model.AsrServer{}, res, 5)
	if updates["status"] != model.AsrServerStatusUp {
		t.Fatalf("expected up, got %v", updates["status"])
	}
}

func TestProbeConnectionRefused(t *testing.T) {
	res := NewAsrProber(time.Second).Probe(context.Background(), "http://127.0.0.1:1", "")
	if res.Ready || res.Err == nil {
		t.Fatalf("expected connection failure, got %+v", res)
	}
}
