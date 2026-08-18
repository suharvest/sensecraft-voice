package jobmanager

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/util/log"
)

// VoiceController 语音控制器接口
type VoiceController interface {
	IsRunning() bool
	IsManualStop() bool
	StartByConfig(ctx context.Context) error
	Stop(ctx context.Context) error
}

// MicrophoneMonitorJob 麦克风监控任务
type MicrophoneMonitorJob struct {
	mu                   sync.RWMutex
	asrCacheEnabled      bool
	deviceID             string
	voiceController      VoiceController
	lastDeviceState      bool
	lastCheckTime        time.Time
	stateStableCount     int   // 状态稳定计数器
	stateStableThreshold int   // 状态稳定阈值
	consecutiveErrors    int   // 连续错误计数
	maxConsecutiveErrors int   // 最大连续错误数
	lastError            error // 最后一次错误

	// 重启重试相关字段
	autoStart            bool          // 是否自动启动
	restartRetryCount    int           // 当前重启重试次数
	maxRestartRetries    int           // 最大重启重试次数
	restartRetryInterval time.Duration // 重启重试间隔
	lastRestartTime      time.Time     // 最后一次重启时间
	isRestarting         bool          // 是否正在重启过程中
	forceRestart         bool          // 是否强制重启（用于重启时的强制stop/start）
}

// NewMicrophoneMonitorJob 创建麦克风监控任务
func NewMicrophoneMonitorJob(asrCacheEnabled bool, deviceID string, voiceController VoiceController) *MicrophoneMonitorJob {
	return &MicrophoneMonitorJob{
		asrCacheEnabled:      asrCacheEnabled,
		deviceID:             deviceID,
		voiceController:      voiceController,
		stateStableThreshold: 2,               // 需要连续2次检查状态一致才执行操作
		maxConsecutiveErrors: 5,               // 最大连续错误数
		autoStart:            true,            // 默认启用自动启动
		maxRestartRetries:    3,               // 最大重启重试次数
		restartRetryInterval: 5 * time.Second, // 重启重试间隔5秒
		forceRestart:         true,            // 启动时默认需要强制重启
	}
}

// NewMicrophoneMonitorJobWithAutoStart 创建带自动启动配置的麦克风监控任务
func NewMicrophoneMonitorJobWithAutoStart(asrCacheEnabled bool, deviceID string, voiceController VoiceController, autoStart bool) *MicrophoneMonitorJob {
	return &MicrophoneMonitorJob{
		asrCacheEnabled:      asrCacheEnabled,
		deviceID:             deviceID,
		voiceController:      voiceController,
		stateStableThreshold: 2, // 需要连续2次检查状态一致才执行操作
		maxConsecutiveErrors: 5, // 最大连续错误数
		autoStart:            autoStart,
		maxRestartRetries:    3,               // 最大重启重试次数
		restartRetryInterval: 5 * time.Second, // 重启重试间隔5秒
		forceRestart:         true,            // 启动时默认需要强制重启
	}
}

// Name 返回任务名称
func (j *MicrophoneMonitorJob) Name() string {
	return "microphone-monitor"
}

// CronSpec 返回cron表达式，每30秒执行一次
func (j *MicrophoneMonitorJob) CronSpec() string {
	return "@every 30s" // 每30秒执行一次
}

