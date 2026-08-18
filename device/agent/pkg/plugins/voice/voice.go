package voice

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
	"unsafe"

	appcfg "github.com/YOUR-ORG/sensecraft-voice/device/agent/cmd/app/config"
	"github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/util/device"
	logutil "github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/util/log"
	"github.com/gen2brain/malgo"
)

// Manager 负责麦克风录音的生命周期管理
type Manager struct {
	mu          sync.RWMutex
	sinks       []sink
	running     bool
	softMuted   bool
	currentOpts appcfg.VoiceOptions

	ctx *malgo.AllocatedContext
	dev *malgo.Device

	// 扇出 sinks
	// 添加 RemoteSink 字段
	remoteSink *RemoteSink

	// 添加 ASRCacheSink 字段
	asrCacheSink *ASRCacheSink

	// 手动控制标志位
	manualStop bool // 用户手动停止标志，为true时不自动启动录音
}

var (
	mgr     *Manager
	mgrOnce sync.Once
)

func GetManager() *Manager {
	mgrOnce.Do(func() {
		mgr = &Manager{}
	})
	return mgr
}

// QuickRecord 录制指定秒数的 WAV 到目录，返回最终文件路径
func (m *Manager) QuickRecord(ctx context.Context, seconds int, sampleRate int, channels int, deviceId string, dir string) (string, error) {
	if seconds <= 0 {
		seconds = 10
	}
	m.mu.Lock()
	needStart := !m.running
	m.mu.Unlock()

	if needStart {
		opts := appcfg.VoiceOptions{DeviceID: deviceId, SampleRate: sampleRate, Channels: channels, Format: "wav", Output: "file", FilePath: dir}
		if err := m.startDevice(ctx, &opts); err != nil {
			return "", err
		}
	}

	// 创建 clip sink 并挂载
	s, err := newClipFileSink(dir, seconds, m.currentOpts.SampleRate, m.currentOpts.Channels)
	if err != nil {
		if needStart {
			_ = m.stopDevice(ctx)
		}
		return "", err
	}
	m.mu.Lock()
	m.sinks = append(m.sinks, s)
	m.mu.Unlock()

	// 轮询等待 clip 完成
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		if ctx.Err() != nil {
			break
		}
		if s.IsDone() {
			break
		}
		<-ticker.C
	}

	path := s.Path()
	// 从列表移除
	m.mu.Lock()
	for i, sk := range m.sinks {
		if sk == s {
			m.sinks = append(m.sinks[:i], m.sinks[i+1:]...)
			break
		}
	}
	m.mu.Unlock()

	if needStart {
		_ = m.stopDevice(ctx)
	}
	return path, nil
}

// StartContinuous 根据配置添加持续文件 sink；如设备未运行则先启动
func (m *Manager) StartContinuous(ctx context.Context, opts appcfg.VoiceOptions, remoteCfg *appcfg.RemoteAudioStreamOptions, baseURL string) error {
	if err := m.ensureDefaults(&opts); err != nil {
		return err
	}

	// 清除手动停止标志
	m.SetManualStop(false)

	m.mu.Lock()
	running := m.running
	m.mu.Unlock()
	if !running {
		if err := m.startDevice(ctx, &opts); err != nil {
			return err
		}
	}
	// 根据 output 添加 sink
	out := strings.ToLower(opts.Output)
	if out == "file" || out == "both" {
		format := audioFormatWAV
		if strings.ToLower(opts.Format) == "pcm16" {
			format = audioFormatPCM16
		}
		s, err := newContinuousFileSink(opts.FilePath, format, opts.SampleRate, opts.Channels)
		if err != nil {
			return err
		}
		m.mu.Lock()
		m.sinks = append(m.sinks, s)
		m.mu.Unlock()
	}
	if out == "stream" || out == "both" {
		ws, err := newWSSink(opts)
		if err != nil {
			return err
		}
		m.mu.Lock()
		m.sinks = append(m.sinks, ws)
		m.mu.Unlock()
	}

	// 启动远程音频流（如果启用）
	if remoteCfg != nil && remoteCfg.Enabled {
		if err := m.StartRemoteStream(ctx, *remoteCfg, baseURL); err != nil {
			logutil.Warnf("failed to start remote stream: %v", err)
		}
	}

	// 启动ASR缓存（如果启用）
	if opts.ASRCache.Enabled {
		if err := m.StartASRCache(ctx, opts.ASRCache, baseURL); err != nil {
			logutil.Warnf("failed to start ASR cache: %v", err)
		}
	} else if remoteCfg != nil && remoteCfg.Enabled {
		// 如果没有启用ASR缓存但启用了远程流，则直接订阅ASR结果
		go m.subscribeASRResults(m.remoteSink)
	}

	return nil
}

