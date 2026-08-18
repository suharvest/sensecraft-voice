## Voice 录音接口

- 统一返回体：`{"code":200,"message":"success","result":{...}}`；异常：`{"code":400,"message":"错误信息"}`。

### 开始/停止录音
- method: POST
- URL: /v/voice/record
- 正常返回码: 200
- 异常返回码: badRequest(400), unauthorized(401), forbidden(403)
- 请求体(JSON):
  - action: string, 必填，"start" | "stop"
  - deviceId: string, 可选，默认系统输入设备
  - sampleRate: int, 可选，默认 16000
  - channels: int, 可选，默认 1
  - filePath: string, 可选；未提供时默认写入 `./recordings/voice/{yyyyMMdd_HHmmss}_{uuid}.wav`；若配置为目录，则文件名为开始时间戳
  - softMute: bool, 可选，是否软静音（丢帧不写出）
  - output: string, 可选，"file" | "stream" | "both"；不传则使用 `config.yaml` 中的 `voice.output`
  - manualStop: bool, 可选，仅对stop操作有效，是否手动停止（默认true）
- 说明：
  - 覆盖规则：请求体中提供的 `deviceId`、`sampleRate`、`channels`、`filePath`、`output` 将覆盖 `config.yaml` 对应字段；未提供的字段沿用配置。
  - 输出格式：支持 `format=pcm16`（原始 PCM）与 `format=wav`（标准 WAV）。当前通过配置与内部策略确定；未传 `filePath` 时：若 `format=wav` 则后缀 `.wav`；若 `format=pcm16` 则 `.pcm`。
  - 手动停止：当 `action=stop` 时，`manualStop=true`（默认）表示手动停止，监控任务不会自动启动录音；`manualStop=false` 表示系统停止，监控任务可以自动启动录音。
- 响应结果:
  - result.running: bool，当前录音状态
  - result.manual_stop: bool，是否处于手动停止状态（仅stop操作返回）

### 查询录音状态
- method: GET
- URL: /v1/voice/status
- 正常返回码: 200
- 响应结果:
  - result.running: bool，当前录音状态
  - result.manual_stop: bool，是否处于手动停止状态

### 麦克风测试（边录边放）
- method: POST
- URL: /v1/voice/test
- 正常返回码: 200
- 异常返回码: badRequest(400), unauthorized(401), forbidden(403)
- 请求体(JSON):
  - action: string, 必填，"start" | "stop"
  - deviceId: string, 可选，默认系统输入设备
  - sampleRate: int, 可选，默认 16000
  - channels: int, 可选，默认 1
- 响应结果:
  - result.testing: bool，当前是否处于测试回路

### 实时识别订阅（WebSocket）
- method: GET
- URL: /v1/voice/asr-ws
- 说明:
  - 客户端连接后即可实时接收来自外部 ASR 的关键 JSON 文本消息。
  - 仅下行推送，客户端无需发送任何消息。
- 消息类型:
  - type: "connection" | "final" | "error"
- 示例:
  - connection
    ```json
    { "type": "connection", "message": "WebSocket connected, ready for audio", "session_id": "9a8e8ede86b2c84df98623405d8855cf" }
    ```
  - final（示例）
    ```json
    {
      "type": "final",
      "text": "and presented him with 50 pieces of code",
      "timestamp": 1756374780577,
      "textLength": 40,
      "wordCount": 8,
      "sessionID": "dee8ccad88dfaa0295a2049bd0cdfcc9",
      "speaker": { "identified": true, "speaker_id": "speaker_01", "speaker_name": "用户_01", "confidence": 0.78 }
    }
    ```
  - error
    ```json
    { "type": "error", "message": "some error", "code": 500 }
    ```

### 外部 ASR 推流（说明）
- 系统作为客户端根据配置 `voice.wsUrl` 连接外部 ASR WebSocket，仅发送二进制 PCM16 LE 音频帧（默认 16kHz、单声道）。
- 发送分片：建议 4096～8192 字节；发送节奏约 50～100ms（系统内部已处理）。
- 外部返回的 `connection`/`final`/`error` 会被打印到日志，并通过上述 `/v1/voice/asr-ws` 转发给订阅者。