// Do 执行监控任务
func (j *MicrophoneMonitorJob) Do(ctx *JobContext) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	// 如果ASR缓存未启用，跳过监控
	if !j.asrCacheEnabled {
		ctx.WithLogFields(map[string]interface{}{
			"message": "ASR cache disabled, skipping microphone monitoring",
		})
		return nil
	}

	// 检查是否在重启重试过程中
	if j.isRestarting {
		return j.handleRestartRetry(ctx)
	}

	// 如果连续错误过多，暂时跳过检查
	if j.consecutiveErrors >= j.maxConsecutiveErrors {
		ctx.WithLogFields(map[string]interface{}{
			"message":                "too many consecutive errors, skipping check",
			"consecutive_errors":     j.consecutiveErrors,
			"max_consecutive_errors": j.maxConsecutiveErrors,
			"last_error":             j.lastError.Error(),
		})
		return nil
	}

	// 检查设备状态
	deviceAvailable, err := j.isMicrophoneAvailable()
	if err != nil {
		j.consecutiveErrors++
		j.lastError = err
		ctx.WithLogFields(map[string]interface{}{
			"error":              err.Error(),
			"message":            "failed to check device status",
			"consecutive_errors": j.consecutiveErrors,
		})
		return nil // 不返回错误，避免影响其他任务
	}

	// 重置错误计数
	j.consecutiveErrors = 0
	j.lastError = nil

	// 更新状态稳定计数器
	if deviceAvailable == j.lastDeviceState {
		j.stateStableCount++
	} else {
		j.stateStableCount = 1
		j.lastDeviceState = deviceAvailable
	}

	j.lastCheckTime = time.Now()

	// 记录当前状态
	ctx.WithLogFields(map[string]interface{}{
		"device_available":       deviceAvailable,
		"last_device_state":      j.lastDeviceState,
		"state_stable_count":     j.stateStableCount,
		"state_stable_threshold": j.stateStableThreshold,
		"consecutive_errors":     j.consecutiveErrors,
	})

	// 只有状态稳定时才执行操作
	if j.stateStableCount >= j.stateStableThreshold {
		if err := j.syncRecordingState(ctx, deviceAvailable); err != nil {
			j.consecutiveErrors++
			j.lastError = err
			ctx.WithLogFields(map[string]interface{}{
				"error":              err.Error(),
				"message":            "failed to sync recording state",
				"consecutive_errors": j.consecutiveErrors,
			})
			return nil // 不返回错误，避免影响其他任务
		}
	}

	return nil
}

// isMicrophoneAvailable 检查麦克风是否可用（使用系统命令）
func (j *MicrophoneMonitorJob) isMicrophoneAvailable() (bool, error) {
	var cmd *exec.Cmd
	var devicePattern string

	switch runtime.GOOS {
	case "linux":
		// Linux 使用 arecord -l
		cmd = exec.Command("arecord", "-l")
		devicePattern = "card"
	case "darwin":
		// macOS 使用 system_profiler
		cmd = exec.Command("system_profiler", "SPAudioDataType")
		devicePattern = "Input"
	case "windows":
		// Windows 使用 PowerShell
		cmd = exec.Command("powershell", "-Command", "Get-WmiObject -Class Win32_SoundDevice | Where-Object {$_.Status -eq 'OK'}")
		devicePattern = "Name"
	default:
		log.Warnf("Unsupported platform for microphone detection: %s", runtime.GOOS)
		return false, nil
	}

	// 执行命令
	output, err := cmd.Output()
	if err != nil {
		log.Debugf("Audio device check command failed: %v", err)
		return false, err
	}

	outputStr := string(output)
	log.Debugf("Audio device check output: %s", outputStr)

	// 检查是否有音频设备
	if !strings.Contains(outputStr, devicePattern) {
		return false, nil
	}

	// 如果指定了设备ID，检查特定设备
	if j.deviceID != "" {
		return j.checkSpecificDevice(outputStr), nil
	}

	// 没有指定设备ID，只要有任何音频设备就返回true
	return true, nil
}

// checkSpecificDevice 检查特定设备是否存在
func (j *MicrophoneMonitorJob) checkSpecificDevice(output string) bool {
	lines := strings.Split(output, "\n")

	switch runtime.GOOS {
	case "linux":
		// Linux arecord -l 输出格式：
		// card 0: PCH [HDA Intel PCH], device 0: ALC892 Analog [ALC892 Analog]
		for _, line := range lines {
			if strings.Contains(line, "card") && strings.Contains(line, j.deviceID) {
				return true
			}
		}
	case "darwin":
		// macOS system_profiler 输出格式：
		//         ReSpeaker 4 Mic Array (UAC1.0):
		//           Default Input Device: Yes
		for _, line := range lines {
			// 检查设备名称行（8个空格缩进且包含设备ID）
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(line, "        ") && !strings.HasPrefix(line, "          ") && strings.Contains(trimmed, j.deviceID) {
				return true
			}
		}
	case "windows":
		// Windows PowerShell 输出格式：
		// Name: Realtek High Definition Audio
		for _, line := range lines {
			if strings.Contains(line, "Name:") && strings.Contains(line, j.deviceID) {
				return true
			}
		}
	}

	return false
}

