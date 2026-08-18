package voice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	logutil "github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/util/log"
	"github.com/gorilla/websocket"
)

// RemoteSink: 接收 ASR 识别结果并推送到远程 WebSocket 服务
// 支持 mac_address 配置，用于设备标识
type RemoteSink struct {
	mu               sync.Mutex
	baseURL          string
	wsURL            string
	macAddress       string
	headers          http.Header
	maxQueue         int
	maxReconnectWait time.Duration

	conn   *websocket.Conn
	cancel context.CancelFunc
	closed bool

	// 改为接收 ASR 结果消息
	asrResultCh chan map[string]interface{}

	// 连接管理字段
	reconnectCount        int
	lastConnectTime       time.Time
	connectionEstablished bool
}

// RemoteSinkConfig 远程 sink 配置
type RemoteSinkConfig struct {
	BaseURL          string
	MacAddress       string
	Headers          map[string]string
	MaxQueue         int
	MaxReconnectWait time.Duration
}

// newRemoteSink 创建新的远程 sink
func newRemoteSink(cfg RemoteSinkConfig) (*RemoteSink, error) {
	if cfg.BaseURL == "" {
		return nil, errors.New("remote base URL is empty")
	}

	// 构造 WebSocket URL
	baseURL, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	// 将 HTTP/HTTPS 转换为 WebSocket
	wsScheme := "ws"
	if baseURL.Scheme == "https" {
		wsScheme = "wss"
	}

	wsURL := fmt.Sprintf("%s://%s/api/v1/recordings/stream", wsScheme, baseURL.Host)

	// 设置默认值
	q := cfg.MaxQueue
	if q <= 0 {
		q = 64
	}

	maxRe := cfg.MaxReconnectWait
	if maxRe <= 0 {
		maxRe = 15 * time.Second
	}

	// 构造请求头
	headers := http.Header{}
	if cfg.Headers != nil {
		for k, v := range cfg.Headers {
			headers.Set(k, v)
		}
	}

	// 添加必要的请求头
	if headers.Get("User-Agent") == "" {
		headers.Set("User-Agent", "SenseCraft-Voice-Client/1.0")
	}
	if headers.Get("Accept") == "" {
		headers.Set("Accept", "*/*")
	}

	s := &RemoteSink{
		baseURL:          cfg.BaseURL,
		wsURL:            wsURL,
		macAddress:       cfg.MacAddress,
		headers:          headers,
		maxQueue:         q,
		maxReconnectWait: maxRe,
		asrResultCh:      make(chan map[string]interface{}, q),
		lastConnectTime:  time.Now(),
	}

	s.start()
	return s, nil
}

// start 启动远程 sink
func (s *RemoteSink) start() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	go s.loop(ctx)
}

// loop 主循环：负责连接、重连、读写
func (s *RemoteSink) loop(ctx context.Context) {
	backoff := time.Second
	minBackoff := 2 * time.Second

	for {
		if ctx.Err() != nil {
			return
		}

		// 检查是否需要延迟重连
		if s.reconnectCount > 0 {
			timeSinceLastConnect := time.Since(s.lastConnectTime)
			if timeSinceLastConnect < minBackoff {
				waitTime := minBackoff - timeSinceLastConnect
				logutil.Infof("remoteSink: waiting %v before reconnecting", waitTime)
				time.Sleep(waitTime)
			}
		}

		logutil.Infof("remoteSink: connecting to %s (attempt %d)", s.wsURL, s.reconnectCount+1)
		s.lastConnectTime = time.Now()

		// 创建带超时的拨号器
		dialer := websocket.Dialer{
			HandshakeTimeout:  10 * time.Second,
			EnableCompression: false, // 禁用压缩以提高稳定性
		}

		c, resp, err := dialer.Dial(s.wsURL, s.headers)
		if err != nil {
			if resp != nil {
				logutil.Errorf("remoteSink: dial failed with response %d: %v", resp.StatusCode, err)
			} else {
				logutil.Errorf("remoteSink: dial failed: %v", err)
			}

			s.reconnectCount++
			backoff *= 2
			if backoff > s.maxReconnectWait {
				backoff = s.maxReconnectWait
			}

			logutil.Infof("remoteSink: will retry in %v", backoff)
			time.Sleep(backoff)
			continue
		}

		// 连接成功，重置重连计数和退避时间
		s.reconnectCount = 0
		backoff = time.Second

		// 设置连接参数
		c.SetReadLimit(1024 * 1024) // 1MB 读取限制
		c.SetPongHandler(func(string) error {
			c.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		logutil.Infof("remoteSink: connected to %s", s.wsURL)

		s.mu.Lock()
		s.conn = c
		s.connectionEstablished = true
		s.mu.Unlock()

		// 发送初始化消息，保持连接活跃
		if err := s.sendInitMessage(); err != nil {
			logutil.Errorf("remoteSink: failed to send init message: %v", err)
			_ = c.Close()
			s.mu.Lock()
			s.connectionEstablished = false
			s.mu.Unlock()
			time.Sleep(backoff)
			continue
		}

		// 启动读写协程
		writeCtx, writeCancel := context.WithCancel(ctx)
		errCh := make(chan error, 3) // 增加缓冲区大小
		go s.writeLoop(writeCtx, errCh)
		go s.readLoop(writeCtx, errCh)
		go s.heartbeatLoop(writeCtx, errCh)

		// 等待错误或上下文取消
		select {
		case <-ctx.Done():
			writeCancel()
			_ = c.Close()
			return
		case err = <-errCh:
			writeCancel()
			_ = c.Close()

			s.mu.Lock()
			s.connectionEstablished = false
			s.mu.Unlock()

			// 改进错误分析和处理
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					logutil.Infof("remoteSink: normal close, will retry")
				} else if errors.Is(err, context.Canceled) {
					logutil.Infof("remoteSink: context canceled, will retry")
				} else {
					logutil.Errorf("remoteSink: connection error: %v", err)

					// 对于连接重置等网络错误，增加延迟
					if isNetworkError(err) {
						logutil.Warnf("remoteSink: network error detected, increasing backoff")
						backoff *= 2
						if backoff > s.maxReconnectWait {
							backoff = s.maxReconnectWait
						}
					}
				}
			}

			// 添加额外延迟，避免过于频繁的重连
			time.Sleep(backoff)
		}
	}
}

