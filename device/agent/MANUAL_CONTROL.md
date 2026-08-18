# 手动控制功能

## 概述

手动控制功能是对麦克风监控功能的增强，允许用户手动停止录音后，监控任务不会自动重新启动录音，直到用户明确重置手动停止状态。

## 功能特性

- **手动停止保护**：用户手动停止录音后，监控任务不会自动启动录音
- **设备丢失响应**：无论手动停止状态如何，设备丢失时都会停止录音
- **状态重置**：提供API接口重置手动停止状态，恢复自动控制
- **状态查询**：API返回录音状态和手动停止状态

## 工作原理

### 状态标志

系统维护一个`manualStop`标志位：
- `true`：用户手动停止，监控任务不会自动启动录音
- `false`：正常状态，监控任务可以自动控制录音

### 状态决策矩阵

| 设备状态 | 录音状态 | manualStop | 操作 | 说明 |
|---------|---------|------------|------|------|
| 存在 | 停止 | false | start | 设备可用且未手动停止，启动录音 |
| 存在 | 停止 | true | 无操作 | 用户手动停止，不自动启动 |
| 不存在 | 运行 | false | stop | 设备丢失，停止录音 |
| 不存在 | 运行 | true | stop | 设备丢失，停止录音 |
| 不存在 | 停止 | false | 无操作 | 设备不可用，保持停止 |
| 不存在 | 停止 | true | 无操作 | 用户手动停止，保持停止 |

### 操作流程

1. **手动停止**：调用`/v1/voice/record`停止录音时，设置`manualStop = true`
2. **手动启动**：调用`/v1/voice/record`启动录音时，设置`manualStop = false`
3. **设备丢失**：无论`manualStop`状态如何，都会停止录音（不设置`manualStop`）
4. **状态重置**：调用`/v1/voice/reset-manual-stop`重置`manualStop = false`

## API接口

### 1. 查询录音状态

**请求：**
```bash
GET /v1/voice/status
```

**响应：**
```json
{
  "code": 200,
  "message": "success",
  "result": {
    "running": true,
    "manual_stop": false
  }
}
```

**字段说明：**
- `running`：当前录音状态
- `manual_stop`：是否处于手动停止状态

### 2. 系统停止录音

**请求：**
```bash
POST /v1/voice/record
Content-Type: application/json

{
  "action": "stop",
  "manualStop": false
}
```

**响应：**
```json
{
  "code": 200,
  "message": "success",
  "result": {
    "running": false,
    "manual_stop": false
  }
}
```

## 使用场景

### 场景1：用户主动停止录音

1. 用户调用`POST /v1/voice/record`停止录音
2. 系统设置`manual_stop = true`
3. 即使麦克风存在，监控任务也不会自动启动录音
4. 用户需要手动启动或重置状态才能恢复自动控制

### 场景2：麦克风设备丢失

1. 无论`manual_stop`状态如何
2. 监控任务检测到设备丢失
3. 系统自动停止录音（不设置`manual_stop`）
4. 设备恢复后，根据`manual_stop`状态决定是否自动启动

### 场景3：恢复自动控制

1. 用户调用`POST /v1/voice/record`停止录音，并设置`manualStop: false`
2. 系统设置`manual_stop = false`
3. 监控任务重新开始自动控制录音
4. 设备可用时自动启动录音

## 配置要求

手动控制功能需要以下配置：

```yaml
voice:
  asr_cache:
    enabled: true  # 必须启用ASR缓存才能启用麦克风监控
```

## 测试

使用提供的测试脚本：

```bash
./test_manual_control.sh
```

测试步骤：
1. 检查初始状态
2. 等待监控任务自动启动录音
3. 手动停止录音，验证`manual_stop = true`
4. 等待监控任务，验证不会自动启动
5. 使用系统停止方式停止录音
6. 等待监控任务，验证重新自动控制

## 实现细节

### 核心修改

1. **Voice Manager**：
   - 添加`manualStop`标志位
   - 添加`SetManualStop()`和`IsManualStop()`方法
   - 修改`StopContinuousWithReason()`方法支持手动停止

2. **Voice Controller**：
   - 添加`StopWithReason()`、`IsManualStop()`、`ResetManualStop()`方法
   - 修改`StartByConfig()`方法清除手动停止标志

3. **API Routes**：
   - 修改`/v1/voice/record`接口使用`StopWithReason()`
   - 添加`/v1/voice/reset-manual-stop`接口
   - 修改`/v1/voice/status`接口返回手动停止状态

4. **麦克风监控**：
   - 更新监控逻辑考虑`manualStop`标志
   - 设备丢失时使用系统停止（不设置手动标志）

### 状态持久化

当前实现中，`manualStop`标志在内存中维护，服务重启后会丢失。如果需要持久化，可以考虑：

1. 将状态保存到配置文件
2. 将状态保存到数据库
3. 将状态保存到Redis等缓存系统

## 注意事项

1. **服务重启**：服务重启后`manual_stop`状态会重置为`false`
2. **并发安全**：所有状态操作都有互斥锁保护
3. **错误处理**：状态操作失败不会影响录音功能
4. **日志记录**：所有状态变化都会记录详细日志

## 故障排除

### 常见问题

1. **手动停止后仍然自动启动**
   - 检查`manual_stop`标志是否正确设置
   - 查看监控任务日志
   - 确认监控任务是否正常运行

2. **重置状态后不自动启动**
   - 检查设备是否可用
   - 查看监控任务状态
   - 确认ASR缓存是否启用

3. **状态不一致**
   - 查看API返回的状态信息
   - 检查服务日志
   - 确认是否有并发操作

### 调试信息

监控任务会记录详细的状态信息：

```json
{
  "action": "start_recording",
  "reason": "device available, recording not running, not manually stopped",
  "device_available": true,
  "is_running": false,
  "is_manual_stop": false
}
```

## 未来改进

1. **状态持久化**：支持状态持久化，服务重启后保持状态
2. **超时机制**：添加手动停止超时，自动恢复自动控制
3. **配置选项**：支持配置是否启用手动控制功能
4. **监控指标**：添加手动控制相关的监控指标
5. **用户界面**：在Web界面中显示手动停止状态