// convertASRCacheConfig 转换配置
func convertASRCacheConfig(config appcfg.ASRCacheOptions) ASRCacheConfig {
	return ASRCacheConfig{
		Enabled:         config.Enabled,
		CacheDir:        config.CacheDir,
		MaxRetries:      config.MaxRetries,
		RetryInterval:   config.RetryInterval,
		CacheExpiry:     config.CacheExpiry,
		MaxCacheSize:    config.MaxCacheSize,
		CleanupInterval: config.CleanupInterval,
		HTTPBatch: HTTPBatchConfig{
			Enabled:          config.HTTPBatch.Enabled,
			BatchSize:        config.HTTPBatch.BatchSize,
			UploadInterval:   config.HTTPBatch.UploadInterval,
			MaxRetryAttempts: config.HTTPBatch.MaxRetryAttempts,
			Timeout:          config.HTTPBatch.Timeout,
		},
	}
}

// StartASRCache 启动ASR缓存
func (m *Manager) StartASRCache(ctx context.Context, config appcfg.ASRCacheOptions, baseURL string) error {
	if !config.Enabled {
		return nil
	}

	// 转换配置
	cacheConfig := convertASRCacheConfig(config)

	// 获取MAC地址
	macAddress := ""
	if m.remoteSink != nil {
		macAddress = m.remoteSink.GetMacAddress()
	}

	// 创建ASR缓存Sink
	asrCacheSink := NewASRCacheSink(cacheConfig, macAddress, m.remoteSink)

	// 设置RemoteSink引用
	asrCacheSink.SetRemoteSink(m.remoteSink)

	// 启动ASR结果订阅，将结果转发给asr_cache_sink
	go m.subscribeASRResultsForCache(asrCacheSink)

	m.mu.Lock()
	m.asrCacheSink = asrCacheSink
	m.mu.Unlock()

	logutil.Infof("ASR cache started, cache_dir=%s, mac_address=%s", cacheConfig.CacheDir, macAddress)
	return nil
}

// subscribeASRResultsForCache 订阅ASR结果并转发给asr_cache_sink
func (m *Manager) subscribeASRResultsForCache(asrCacheSink *ASRCacheSink) {
	logutil.Infof("ASR cache: ASR result subscription started")

	// 获取 ASR Hub 实例
	hub := GetASRHub()

	// 订阅 ASR 结果，缓冲区大小为 32
	subCh, unsubscribe := hub.Subscribe(32)
	defer unsubscribe()

	// 监听 ASR 结果
	for data := range subCh {
		// 解析 ASR 结果
		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			logutil.Warnf("ASR cache: failed to parse ASR result: %v", err)
			continue
		}

		// 只处理 final 类型的 ASR 结果
		if msgType, ok := result["type"].(string); ok && msgType == "final" {
			logutil.Debugf("ASR cache: received ASR result: %s", result["text"])

			// 推送给 asr_cache_sink
			asrCacheSink.OnASRResult(result)
		}
	}

	logutil.Infof("ASR cache: ASR result subscription ended")
}

