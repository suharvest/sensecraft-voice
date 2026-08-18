package voice

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type audioFormatType string

const (
	audioFormatPCM16 audioFormatType = "pcm16"
	audioFormatWAV   audioFormatType = "wav"
)

// sink 接口：接收 PCM16 little-endian 字节流
type sink interface {
	OnData(pcm []byte)
	Close()
	IsDone() bool
	Path() string
}

// base sink
type baseSink struct {
	outF *os.File
	path string
	done bool
}

func (b *baseSink) Close() {
	if b.outF != nil {
		_ = b.outF.Close()
		b.outF = nil
	}
}
func (b *baseSink) IsDone() bool { return b.done }
func (b *baseSink) Path() string { return b.path }

// ---- WebSocket sink（仅发送 PCM，打印 connection/final/error）----
// 说明：wsSink 异步发送，读协程仅解析并打印部分事件
// 为简化初版，不做事件分发，仅日志输出。
// 后续可抽象到单独文件。

// 为避免新增依赖，这里使用标准库 net/http + gorilla/websocket 可替换。
// 由于当前项目未引入 gorilla/websocket，这里暂不实现具体逻辑，
// 在 voice.Manager 中将根据配置占位。

// ContinuousFileSink 持续写文件
type ContinuousFileSink struct {
	baseSink
	format       audioFormatType
	sampleRate   int
	channels     int
	bytesWritten int64
}

// ClipFileSink 写定长秒数
type ClipFileSink struct {
	baseSink
	format         audioFormatType
	sampleRate     int
	channels       int
	bytesRemaining int64
	bytesWritten   int64
}

// newFilePath 构造目录+时间名
func newFilePath(dir string, format audioFormatType) (string, error) {
	dir = filepath.Clean(dir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	ext := ".wav"
	if format == audioFormatPCM16 {
		ext = ".pcm"
	}
	name := time.Now().Format("20060102_150405") + ext
	return filepath.Join(dir, name), nil
}

func writeWAVHeader(f *os.File, sampleRate, channels int) error {
	var (
		audioFormat   uint16 = 1
		bitsPerSample uint16 = 16
		byteRate             = uint32(sampleRate * channels * int(bitsPerSample) / 8)
		blockAlign    uint16 = uint16(channels * int(bitsPerSample) / 8)
	)
	if _, err := f.Write([]byte("RIFF")); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint32(0)); err != nil {
		return err
	}
	if _, err := f.Write([]byte("WAVE")); err != nil {
		return err
	}
	if _, err := f.Write([]byte("fmt ")); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint32(16)); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, audioFormat); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint16(channels)); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint32(sampleRate)); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, byteRate); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, blockAlign); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, bitsPerSample); err != nil {
		return err
	}
	if _, err := f.Write([]byte("data")); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint32(0)); err != nil {
		return err
	}
	return nil
}

func finalizeWAV(f *os.File, dataSize int64) {
	if f == nil {
		return
	}
	if _, err := f.Seek(40, 0); err == nil {
		_ = binary.Write(f, binary.LittleEndian, uint32(dataSize))
	}
	if _, err := f.Seek(4, 0); err == nil {
		_ = binary.Write(f, binary.LittleEndian, uint32(36+dataSize))
	}
	_, _ = f.Seek(0, 2)
}

func newContinuousFileSink(dirOrFile string, format audioFormatType, sampleRate, channels int) (*ContinuousFileSink, error) {
	path := dirOrFile
	// 若传入是目录则生成文件
	if strings.HasSuffix(strings.ToLower(dirOrFile), ".wav") || strings.HasSuffix(strings.ToLower(dirOrFile), ".pcm") {
		// treat as file path
		if err := os.MkdirAll(filepath.Dir(dirOrFile), 0755); err != nil {
			return nil, err
		}
	} else {
		// treat as dir
		p, err := newFilePath(dirOrFile, format)
		if err != nil {
			return nil, err
		}
		path = p
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	s := &ContinuousFileSink{baseSink: baseSink{outF: f, path: path}, format: format, sampleRate: sampleRate, channels: channels}
	if format == audioFormatWAV {
		if err := writeWAVHeader(f, sampleRate, channels); err != nil {
			_ = f.Close()
			return nil, err
		}
	}
	return s, nil
}

func (s *ContinuousFileSink) OnData(pcm []byte) {
	if s.done || s.outF == nil || len(pcm) == 0 {
		return
	}
	if n, err := s.outF.Write(pcm); err == nil {
		s.bytesWritten += int64(n)
	}
}

func (s *ContinuousFileSink) Close() {
	if s.format == audioFormatWAV {
		finalizeWAV(s.outF, s.bytesWritten)
	}
	s.done = true
	s.baseSink.Close()
}

func newClipFileSink(dir string, seconds int, sampleRate, channels int) (*ClipFileSink, error) {
	path, err := newFilePath(dir, audioFormatWAV)
	if err != nil {
		return nil, err
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	if err := writeWAVHeader(f, sampleRate, channels); err != nil {
		_ = f.Close()
		return nil, err
	}
	bytesTotal := int64(seconds) * int64(sampleRate*channels*2)
	return &ClipFileSink{baseSink: baseSink{outF: f, path: path}, format: audioFormatWAV, sampleRate: sampleRate, channels: channels, bytesRemaining: bytesTotal}, nil
}

func (s *ClipFileSink) OnData(pcm []byte) {
	if s.done || s.outF == nil || s.bytesRemaining <= 0 || len(pcm) == 0 {
		return
	}
	writeN := int64(len(pcm))
	if writeN > s.bytesRemaining {
		writeN = s.bytesRemaining
	}
	if writeN > 0 {
		if n, err := s.outF.Write(pcm[:writeN]); err == nil {
			s.bytesRemaining -= int64(n)
			s.bytesWritten += int64(n)
		}
	}
	if s.bytesRemaining <= 0 {
		finalizeWAV(s.outF, s.bytesWritten)
		s.done = true
		s.baseSink.Close()
	}
}
