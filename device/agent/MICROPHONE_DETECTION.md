# 麦克风检测功能

## 概述

本功能使用系统命令替代 `malgo` 库来检测麦克风设备，解决了原有的 SIGBUS 错误问题，并提供了跨平台兼容性。

## 实现方案

### 平台支持

| 平台 | 命令 | 说明 |
|------|------|------|
| Linux | `arecord -l` | 需要安装 `alsa-utils` 包 |
| macOS | `system_profiler SPAudioDataType` | 系统自带 |
| Windows | PowerShell WMI | 系统自带 |

### 核心优势

1. **避免 CGO 问题**：不再使用 `malgo` 库，避免了 SIGBUS 错误
2. **跨平台兼容**：支持 Linux、macOS、Windows
3. **简单可靠**：使用系统标准工具，稳定性更好
4. **性能优秀**：执行速度快，资源占用少

## 使用方法

### 检测任何音频设备

```go
available, err := isMicrophoneAvailable("")
if err != nil {
    log.Printf("检测失败: %v", err)
} else {
    log.Printf("音频设备可用: %t", available)
}
```

### 检测特定设备

```go
available, err := isMicrophoneAvailable("ReSpeaker")
if err != nil {
    log.Printf("检测失败: %v", err)
} else {
    log.Printf("ReSpeaker 设备可用: %t", available)
}
```

## 配置要求

### Linux
```bash
# Ubuntu/Debian
sudo apt-get install alsa-utils

# CentOS/RHEL
sudo yum install alsa-utils
```

### macOS
无需额外安装，使用系统自带的 `system_profiler` 命令。

### Windows
无需额外安装，使用系统自带的 PowerShell 和 WMI。

## 输出格式解析

### Linux (arecord -l)
```
card 0: PCH [HDA Intel PCH], device 0: ALC892 Analog [ALC892 Analog]
card 1: ReSpeaker [ReSpeaker 4 Mic Array], device 0: USB Audio [USB Audio]
```

### macOS (system_profiler)
```
        ReSpeaker 4 Mic Array (UAC1.0):
          Default Input Device: Yes
          Input Channels: 1
          Manufacturer: SEEED
```

### Windows (PowerShell)
```
Name: Realtek High Definition Audio
Status: OK
```

## 错误处理

- 命令执行失败时返回错误
- 不支持的平台返回 `false, nil`
- 连续错误过多时会暂时跳过检查

## 监控任务配置

```yaml
voice:
  asr_cache:
    enabled: true
  device_id: "ReSpeaker"  # 可选，指定特定设备
```

## 日志输出

```
[INFO] Audio device check output: [系统命令输出]
[DEBUG] Audio device check command failed: [错误信息]
[WARN] Unsupported platform for microphone detection: [平台名称]
```

## 测试

运行测试脚本验证功能：

```bash
# 测试系统命令
./test_microphone_detection.sh

# 测试 Go 程序
go run test_microphone_go.go
```

## 故障排除

### 1. Linux 上 arecord 命令不存在
```bash
sudo apt-get install alsa-utils
```

### 2. macOS 上权限问题
确保应用有访问音频设备的权限：
- 系统偏好设置 → 安全性与隐私 → 隐私 → 麦克风

### 3. Windows 上 PowerShell 执行策略
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

## 性能特点

- **执行时间**：通常 < 100ms
- **内存占用**：极低，无 CGO 开销
- **CPU 占用**：极低，仅执行系统命令
- **稳定性**：高，使用系统标准工具