// syncRecordingState 同步录音状态
func (j *MicrophoneMonitorJob) syncRecordingState(ctx *JobContext, deviceAvailable bool) error {
	if j.voiceController == nil {
		ctx.WithLogFields(map[string]interface{}{
			"message": "voice controller not available, skipping state sync",
		})
		return nil
	}

	// 获取当前录音状态和手动停止状态
	isRunning := j.voiceController.IsRunning()
	isManualStop := j.voiceController.IsManualStop()

	// 检查是否需要强制重启
	if j.forceRestart && deviceAvailable && !isManualStop {
		ctx.WithLogFields(map[string]interface{}{
			"action":           "force_restart_recording",
			"reason":           "force restart requested, device available, not manually stopped",
			"device_available": deviceAvailable,
			"is_running":       isRunning,
			"is_manual_stop":   isManualStop,
		})

		// 重置强制重启标志
		j.forceRestart = false

		// 使用重启重试机制
		return j.startRecordingWithRetry(ctx)
	}

	// 根据状态矩阵决定操作
	if deviceAvailable && !isRunning && !isManualStop {
		// 设备可用、录音未运行且未手动停止，启动录音
		ctx.WithLogFields(map[string]interface{}{
			"action":           "start_recording",
			"reason":           "device available, recording not running, not manually stopped",
			"device_available": deviceAvailable,
			"is_running":       isRunning,
			"is_manual_stop":   isManualStop,
		})

		// 如果启用了自动启动，使用重启重试机制
		if j.autoStart {
			return j.startRecordingWithRetry(ctx)
		} else {
			return j.voiceController.StartByConfig(context.Background())
		}
	} else if !deviceAvailable && isRunning {
		// 设备不可用但录音在运行，停止录音（系统停止，不设置手动停止标志）
		ctx.WithLogFields(map[string]interface{}{
			"action":           "stop_recording",
			"reason":           "device not available but recording is running",
			"device_available": deviceAvailable,
			"is_running":       isRunning,
			"is_manual_stop":   isManualStop,
		})
		return j.voiceController.Stop(context.Background())
	}

	// 状态一致，无需操作
	ctx.WithLogFields(map[string]interface{}{
		"message":          "recording state matches device state or manually stopped, no action needed",
		"device_available": deviceAvailable,
		"is_running":       isRunning,
		"is_manual_stop":   isManualStop,
	})

	return nil
}

// GetStatus 获取监控状态
func (j *MicrophoneMonitorJob) GetStatus() map[string]interface{} {
	j.mu.RLock()
	defer j.mu.RUnlock()

	status := map[string]interface{}{
		"enabled":                j.asrCacheEnabled,
		"device_id":              j.deviceID,
		"last_device_state":      j.lastDeviceState,
		"last_check_time":        j.lastCheckTime,
		"state_stable_count":     j.stateStableCount,
		"state_stable_threshold": j.stateStableThreshold,
		"consecutive_errors":     j.consecutiveErrors,
		"max_consecutive_errors": j.maxConsecutiveErrors,
		"platform":               runtime.GOOS,
		"auto_start":             j.autoStart,
		"is_restarting":          j.isRestarting,
		"force_restart":          j.forceRestart,
		"restart_retry_count":    j.restartRetryCount,
		"max_restart_retries":    j.maxRestartRetries,
		"restart_retry_interval": j.restartRetryInterval.String(),
		"last_restart_time":      j.lastRestartTime,
	}

	if j.lastError != nil {
		status["last_error"] = j.lastError.Error()
	}

	return status
}

