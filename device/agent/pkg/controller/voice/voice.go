package voice

import (
	"context"

	appcfg "github.com/YOUR-ORG/sensecraft-voice/device/agent/cmd/app/config"
	pvoice "github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/plugins/voice"
)

type Interface interface {
	StartByConfig(ctx context.Context) error
	Stop(ctx context.Context) error
	StopWithReason(ctx context.Context, isManualStop bool) error
	QuickRecord(ctx context.Context, seconds int, sampleRate int, channels int, deviceId string, dir string) (string, error)
	StartWithOverride(ctx context.Context, override appcfg.VoiceOptions) error
	IsRunning() bool
	IsManualStop() bool
	ResetManualStop() error
	UpdateRemoteConfig(baseURL string) error
}

type Getter interface {
	Voice() Interface
}

type svc struct {
	cfg appcfg.Config
}

func New(cfg appcfg.Config) Interface { return &svc{cfg: cfg} }

func (s *svc) StartByConfig(ctx context.Context) error {
	return pvoice.GetManager().StartContinuous(ctx, s.cfg.Voice, &s.cfg.Remote.AudioStream, s.cfg.Remote.BaseURL)
}
func (s *svc) Stop(ctx context.Context) error { return pvoice.GetManager().StopContinuous(ctx) }

func (s *svc) QuickRecord(ctx context.Context, seconds int, sampleRate int, channels int, deviceId string, dir string) (string, error) {
	if sampleRate == 0 {
		sampleRate = s.cfg.Voice.SampleRate
	}
	if channels == 0 {
		channels = s.cfg.Voice.Channels
	}
	// if dir == "" {
	// 	dir = "./recordings/voice"
	// }
	dir = "./recordings/voice/test"
	return pvoice.GetManager().QuickRecord(ctx, seconds, sampleRate, channels, deviceId, dir)
}

func (s *svc) StartWithOverride(ctx context.Context, override appcfg.VoiceOptions) error {
	base := s.cfg.Voice
	if override.DeviceID != "" {
		base.DeviceID = override.DeviceID
	}
	if override.SampleRate > 0 {
		base.SampleRate = override.SampleRate
	}
	if override.Channels > 0 {
		base.Channels = override.Channels
	}
	if override.Format != "" {
		base.Format = override.Format
	}
	if override.Output != "" {
		base.Output = override.Output
	}
	if override.FilePath != "" {
		base.FilePath = override.FilePath
	}
	if override.OnDeviceLost != "" {
		base.OnDeviceLost = override.OnDeviceLost
	}
	if override.SegmentSeconds > 0 {
		base.SegmentSeconds = override.SegmentSeconds
	}
	if override.WSUrl != "" {
		base.WSUrl = override.WSUrl
	}
	if len(override.WSHeaders) > 0 {
		base.WSHeaders = override.WSHeaders
	}
	if override.WSChunkBytes > 0 {
		base.WSChunkBytes = override.WSChunkBytes
	}
	if override.WSMaxQueue > 0 {
		base.WSMaxQueue = override.WSMaxQueue
	}
	if override.WSMaxReconnectDelay > 0 {
		base.WSMaxReconnectDelay = override.WSMaxReconnectDelay
	}
	return pvoice.GetManager().StartContinuous(ctx, base, &s.cfg.Remote.AudioStream, s.cfg.Remote.BaseURL)
}

func (s *svc) IsRunning() bool {
	return pvoice.GetManager().IsRunning()
}

func (s *svc) StopWithReason(ctx context.Context, isManualStop bool) error {
	return pvoice.GetManager().StopContinuousWithReason(ctx, isManualStop)
}

func (s *svc) IsManualStop() bool {
	return pvoice.GetManager().IsManualStop()
}

func (s *svc) ResetManualStop() error {
	pvoice.GetManager().SetManualStop(false)
	return nil
}

// UpdateRemoteConfig 更新远程配置
func (s *svc) UpdateRemoteConfig(baseURL string) error {
	// 更新配置
	s.cfg.Remote.BaseURL = baseURL

	// 通知voice manager更新remote sink
	return pvoice.GetManager().UpdateRemoteSink(baseURL, &s.cfg.Remote.AudioStream)
}
