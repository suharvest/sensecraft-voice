package audio_recordings

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/config"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/plugins/minio"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/types"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/util/errors"
	"github.com/google/uuid"
	"k8s.io/klog/v2"
)

// AudioRecordingController 音频录音控制器
type AudioRecordingController struct {
	config      config.Config
	factory     db.ShareDaoFactory
	minioClient minio.Client
}

// NewAudioRecordingController 创建音频录音控制器
func NewAudioRecordingController(cfg config.Config, factory db.ShareDaoFactory, minioClient minio.Client) *AudioRecordingController {
	return &AudioRecordingController{
		config:      cfg,
		factory:     factory,
		minioClient: minioClient,
	}
}

// AudioRecordingGetter 音频录音获取器接口
type AudioRecordingGetter interface {
	AudioRecording() AudioRecordingInterface
}

// UploadAudioRecording 上传音频录音
func (a *AudioRecordingController) UploadAudioRecording(ctx context.Context, fileHeader *types.FileHeader, req *types.AudioRecordingUploadRequest) (*types.AudioRecordingUploadResponse, error) {
	// 1. 验证文件
	if err := a.validateAudioFile(fileHeader); err != nil {
		return nil, err
	}

	// 2. 验证请求参数
	if err := a.validateUploadRequest(req); err != nil {
		return nil, err
	}

	// 3. 检查是否已存在相同的session_id+audio_id组合
	existing, err := a.factory.AudioRecording().GetBySessionIDAndAudioID(ctx, req.SessionID, req.AudioID)
	if err == nil && existing != nil {
		klog.Warningf("Audio recording already exists: session_id=%s, audio_id=%s", req.SessionID, req.AudioID)
		return nil, errors.ErrAlreadyExists
	}

	// 4. 生成文件信息
	fileID := uuid.New().String()
	fileSize := fileHeader.Size
	contentType := "audio/wav"

	// 5. 生成MinIO存储路径
	minioPath := a.generateAudioPath(req.SessionID, req.AudioID)

	// 6. 读取文件内容到内存（用于校验和计算和MinIO上传）
	fileData, err := io.ReadAll(fileHeader.File)
	if err != nil {
		klog.Errorf("Failed to read file content: %v", err)
		return nil, errors.ErrServerInternal
	}

	// 7. 计算文件校验和
	checksum, err := a.calculateChecksumFromBytes(fileData)
	if err != nil {
		klog.Errorf("Failed to calculate file checksum: %v", err)
		return nil, errors.ErrServerInternal
	}
	klog.V(2).Infof("File checksum: %s", checksum)

	// 8. 上传到MinIO
	if err := a.minioClient.UploadFile(ctx, "", minioPath, bytes.NewReader(fileData), fileSize, contentType); err != nil {
		klog.Errorf("Failed to upload audio file to MinIO: %v", err)
		return nil, errors.ErrMinIOUploadFailed
	}

	// 9. 保存到数据库
	recording := &model.AudioRecording{
		ID:         fileID,
		SessionID:  req.SessionID,
		AudioID:    req.AudioID,
		MacAddress: req.MacAddress,
		FilePath:   minioPath,
		FileSize:   fileSize,
		UploadTime: model.GetCurrentTimestamp(),
		Status:     model.AudioRecordingStatusNormal,
	}

	createdRecording, err := a.factory.AudioRecording().Create(ctx, recording)
	if err != nil {
		// 回滚MinIO文件
		if deleteErr := a.minioClient.DeleteFile(ctx, "", minioPath); deleteErr != nil {
			klog.Errorf("Failed to rollback MinIO file %s: %v", minioPath, deleteErr)
		}
		klog.Errorf("Failed to save audio recording record to database: %v", err)
		return nil, errors.ErrServerInternal
	}

	return &types.AudioRecordingUploadResponse{
		ID:         createdRecording.ID,
		SessionID:  createdRecording.SessionID,
		AudioID:    createdRecording.AudioID,
		FileSize:   createdRecording.FileSize,
		UploadTime: createdRecording.UploadTime,
	}, nil
}

