package voice

import (
	"sync"
)

// ASRHub 提供进程内广播：下游 WS 订阅者会收到来自 wsSink 的消息
type ASRHub struct {
	mu          sync.Mutex
	subscribers map[chan []byte]struct{}
}

var (
	hubInst *ASRHub
	hubOnce sync.Once
)

func GetASRHub() *ASRHub {
	hubOnce.Do(func() {
		hubInst = &ASRHub{subscribers: make(map[chan []byte]struct{})}
	})
	return hubInst
}

func (h *ASRHub) Subscribe(buffer int) (ch chan []byte, unsubscribe func()) {
	if buffer <= 0 {
		buffer = 16
	}
	ch = make(chan []byte, buffer)
	h.mu.Lock()
	h.subscribers[ch] = struct{}{}
	h.mu.Unlock()
	return ch, func() {
		h.mu.Lock()
		if _, ok := h.subscribers[ch]; ok {
			delete(h.subscribers, ch)
			close(ch)
		}
		h.mu.Unlock()
	}
}

func (h *ASRHub) Broadcast(msg []byte) {
	if len(msg) == 0 {
		return
	}
	h.mu.Lock()
	for ch := range h.subscribers {
		select {
		case ch <- append([]byte(nil), msg...):
		default:
			// 丢弃慢消费者的数据，保证实时性
		}
	}
	h.mu.Unlock()
}