// StartRemoteStream 启动远程音频流
func (m *Manager) StartRemoteStream(ctx context.Context, remoteCfg appcfg.RemoteAudioStreamOptions, baseURL string) error {
	if !remoteCfg.Enabled {
		return nil
	}

	// 获取 MAC 地址
	macAddress := remoteCfg.MacAddress
	if macAddress == "" {
		// 自动获取 MAC 地址
		var err error
		macAddress, err = m.getMACAddress()
		if err != nil {
			logutil.Warnf("failed to get MAC address: %v, using empty", err)
		}
	}

	// 创建远程 sink
	cfg := RemoteSinkConfig{
		BaseURL:          baseURL,
		MacAddress:       macAddress,
		Headers:          remoteCfg.Headers,
		MaxQueue:         remoteCfg.MaxQueue,
		MaxReconnectWait: remoteCfg.MaxReconnectDelay,
	}

	remoteSink, err := newRemoteSink(cfg)
	if err != nil {
		return fmt.Errorf("failed to create remote sink: %w", err)
	}

	// 存储 RemoteSink 到 Manager 中
	m.mu.Lock()
	m.remoteSink = remoteSink
	m.mu.Unlock()

	logutil.Infof("remote sink started, connecting to %s with MAC: %s", baseURL, macAddress)
	return nil
}

// UpdateRemoteSink 更新远程sink配置
func (m *Manager) UpdateRemoteSink(baseURL string, remoteCfg *appcfg.RemoteAudioStreamOptions) error {
	if remoteCfg == nil || !remoteCfg.Enabled {
		return fmt.Errorf("remote stream is not enabled")
	}

	// 验证baseURL
	if baseURL == "" {
		return fmt.Errorf("base URL cannot be empty")
	}

	logutil.Infof("updating remote sink to %s", baseURL)

	// 停止当前的remote sink
	m.mu.Lock()
	oldRemoteSink := m.remoteSink
	asrCacheSink := m.asrCacheSink
	m.remoteSink = nil
	m.mu.Unlock()

	// 关闭旧的remote sink
	if oldRemoteSink != nil {
		oldRemoteSink.Close()
		logutil.Infof("old remote sink closed")
	}

	// 创建新的remote sink
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := m.StartRemoteStream(ctx, *remoteCfg, baseURL); err != nil {
		// 如果创建失败，尝试恢复旧的sink
		if oldRemoteSink != nil {
			logutil.Warnf("failed to create new remote sink, attempting to restore old one: %v", err)
			m.mu.Lock()
			m.remoteSink = oldRemoteSink
			m.mu.Unlock()
		}
		return fmt.Errorf("failed to start new remote sink: %w", err)
	}

	// 更新ASR cache sink的remote sink引用（如果存在）
	m.mu.RLock()
	newRemoteSink := m.remoteSink
	m.mu.RUnlock()

	if asrCacheSink != nil && newRemoteSink != nil {
		// 如果ASR cache sink存在，更新其remote sink引用
		asrCacheSink.SetRemoteSink(newRemoteSink)
		logutil.Infof("ASR cache sink remote sink reference updated")
	} else if asrCacheSink == nil && newRemoteSink != nil {
		// 如果ASR cache sink不存在，直接订阅ASR结果给remote sink
		go m.subscribeASRResults(newRemoteSink)
		logutil.Infof("ASR results subscription restarted for new remote sink")
	}

	logutil.Infof("remote sink successfully updated to %s", baseURL)
	return nil
}

// subscribeASRResults 订阅 ASR 结果并转发给 remote_sink
func (m *Manager) subscribeASRResults(remoteSink *RemoteSink) {
	logutil.Infof("remote sink: ASR result subscription started")

	// 获取 ASR Hub 实例
	hub := GetASRHub()

	// 订阅 ASR 结果，缓冲区大小为 32
	subCh, unsubscribe := hub.Subscribe(32)
	defer unsubscribe()

	// 监听 ASR 结果
	for data := range subCh {
		// 解析 ASR 结果
		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			logutil.Warnf("remote sink: failed to parse ASR result: %v", err)
			continue
		}

		// 只处理 final 类型的 ASR 结果
		if msgType, ok := result["type"].(string); ok && msgType == "final" {
			logutil.Infof("remote sink: received ASR result: %s", result["text"])

			// 推送给 remote_sink
			remoteSink.OnASRResult(result)
		}
	}

	logutil.Infof("remote sink: ASR result subscription ended")
}

// StopContinuous 关闭所有持续 sink（保留设备可选，这里直接关闭设备）
func (m *Manager) StopContinuous(ctx context.Context) error {
	return m.StopContinuousWithReason(ctx, false) // 默认为系统停止
}

