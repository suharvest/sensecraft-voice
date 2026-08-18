### WebSocket 实时语音识别接口文档（含流式声纹识别）

#### 基本信息
- 接入地址
  - ws: ws://{host}:{port}/ws（示例: ws://localhost:8080/ws）
- 协议
  - 客户端仅发送二进制音频帧（PCM）
  - 服务端仅发送 JSON 文本消息
- 会话标识
  - session_id 在连接建立时生成，并通过第一条 “connection” 消息返回；在该连接生命周期内不变

### 客户端发送

- 二进制音频（raw PCM）
  - 采样率: 16000 Hz（与配置 audio.sample_rate 一致）
  - 格式: 16-bit signed PCM（小端）
  - 声道: 单声道（双声道将被均值为单声道）
  - 分片建议: 每片 4096～8192 字节，按实时节奏发送
  - 发送节奏: 连续小片，期间可间隔少量延时（50～100ms）

不需要发送任何 JSON 控制消息，音频即数据。

### 服务端返回（JSON）

- 统一外层字段
  - type: "connection" | "vad" | "final" | "error"
  - 其他字段见各类型说明

1) 连接确认
```json
{ "type": "connection", "message": "WebSocket connected, ready for audio", "session_id": "9a8e8ede86b2c84df98623405d8855cf" }
```

2) VAD 状态
- 静音
```json
{ "type": "vad", "status": "silence", "segments": 0, "sessionID": "..." }
```
- 检测到语音（分段数量为提示）
```json
{ "type": "vad", "status": "speech_detected", "segments": 1, "sessionID": "..." }
```
- 语音段完成（Ten-VAD 分支可见）
```json
{ "type": "vad", "status": "speech_segment_complete", "duration": 2.10, "sessionID": "..." }
```

3) 最终识别结果（ASR + 声纹）
- 当累积语音达到最小声纹时长后（speaker.min_audio_seconds），final 将包含 speaker 字段
```json
{
  "type": "final",
  "text": "and presented him with 50 pieces of code",
  "timestamp": 1756374780577,
  "textLength": 40,
  "wordCount": 8,
  "sessionID": "dee8ccad88dfaa0295a2049bd0cdfcc9",
  "speaker": {
    "identified": true,
    "speaker_id": "speaker_01",
    "speaker_name": "用户_01",
    "confidence": 0.78
  }
}