package voice

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	appcfg "github.com/YOUR-ORG/sensecraft-voice/device/agent/cmd/app/config"
	logutil "github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/util/log"
	"github.com/gorilla/websocket"
)

// wsSink: 将 PCM16 小端字节推送到外部 WebSocket
// 仅打印服务端的 connection/final/error 消息
type wsSink struct {
	mu               sync.Mutex
	writeMu          sync.Mutex // 专门用于保护WebSocket写入操作
	url              string
	headers          http.Header
	chunkBytes       int
	maxQueue         int
	maxReconnectWait time.Duration

	conn   *websocket.Conn
	cancel context.CancelFunc
	closed bool

	sendCh chan []byte
}

func newWSSink(opts appcfg.VoiceOptions) (*wsSink, error) {
	if opts.WSUrl == "" {
		return nil, errors.New("ws url is empty")
	}
	hdr := http.Header{}
	for k, v := range opts.WSHeaders {
		hdr.Set(k, v)
	}
	chunk := opts.WSChunkBytes
	if chunk <= 0 {
		chunk = 8192
	}
	q := opts.WSMaxQueue
	if q <= 0 {
		q = 64
	}
	maxRe := opts.WSMaxReconnectDelay
	if maxRe <= 0 {
		maxRe = 15 * time.Second
	}
	s := &wsSink{
		url:              opts.WSUrl,
		headers:          hdr,
		chunkBytes:       chunk,
		maxQueue:         q,
		maxReconnectWait: maxRe,
		sendCh:           make(chan []byte, q),
	}
	s.start()
	return s, nil
}

func (s *wsSink) start() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	go s.loop(ctx)
}

func (s *wsSink) loop(ctx context.Context) {
	// 负责连接、重连、读写
	backoff := time.Second
	for {
		if ctx.Err() != nil {
			return
		}
		c, _, err := websocket.DefaultDialer.Dial(s.url, s.headers)
		if err != nil {
			logutil.Errorf("wsSink: dial failed to %s: %v (retrying in %v)", s.url, err, backoff)
			time.Sleep(backoff)
			backoff *= 2
			if backoff > s.maxReconnectWait {
				backoff = s.maxReconnectWait
			}
			continue
		}
		backoff = time.Second

		s.mu.Lock()
		s.conn = c
		s.mu.Unlock()

		// 设置连接参数
		c.SetReadLimit(1024 * 1024) // 1MB 读取限制
		c.SetPongHandler(func(string) error {
			c.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		// writer/readers
		writeCtx, writeCancel := context.WithCancel(ctx)
		errCh := make(chan error, 3) // 增加心跳协程
		go s.writeLoop(writeCtx, errCh)
		go s.readLoop(writeCtx, errCh)
		go s.heartbeatLoop(writeCtx, errCh)

		// wait for error or ctx done
		select {
		case <-ctx.Done():
			writeCancel()
			_ = c.Close()
			return
		case err = <-errCh:
			writeCancel()
			_ = c.Close()
			logutil.Errorf("wsSink: connection closed: %v", err)
			// retry
		}
	}
}

func (s *wsSink) writeLoop(ctx context.Context, errCh chan<- error) {
	// 聚合发送，尽量按 chunkBytes 切片
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()
	var buf []byte
	for {
		select {
		case <-ctx.Done():
			errCh <- ctx.Err()
			return
		case b := <-s.sendCh:
			buf = append(buf, b...)
			for len(buf) >= s.chunkBytes {
				piece := buf[:s.chunkBytes]
				if err := s.writeBinary(piece); err != nil {
					errCh <- err
					return
				}
				buf = buf[s.chunkBytes:]
			}
		case <-ticker.C:
			if len(buf) > 0 {
				if err := s.writeBinary(buf); err != nil {
					errCh <- err
					return
				}
				buf = buf[:0]
			}
		}
	}
}

func (s *wsSink) writeBinary(b []byte) error {
	s.mu.Lock()
	c := s.conn
	s.mu.Unlock()
	if c == nil {
		return errors.New("ws not connected")
	}

	// 使用写入互斥锁保护WebSocket写入操作
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	c.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return c.WriteMessage(websocket.BinaryMessage, b)
}

func (s *wsSink) readLoop(ctx context.Context, errCh chan<- error) {
	for {
		select {
		case <-ctx.Done():
			errCh <- ctx.Err()
			return
		default:
		}
		s.mu.Lock()
		c := s.conn
		s.mu.Unlock()
		if c == nil {
			errCh <- errors.New("ws not connected")
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
		// 尝试解析 JSON 并仅打印 connection/final/error，同时广播
		var m map[string]any
		if err := json.Unmarshal(data, &m); err != nil {
			continue
		}
		t, _ := m["type"].(string)
		switch t {
		case "connection":
			logutil.Infof("asr[connection]: %s", string(data))
			GetASRHub().Broadcast(data)
		case "final":
			logutil.Infof("asr[final]: %s", string(data))
			GetASRHub().Broadcast(data)
		case "error":
			logutil.Errorf("asr[error]: %s", string(data))
			GetASRHub().Broadcast(data)
		default:
			// ignore others for now
		}
	}
}

// sink interface
func (s *wsSink) OnData(pcm []byte) {
	if len(pcm) == 0 {
		return
	}
	if s.closed {
		return
	}
	// 非阻塞入队，满了则丢弃最旧，保持实时性
	select {
	case s.sendCh <- append([]byte(nil), pcm...):
	default:
		// 丢弃一个旧包，再放入
		select {
		case <-s.sendCh:
		default:
		}
		select {
		case s.sendCh <- append([]byte(nil), pcm...):
		default:
		}
	}
}

func (s *wsSink) Close() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	cancel := s.cancel
	c := s.conn
	s.conn = nil
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if c != nil {
		_ = c.Close()
	}
}

// heartbeatLoop 心跳循环，定期发送心跳消息保持连接
func (s *wsSink) heartbeatLoop(ctx context.Context, errCh chan<- error) {
	ticker := time.NewTicker(30 * time.Second) // 每30秒发送一次心跳
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 使用互斥锁保护写入操作，避免并发写入
			s.mu.Lock()
			if s.conn != nil && !s.closed {
				if err := s.sendHeartbeat(); err != nil {
					logutil.Warnf("wsSink: failed to send heartbeat: %v", err)
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
func (s *wsSink) sendHeartbeat() error {
	// 注意：这个方法应该在调用前已经获得互斥锁
	c := s.conn
	if c == nil {
		return errors.New("websocket not connected")
	}

	// 发送心跳消息
	heartbeatMsg := map[string]interface{}{
		"type":      "heartbeat",
		"timestamp": time.Now().UnixMilli(),
	}

	data, err := json.Marshal(heartbeatMsg)
	if err != nil {
		return errors.New("failed to marshal heartbeat message")
	}

	// 使用写入互斥锁保护WebSocket写入操作
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	c.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return c.WriteMessage(websocket.TextMessage, data)
}

func (s *wsSink) IsDone() bool { return s.closed }
func (s *wsSink) Path() string { return "" }