// startRecordingWithRetry 使用重启重试机制启动录音
func (j *MicrophoneMonitorJob) startRecordingWithRetry(ctx *JobContext) error {
	// 初始化重启重试状态
	j.restartRetryCount = 0
	j.isRestarting = true
	j.lastRestartTime = time.Now()

	ctx.WithLogFields(map[string]interface{}{
		"message":                "starting recording with restart retry mechanism",
		"max_restart_retries":    j.maxRestartRetries,
		"restart_retry_interval": j.restartRetryInterval,
	})

	// 立即执行第一次重启尝试
	return j.performRestartCycle(ctx)
}

// handleRestartRetry 处理重启重试逻辑
func (j *MicrophoneMonitorJob) handleRestartRetry(ctx *JobContext) error {
	// 检查是否超过重试间隔
	if time.Since(j.lastRestartTime) < j.restartRetryInterval {
		ctx.WithLogFields(map[string]interface{}{
			"message":         "waiting for restart retry interval",
			"time_since_last": time.Since(j.lastRestartTime),
			"retry_interval":  j.restartRetryInterval,
		})
		return nil
	}

	// 检查是否超过最大重试次数
	if j.restartRetryCount >= j.maxRestartRetries {
		ctx.WithLogFields(map[string]interface{}{
			"message":             "max restart retries exceeded, stopping retry mechanism",
			"restart_retry_count": j.restartRetryCount,
			"max_restart_retries": j.maxRestartRetries,
		})
		j.isRestarting = false
		return nil
	}

	// 执行重启周期
	return j.performRestartCycle(ctx)
}

// performRestartCycle 执行一次重启周期（停止+启动）
func (j *MicrophoneMonitorJob) performRestartCycle(ctx *JobContext) error {
	j.restartRetryCount++
	j.lastRestartTime = time.Now()

	ctx.WithLogFields(map[string]interface{}{
		"message":             "performing restart cycle",
		"restart_retry_count": j.restartRetryCount,
		"max_restart_retries": j.maxRestartRetries,
	})

	// 先停止录音
	if j.voiceController.IsRunning() {
		ctx.WithLogFields(map[string]interface{}{
			"action": "stopping recording before restart",
		})
		if err := j.voiceController.Stop(context.Background()); err != nil {
			ctx.WithLogFields(map[string]interface{}{
				"error":  err.Error(),
				"action": "failed to stop recording",
			})
		}

		// 等待一小段时间确保停止完成
		time.Sleep(1 * time.Second)
	}

	// 启动录音
	ctx.WithLogFields(map[string]interface{}{
		"action": "starting recording after restart",
	})
	if err := j.voiceController.StartByConfig(context.Background()); err != nil {
		ctx.WithLogFields(map[string]interface{}{
			"error":  err.Error(),
			"action": "failed to start recording",
		})
		return err
	}

	// 检查启动是否成功
	time.Sleep(2 * time.Second) // 等待启动完成
	if j.voiceController.IsRunning() {
		ctx.WithLogFields(map[string]interface{}{
			"message":             "recording started successfully after restart",
			"restart_retry_count": j.restartRetryCount,
		})
		j.isRestarting = false // 成功启动，结束重试
		return nil
	} else {
		ctx.WithLogFields(map[string]interface{}{
			"message":             "recording failed to start after restart, will retry",
			"restart_retry_count": j.restartRetryCount,
		})
		return nil // 继续重试
	}
}

// SetAutoStart 设置自动启动配置
func (j *MicrophoneMonitorJob) SetAutoStart(autoStart bool) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.autoStart = autoStart
}

// GetAutoStart 获取自动启动配置
func (j *MicrophoneMonitorJob) GetAutoStart() bool {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.autoStart
}

// TriggerForceRestart 触发强制重启
func (j *MicrophoneMonitorJob) TriggerForceRestart() {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.forceRestart = true
}

// GetForceRestart 获取强制重启状态
func (j *MicrophoneMonitorJob) GetForceRestart() bool {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.forceRestart
}