// GetAudioRecording 获取音频录音信息
func (a *AudioRecordingController) GetAudioRecording(ctx context.Context, id string) (*types.AudioRecordingInfo, error) {
	recording, err := a.factory.AudioRecording().GetByID(ctx, id)
	if err != nil {
		klog.Errorf("Failed to get audio recording %s: %v", id, err)
		return nil, errors.ErrNotFound
	}

	return &types.AudioRecordingInfo{
		ID:         recording.ID,
		SessionID:  recording.SessionID,
		AudioID:    recording.AudioID,
		MacAddress: recording.MacAddress,
		FilePath:   recording.FilePath,
		FileSize:   recording.FileSize,
		UploadTime: recording.UploadTime,
		Status:     recording.Status,
		CreatedAt:  recording.CreatedAt,
		UpdatedAt:  recording.UpdatedAt,
	}, nil
}

// GetAudioRecordingBySessionAndAudio 根据session_id和audio_id获取音频录音信息
func (a *AudioRecordingController) GetAudioRecordingBySessionAndAudio(ctx context.Context, sessionID, audioID string) (*types.AudioRecordingInfo, error) {
	recording, err := a.factory.AudioRecording().GetBySessionIDAndAudioID(ctx, sessionID, audioID)
	if err != nil {
		klog.Errorf("Failed to get audio recording session_id=%s audio_id=%s: %v", sessionID, audioID, err)
		return nil, errors.ErrNotFound
	}

	return &types.AudioRecordingInfo{
		ID:         recording.ID,
		SessionID:  recording.SessionID,
		AudioID:    recording.AudioID,
		MacAddress: recording.MacAddress,
		FilePath:   recording.FilePath,
		FileSize:   recording.FileSize,
		UploadTime: recording.UploadTime,
		Status:     recording.Status,
		CreatedAt:  recording.CreatedAt,
		UpdatedAt:  recording.UpdatedAt,
	}, nil
}

// ListAudioRecordings 获取音频录音列表
func (a *AudioRecordingController) ListAudioRecordings(ctx context.Context, req *types.AudioRecordingListRequest) (*types.AudioRecordingListResponse, error) {
	// 构建数据库查询请求
	daoReq := db.AudioRecordingListRequest{
		SessionID:  req.SessionID,
		MacAddress: req.MacAddress,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		Status:     req.Status,
		Offset:     req.Offset,
		Limit:      req.Limit,
	}

	// 获取总数
	total, err := a.factory.AudioRecording().Count(ctx, daoReq)
	if err != nil {
		return nil, errors.ErrServerInternal
	}

	// 获取列表数据
	recordings, err := a.factory.AudioRecording().List(ctx, daoReq)
	if err != nil {
		return nil, errors.ErrServerInternal
	}

	// 转换为响应格式
	items := make([]*types.AudioRecordingInfo, 0, len(recordings))
	for _, recording := range recordings {
		// 生成播放链接
		playURL := a.generatePlayURL(recording.SessionID, recording.AudioID)

		items = append(items, &types.AudioRecordingInfo{
			ID:         recording.ID,
			SessionID:  recording.SessionID,
			AudioID:    recording.AudioID,
			MacAddress: recording.MacAddress,
			FilePath:   recording.FilePath,
			FileSize:   recording.FileSize,
			UploadTime: recording.UploadTime,
			Status:     recording.Status,
			CreatedAt:  recording.CreatedAt,
			UpdatedAt:  recording.UpdatedAt,
			PlayURL:    playURL,
		})
	}

	return &types.AudioRecordingListResponse{
		Total: total,
		Items: items,
	}, nil
}

