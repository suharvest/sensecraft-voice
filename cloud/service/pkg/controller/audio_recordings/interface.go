package audio_recordings

import (
	"context"
	"io"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/plugins/minio"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/types"
)

// AudioRecordingInterface 音频录音控制器接口
type AudioRecordingInterface interface {
	// UploadAudioRecording 上传音频录音
	UploadAudioRecording(ctx context.Context, fileHeader *types.FileHeader, req *types.AudioRecordingUploadRequest) (*types.AudioRecordingUploadResponse, error)
	// GetAudioRecording 获取音频录音信息
	GetAudioRecording(ctx context.Context, id string) (*types.AudioRecordingInfo, error)
	// GetAudioRecordingBySessionAndAudio 根据session_id和audio_id获取音频录音信息
	GetAudioRecordingBySessionAndAudio(ctx context.Context, sessionID, audioID string) (*types.AudioRecordingInfo, error)
	// ListAudioRecordings 获取音频录音列表
	ListAudioRecordings(ctx context.Context, req *types.AudioRecordingListRequest) (*types.AudioRecordingListResponse, error)
	// DownloadAudioRecording 下载音频录音
	DownloadAudioRecording(ctx context.Context, id string) (io.ReadCloser, *types.AudioRecordingInfo, error)
	// PlayAudioRecording 直接播放音频录音（通过session_id和audio_id）
	PlayAudioRecording(ctx context.Context, sessionID, audioID string) (io.ReadCloser, *types.AudioRecordingInfo, error)
	// DeleteAudioRecording 删除音频录音
	DeleteAudioRecording(ctx context.Context, id string) error
	// GetMinIOClient 获取MinIO客户端
	GetMinIOClient() minio.Client
}