// StopContinuousWithReason 关闭所有持续 sink，并指定停止原因
func (m *Manager) StopContinuousWithReason(ctx context.Context, isManualStop bool) error {
	m.mu.Lock()
	sinks := m.sinks
	asrCacheSink := m.asrCacheSink
	m.sinks = nil
	m.asrCacheSink = nil
	m.mu.Unlock()

	// 关闭所有sinks
	for _, s := range sinks {
		s.Close()
	}

	// 关闭ASR缓存Sink
	if asrCacheSink != nil {
		asrCacheSink.Close()
	}

	// 设置手动停止标志
	if isManualStop {
		m.SetManualStop(true)
	}

	return m.stopDevice(ctx)
}

// 内部：启动设备并设置回调分发
func (m *Manager) startDevice(ctx context.Context, opts *appcfg.VoiceOptions) error {
	if err := m.ensureDefaults(opts); err != nil {
		return err
	}
	context, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {})
	if err != nil {
		return err
	}
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = uint32(opts.Channels)
	deviceConfig.SampleRate = uint32(opts.SampleRate)
	// Enumerate capture devices once and keep `infos` alive until InitDevice
	// has consumed the chosen DeviceID pointer below — &infos[idx].ID is only
	// valid while this slice is still referenced.
	//
	// The DeviceID is a Go-heap pointer that gets embedded in deviceConfig and
	// handed to C inside malgo.InitDevice (cgo). Under Go's pointer-passing
	// rules a Go pointer reachable from a value passed to C must be *pinned*,
	// otherwise the runtime panics with "cgo argument has Go pointer to
	// unpinned Go pointer". Pin it for the duration of InitDevice; once
	// InitDevice has copied the ID into its own (C-side) device handle the
	// pointer is no longer needed, so unpinning at function return is safe.
	var pinner runtime.Pinner
	defer pinner.Unpin()
	infos, _ := context.Devices(malgo.Capture)
	if idx := pickInputDevice(infos, opts.DeviceID); idx >= 0 {
		pinner.Pin(&infos[idx].ID)
		deviceConfig.Capture.DeviceID = unsafe.Pointer(&infos[idx].ID)
		logutil.Infof("voice: using input device: %s", infos[idx].Name())
	} else {
		logutil.Infof("voice: using input device: %s", "<system-default>")
	}
	d, err := malgo.InitDevice(context.Context, deviceConfig, malgo.DeviceCallbacks{
		Data: func(pOutputSample, pInputSample []byte, framecount uint32) {
			m.mu.Lock()
			sinks := append([]sink(nil), m.sinks...)
			softMuted := m.softMuted
			m.mu.Unlock()
			if softMuted || len(pInputSample) == 0 {
				return
			}
			for _, s := range sinks {
				s.OnData(pInputSample)
			}
		},
	})
	if err != nil {
		context.Uninit()
		context.Free()
		return err
	}
	if err = d.Start(); err != nil {
		d.Uninit()
		context.Uninit()
		context.Free()
		return err
	}
	m.mu.Lock()
	m.ctx = context
	m.dev = d
	m.currentOpts = *opts
	m.running = true
	m.mu.Unlock()
	return nil
}

func (m *Manager) stopDevice(ctx context.Context) error {
	m.mu.Lock()
	dev := m.dev
	context := m.ctx
	m.dev = nil
	m.ctx = nil
	m.running = false
	m.mu.Unlock()
	if dev != nil {
		_ = dev.Stop()
		dev.Uninit()
	}
	if context != nil {
		context.Uninit()
		context.Free()
	}
	return nil
}

func (m *Manager) ensureDefaults(opts *appcfg.VoiceOptions) error {
	if opts.SampleRate == 0 {
		opts.SampleRate = 16000
	}
	if opts.Channels == 0 {
		opts.Channels = 1
	}
	if opts.Format == "" {
		if strings.HasSuffix(strings.ToLower(opts.FilePath), ".pcm") {
			opts.Format = "pcm16"
		} else {
			opts.Format = "wav"
		}
	}
	if opts.Output == "" {
		opts.Output = "file"
	}
	if opts.Output == "file" || opts.Output == "both" {
		dir := opts.FilePath
		if strings.TrimSpace(dir) == "" {
			dir = "./recordings/voice"
		}
		dir = filepath.Clean(dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create output dir failed: %w", err)
		}
		opts.FilePath = dir
	}
	return nil
}