// isNetworkError 判断是否为网络相关错误
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return contains(errStr, "connection reset by peer") ||
		contains(errStr, "broken pipe") ||
		contains(errStr, "connection refused") ||
		contains(errStr, "no route to host") ||
		contains(errStr, "network is unreachable")
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

// containsSubstring 简单的子字符串检查
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// writeLoop 写入循环：发送 ASR 结果到远程服务
func (s *RemoteSink) writeLoop(ctx context.Context, errCh chan<- error) {
	for {
		select {
		case <-ctx.Done():
			errCh <- ctx.Err()
			return
		case asrResult := <-s.asrResultCh:
			if err := s.writeASRResult(asrResult); err != nil {
				errCh <- err
				return
			}
		}
	}
}

// writeASRResult 发送 ASR 结果到远程服务
func (s *RemoteSink) writeASRResult(asrResult map[string]interface{}) error {
	s.mu.Lock()
	c := s.conn
	established := s.connectionEstablished
	s.mu.Unlock()

	if c == nil || !established {
		return errors.New("websocket not connected or not established")
	}

	// 添加设备信息到 ASR 结果
	if s.macAddress != "" {
		asrResult["mac_address"] = s.macAddress
	}

	data, err := json.Marshal(asrResult)
	if err != nil {
		return fmt.Errorf("failed to marshal ASR result: %w", err)
	}
	logutil.Infof("remoteSink: sending ASR result: %s", string(data))

	c.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return c.WriteMessage(websocket.TextMessage, data)
}

// writeBinary 写入二进制数据（保留用于兼容性）
func (s *RemoteSink) writeBinary(b []byte) error {
	// 将二进制数据转换为 ASR 结果格式
	asrResult := map[string]interface{}{
		"type":      "binary_data",
		"data":      b,
		"timestamp": time.Now().UnixMilli(),
		"length":    len(b),
	}
	return s.writeASRResult(asrResult)
}

// readLoop 读取循环：处理远程服务响应（类似 ASR 结果）
func (s *RemoteSink) readLoop(ctx context.Context, errCh chan<- error) {
	for {
		select {
		case <-ctx.Done():
			errCh <- ctx.Err()
			return
		default:
		}

		s.mu.Lock()
		c := s.conn
		established := s.connectionEstablished
		s.mu.Unlock()

		if c == nil || !established {
			errCh <- errors.New("websocket not connected or not established")
			return
		}

		c.SetReadDeadline(time.Now().Add(60 * time.Second))
		mt, data, err := c.ReadMessage()
		if err != nil {
			errCh <- err
			return
		}

		if mt != websocket.TextMessage {
			continue
		}

		// 解析远程服务响应（类似 ASR 结果）
		var response map[string]interface{}
		if err := json.Unmarshal(data, &response); err != nil {
			logutil.Warnf("remoteSink: failed to parse response: %v", err)
			continue
		}

		// 记录重要响应，格式与 ws_sink 一致
		if msgType, ok := response["type"].(string); ok {
			switch msgType {
			case "connection":
				logutil.Infof("remoteSink[connection]: %s", string(data))
			case "final":
				logutil.Infof("remoteSink[final]: %s", string(data))
			case "error":
				logutil.Errorf("remoteSink[error]: %s", string(data))
			default:
				logutil.Debugf("remoteSink[%s]: %s", msgType, string(data))
			}
		}
	}
}