// DownloadAudioRecording 下载音频录音
func (a *AudioRecordingController) DownloadAudioRecording(ctx context.Context, id string) (io.ReadCloser, *types.AudioRecordingInfo, error) {
	// 获取录音信息
	recordingInfo, err := a.GetAudioRecording(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	// 从MinIO下载文件
	reader, err := a.minioClient.DownloadFile(ctx, "", recordingInfo.FilePath)
	if err != nil {
		klog.Errorf("Failed to download audio file from MinIO: %v", err)
		return nil, nil, errors.ErrServerInternal
	}

	return reader, recordingInfo, nil
}

// PlayAudioRecording 直接播放音频录音（通过session_id和audio_id）
func (a *AudioRecordingController) PlayAudioRecording(ctx context.Context, sessionID, audioID string) (io.ReadCloser, *types.AudioRecordingInfo, error) {
	// 获取录音信息
	recordingInfo, err := a.GetAudioRecordingBySessionAndAudio(ctx, sessionID, audioID)
	if err != nil {
		return nil, nil, err
	}

	// 从MinIO下载文件
	reader, err := a.minioClient.DownloadFile(ctx, "", recordingInfo.FilePath)
	if err != nil {
		klog.Errorf("Failed to download audio file from MinIO: %v", err)
		return nil, nil, errors.ErrServerInternal
	}

	return reader, recordingInfo, nil
}

// DeleteAudioRecording 删除音频录音
func (a *AudioRecordingController) DeleteAudioRecording(ctx context.Context, id string) error {
	// 获取录音信息
	recording, err := a.factory.AudioRecording().GetByID(ctx, id)
	if err != nil {
		return errors.ErrNotFound
	}

	// 从MinIO删除文件
	if err := a.minioClient.DeleteFile(ctx, "", recording.FilePath); err != nil {
		klog.Errorf("Failed to delete audio file from MinIO: %v", err)
		// 继续执行数据库删除，避免数据不一致
	}

	// 从数据库软删除
	if err := a.factory.AudioRecording().Delete(ctx, id); err != nil {
		klog.Errorf("Failed to delete audio recording record from database: %v", err)
		return errors.ErrServerInternal
	}

	return nil
}

// validateAudioFile 验证音频文件
func (a *AudioRecordingController) validateAudioFile(fileHeader *types.FileHeader) error {
	// 检查文件大小
	if fileHeader.Size > a.config.OSS.Processing.MaxFileSize {
		return errors.ErrFileSizeExceeded
	}

	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".wav" {
		return errors.ErrInvalidRequest
	}

	return nil
}

// validateUploadRequest 验证上传请求
func (a *AudioRecordingController) validateUploadRequest(req *types.AudioRecordingUploadRequest) error {
	// 验证session_id格式
	if len(req.SessionID) == 0 {
		return errors.ErrInvalidRequest
	}

	// 验证audio_id格式
	if len(req.AudioID) == 0 {
		return errors.ErrInvalidRequest
	}

	// 验证MAC地址格式 (简单验证)
	// if len(req.MacAddress) != 17 {
	// 	return errors.ErrInvalidRequest
	// }

	return nil
}

// generateAudioPath 生成音频文件存储路径
func (a *AudioRecordingController) generateAudioPath(sessionID, audioID string) string {
	// 生成路径: recordings/data/session_id-audio_id.wav
	path := fmt.Sprintf("recordings/data/%s-%s.wav", sessionID, audioID)
	return path
}

// calculateChecksum 计算文件校验和
func (a *AudioRecordingController) calculateChecksum(file io.Reader) (string, error) {
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// calculateChecksumFromBytes 从字节数组计算文件校验和
func (a *AudioRecordingController) calculateChecksumFromBytes(data []byte) (string, error) {
	hash := md5.New()
	if _, err := hash.Write(data); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// GetMinIOClient 获取MinIO客户端
func (a *AudioRecordingController) GetMinIOClient() minio.Client {
	return a.minioClient
}

// generatePlayURL 生成播放链接
func (a *AudioRecordingController) generatePlayURL(sessionID, audioID string) string {
	baseURL := a.config.Default.BaseURL
	if baseURL == "" {
		baseURL = "http://127.0.0.1:3008" // 默认值
	}
	return fmt.Sprintf("%s/api/v1/audio/session/%s/audio/%s/play", baseURL, sessionID, audioID)
}