// pickInputDevice returns the index of the capture device to use, or -1 when
// none are available. It prefers an explicit config match (exact name, then
// case-insensitive substring), then falls back to a real microphone
// (reSpeaker / XVF / mic-like) and finally the first enumerated device.
//
// The key behaviour change: never silently fall back to the ALSA "default"
// PCM. Inside the container there is no usable default slave, so opening it
// fails with miniaudio "failed to open backend device". Any real enumerated
// capture device (e.g. the reSpeaker on a non-zero card index) beats it.
func pickInputDevice(infos []malgo.DeviceInfo, deviceId string) int {
	if len(infos) == 0 {
		return -1
	}
	deviceId = strings.TrimSpace(deviceId)
	if deviceId != "" {
		for i := range infos {
			if infos[i].Name() == deviceId {
				return i
			}
		}
		lower := strings.ToLower(deviceId)
		for i := range infos {
			if strings.Contains(strings.ToLower(infos[i].Name()), lower) {
				return i
			}
		}
	}
	for _, kw := range []string{"respeaker", "xvf", "array", "mic"} {
		for i := range infos {
			if strings.Contains(strings.ToLower(infos[i].Name()), kw) {
				return i
			}
		}
	}
	return 0
}

func (m *Manager) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

// IsManualStop 检查是否处于手动停止状态
func (m *Manager) IsManualStop() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.manualStop
}

// SetManualStop 设置手动停止状态
func (m *Manager) SetManualStop(manualStop bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.manualStop = manualStop
	logutil.Infof("voice: manual stop flag set to %v", manualStop)
}

// getMACAddress 获取 MAC 地址
func (m *Manager) getMACAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		// 跳过回环接口和虚拟接口
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		// 获取第一个有效的物理接口的 MAC 地址
		if len(iface.HardwareAddr) > 0 {
			return iface.HardwareAddr.String(), nil
		}
	}

	return "", fmt.Errorf("未找到有效的网络接口")
}

// GetDeviceStatus 获取设备状态
func (m *Manager) GetDeviceStatus() map[string]interface{} {
	// 导入设备状态工具包
	deviceStatus := device.GetSystemStatus()

	// 添加语音相关状态
	status := make(map[string]interface{})
	for k, v := range deviceStatus {
		status[k] = v
	}

	// 添加语音管理器状态
	m.mu.RLock()
	status["voice_running"] = m.running
	status["voice_soft_muted"] = m.softMuted
	status["voice_sinks_count"] = len(m.sinks)
	if m.remoteSink != nil {
		status["remote_sink_connected"] = true
	} else {
		status["remote_sink_connected"] = false
	}
	m.mu.RUnlock()

	// 添加远程服务配置信息
	// 这里我们需要从配置中获取remote.base_url
	// 由于voice manager没有直接访问配置的权限，我们暂时设置为空
	// 实际实现中，这个值应该从配置中获取
	status["remote_base_url"] = "" // 将在路由层设置

	return status
}

// GetDeviceInfo 获取详细设备信息
func (m *Manager) GetDeviceInfo() (*device.DeviceInfo, error) {
	return device.GetDeviceInfo()
}

// GetASRCacheSink 获取ASR缓存Sink
func (m *Manager) GetASRCacheSink() *ASRCacheSink {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.asrCacheSink
}

// GetASRCacheStatus 获取ASR缓存状态
func (m *Manager) GetASRCacheStatus() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.asrCacheSink == nil {
		return map[string]interface{}{
			"enabled": false,
			"message": "ASR cache not initialized",
		}
	}

	return m.asrCacheSink.GetStatus()
}

// GetASRCacheMetrics 获取ASR缓存指标
func (m *Manager) GetASRCacheMetrics() ASRCacheMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.asrCacheSink == nil {
		return ASRCacheMetrics{}
	}

	return m.asrCacheSink.GetMetrics()
}