// OnData 接收 ASR 识别结果（不再是音频数据）
func (s *RemoteSink) OnData(asrResult map[string]interface{}) {
	if asrResult == nil {
		return
	}

	if s.closed {
		return
	}

	// 检查连接状态
	s.mu.Lock()
	established := s.connectionEstablished
	s.mu.Unlock()

	if !established {
		logutil.Debugf("remoteSink: connection not established, dropping ASR result")
		return
	}

	// 非阻塞入队，满了则丢弃最旧的结果，保持实时性
	select {
	case s.asrResultCh <- asrResult:
		// 记录第一个结果
		if len(s.asrResultCh) == 1 {
			logutil.Debugf("remoteSink: first ASR result queued")
		}
	default:
		// 丢弃一个旧结果，再放入新结果
		select {
		case <-s.asrResultCh:
		default:
		}
		select {
		case s.asrResultCh <- asrResult:
		default:
			// 如果还是满的，记录警告
			logutil.Warnf("remoteSink: queue full, dropping ASR result")
		}
	}
}

// OnASRResult 专门用于接收 ASR 结果的方法
func (s *RemoteSink) OnASRResult(asrResult map[string]interface{}) {
	s.OnData(asrResult)
}

// Close 关闭远程 sink
func (s *RemoteSink) Close() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	cancel := s.cancel
	c := s.conn
	s.conn = nil
	s.connectionEstablished = false
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	if c != nil {
		_ = c.Close()
	}

	logutil.Infof("remoteSink: closed")
}

// IsDone 检查是否已关闭
func (s *RemoteSink) IsDone() bool {
	return s.closed
}

// IsConnected 检查WebSocket连接是否已建立
func (s *RemoteSink) IsConnected() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.conn != nil && s.connectionEstablished && !s.closed
}

// Path 返回空字符串（远程 sink 没有本地文件路径）
func (s *RemoteSink) Path() string {
	return ""
}

// GetMacAddress 获取 MAC 地址
func (s *RemoteSink) GetMacAddress() string {
	return s.macAddress
}

// GetBaseURL 获取基础 URL
func (s *RemoteSink) GetBaseURL() string {
	return s.baseURL
}

// GetWebSocketURL 获取 WebSocket URL
func (s *RemoteSink) GetWebSocketURL() string {
	return s.wsURL
}

// sendInitMessage 发送初始化消息，保持连接活跃
func (s *RemoteSink) sendInitMessage() error {
	s.mu.Lock()
	c := s.conn
	established := s.connectionEstablished
	s.mu.Unlock()

	if c == nil || !established {
		return errors.New("websocket not connected or not established")
	}

	// 发送初始化消息，包含设备信息
	initMsg := map[string]interface{}{
		"type": "init",
		"device": map[string]interface{}{
			"mac_address": s.macAddress,
			"timestamp":   time.Now().UnixMilli(),
		},
		"capabilities": map[string]interface{}{
			"format":    "json",
			"max_queue": s.maxQueue,
		},
	}

	data, err := json.Marshal(initMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal init message: %w", err)
	}

	c.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return c.WriteMessage(websocket.TextMessage, data)
}

// heartbeatLoop 心跳循环，定期发送心跳消息保持连接
func (s *RemoteSink) heartbeatLoop(ctx context.Context, errCh chan<- error) {
	ticker := time.NewTicker(30 * time.Second) // 每30秒发送一次心跳
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 使用互斥锁保护写入操作，避免并发写入
			s.mu.Lock()
			if s.conn != nil && s.connectionEstablished && !s.closed {
				if err := s.sendHeartbeat(); err != nil {
					logutil.Warnf("remoteSink: failed to send heartbeat: %v", err)
					s.mu.Unlock()
					errCh <- err
					return
				}
			}
			s.mu.Unlock()
		}
	}
}

// sendHeartbeat 发送心跳消息
func (s *RemoteSink) sendHeartbeat() error {
	// 注意：这个方法应该在调用前已经获得互斥锁
	c := s.conn
	if c == nil {
		return errors.New("websocket not connected")
	}

	// 发送心跳消息
	heartbeatMsg := map[string]interface{}{
		"type":      "heartbeat",
		"timestamp": time.Now().UnixMilli(),
		"device": map[string]interface{}{
			"mac_address": s.macAddress,
		},
	}

	data, err := json.Marshal(heartbeatMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal heartbeat message: %w", err)
	}

	c.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return c.WriteMessage(websocket.TextMessage, data)
}
