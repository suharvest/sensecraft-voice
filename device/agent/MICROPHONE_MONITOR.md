# 麦克风监控功能

## 概述

麦克风监控功能是一个基于jobmanager的定时任务，用于监控麦克风设备的插拔状态，并自动控制录音的开始和停止。

## 功能特性

- **自动设备检测**：每30秒检查一次麦克风设备是否可用
- **状态稳定机制**：需要连续2次检查状态一致才执行操作，避免频繁启停
- **错误处理**：连续错误超过5次时暂停监控，避免影响系统稳定性
- **配置驱动**：只有当`asr_cache.enabled=true`时才启用监控
- **状态同步**：确保录音状态与实际设备状态保持一致

## 工作原理

### 状态决策矩阵

| 设备状态 | 当前录音状态 | 操作 | 说明 |
|---------|-------------|------|------|
| 存在 | 运行中 | 无操作 | 正常状态 |
| 存在 | 停止 | start | 设备可用，启动录音 |
| 不存在 | 运行中 | stop | 设备丢失，停止录音 |
| 不存在 | 停止 | 无操作 | 设备不可用，保持停止 |

### 监控流程

1. **设备检测**：使用malgo库获取当前可用的录音设备列表
2. **状态比较**：与上次检查的设备状态进行比较
3. **稳定期检查**：只有状态连续2次一致才执行操作
4. **录音控制**：根据状态矩阵调用相应的录音API
5. **错误处理**：记录错误并实现退避机制

## 配置

### 启用条件

麦克风监控功能只有在以下条件同时满足时才会启用：

```yaml
voice:
  asr_cache:
    enabled: true  # 必须启用ASR缓存
```

### 监控参数

- **检查间隔**：30秒
- **状态稳定阈值**：2次连续检查
- **最大连续错误**：5次
- **监控设备**：使用配置中的`voice.deviceId`，为空则监控默认设备

## 日志

监控任务会记录详细的日志信息：

```json
{
  "device_available": true,
  "last_device_state": false,
  "state_stable_count": 2,
  "state_stable_threshold": 2,
  "consecutive_errors": 0,
  "action": "start_recording",
  "reason": "device available but recording not running"
}
```

## 状态查询

可以通过以下方式查看监控状态：

```bash
# 查看录音状态
curl http://localhost:8080/v1/voice/status

# 查看设备状态
curl http://localhost:8080/v1/voice/device/status
```

## 测试

使用提供的测试脚本：

```bash
./test_microphone_monitor.sh
```

测试步骤：
1. 检查初始状态
2. 等待监控任务执行
3. 手动停止录音
4. 等待监控任务重新启动录音
5. 验证最终状态

## 故障排除

### 常见问题

1. **监控不工作**
   - 检查`asr_cache.enabled`是否为true
   - 查看日志中的错误信息
   - 确认voice controller是否正常初始化

2. **频繁启停**
   - 检查设备是否稳定
   - 查看状态稳定计数是否正常
   - 确认设备检测逻辑是否正确

3. **错误过多**
   - 查看连续错误计数
   - 检查malgo库是否正常工作
   - 确认设备权限是否正确

### 调试信息

监控任务会记录以下调试信息：

- 设备可用性状态
- 状态稳定计数
- 连续错误计数
- 执行的操作和原因
- 错误详情

## 实现细节

### 文件结构

```
pkg/jobmanager/
├── microphone_monitor.go    # 麦克风监控任务实现
├── manager.go              # Job管理器
└── ...

cmd/app/options/
└── options.go              # 任务注册

pkg/controller/voice/
└── voice.go                # 添加IsRunning方法
```

### 关键接口

```go
type VoiceController interface {
    IsRunning() bool
    StartByConfig(ctx context.Context) error
    Stop(ctx context.Context) error
}
```

### 配置集成

麦克风监控任务在`options.go`中注册：

```go
microphoneMonitorJob := jobmanager.NewMicrophoneMonitorJob(
    o.ComponentConfig.Voice.ASRCache.Enabled,
    o.ComponentConfig.Voice.DeviceID,
    o.Controller.Voice(),
)
```

## 注意事项

1. **性能影响**：每30秒创建malgo context可能有一定性能开销
2. **权限要求**：需要音频设备访问权限
3. **平台兼容性**：依赖malgo库的平台支持
4. **错误恢复**：连续错误过多时会暂停监控，需要手动重启服务恢复

## 未来改进

1. **优化性能**：复用malgo context或使用更轻量的检测方法
2. **配置扩展**：支持更多监控参数配置
3. **监控指标**：添加Prometheus指标支持
4. **设备切换**：支持自动切换到备用设备
