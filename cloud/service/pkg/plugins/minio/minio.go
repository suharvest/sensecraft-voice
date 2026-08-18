package minio

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"k8s.io/klog/v2"
)

// Client MinIO客户端接口
type Client interface {
	// 上传文件
	UploadFile(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, contentType string) error
	// 下载文件
	DownloadFile(ctx context.Context, bucketName, objectName string) (io.ReadCloser, error)
	// 删除文件
	DeleteFile(ctx context.Context, bucketName, objectName string) error
	// 检查文件是否存在
	FileExists(ctx context.Context, bucketName, objectName string) (bool, error)
	// 获取文件信息
	GetFileInfo(ctx context.Context, bucketName, objectName string) (minio.ObjectInfo, error)
	// 关闭连接
	Close() error
}

// minioClient MinIO客户端实现
type minioClient struct {
	client     *minio.Client
	bucketName string
	timeout    time.Duration
}

// NewClient 创建MinIO客户端
func NewClient(cfg *config.MinIOOptions) (Client, error) {
	if cfg.Endpoint == "" {
		return nil, errors.New("minio endpoint is required")
	}
	if cfg.AccessKey == "" {
		return nil, errors.New("minio access_key is required")
	}
	if cfg.SecretKey == "" {
		return nil, errors.New("minio secret_key is required")
	}
	if cfg.BucketName == "" {
		return nil, errors.New("minio bucket_name is required")
	}

	// 创建MinIO客户端
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %v", err)
	}

	// 设置超时
	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	mc := &minioClient{
		client:     client,
		bucketName: cfg.BucketName,
		timeout:    timeout,
	}

	// 确保bucket存在
	if err := mc.ensureBucketExists(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %v", err)
	}

	return mc, nil
}

// ensureBucketExists 确保bucket存在
func (mc *minioClient) ensureBucketExists(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, mc.timeout)
	defer cancel()

	exists, err := mc.client.BucketExists(ctx, mc.bucketName)
	if err != nil {
		return err
	}

	if !exists {
		err = mc.client.MakeBucket(ctx, mc.bucketName, minio.MakeBucketOptions{
			Region: "us-east-1", // 使用默认区域
		})
		if err != nil {
			return err
		}
		klog.Infof("Created bucket: %s", mc.bucketName)
	}

	return nil
}

// UploadFile 上传文件
func (mc *minioClient) UploadFile(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, contentType string) error {
	ctx, cancel := context.WithTimeout(ctx, mc.timeout)
	defer cancel()

	// 如果bucketName为空，使用默认bucket
	if bucketName == "" {
		bucketName = mc.bucketName
	}

	// 上传文件
	_, err := mc.client.PutObject(ctx, bucketName, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file %s: %v", objectName, err)
	}

	klog.V(2).Infof("Successfully uploaded file: %s/%s", bucketName, objectName)
	return nil
}

// DownloadFile 下载文件
func (mc *minioClient) DownloadFile(ctx context.Context, bucketName, objectName string) (io.ReadCloser, error) {
	// 如果bucketName为空，使用默认bucket
	if bucketName == "" {
		bucketName = mc.bucketName
	}

	// 直接使用原始context，不设置超时
	// 让调用方控制超时
	object, err := mc.client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download file %s: %v", objectName, err)
	}

	return object, nil
}

// DeleteFile 删除文件
func (mc *minioClient) DeleteFile(ctx context.Context, bucketName, objectName string) error {
	ctx, cancel := context.WithTimeout(ctx, mc.timeout)
	defer cancel()

	// 如果bucketName为空，使用默认bucket
	if bucketName == "" {
		bucketName = mc.bucketName
	}

	err := mc.client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file %s: %v", objectName, err)
	}

	klog.V(2).Infof("Successfully deleted file: %s/%s", bucketName, objectName)
	return nil
}

// FileExists 检查文件是否存在
func (mc *minioClient) FileExists(ctx context.Context, bucketName, objectName string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, mc.timeout)
	defer cancel()

	// 如果bucketName为空，使用默认bucket
	if bucketName == "" {
		bucketName = mc.bucketName
	}

	_, err := mc.client.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GetFileInfo 获取文件信息
func (mc *minioClient) GetFileInfo(ctx context.Context, bucketName, objectName string) (minio.ObjectInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, mc.timeout)
	defer cancel()

	// 如果bucketName为空，使用默认bucket
	if bucketName == "" {
		bucketName = mc.bucketName
	}

	info, err := mc.client.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return minio.ObjectInfo{}, fmt.Errorf("failed to get file info %s: %v", objectName, err)
	}

	return info, nil
}

// Close 关闭连接
func (mc *minioClient) Close() error {
	// MinIO客户端不需要显式关闭
	return nil
}

// GenerateAudioPath 生成音频文件存储路径
func GenerateAudioPath(deviceID string, sessionID string, chunkIndex int, format string) string {
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")

	// 格式化设备ID，将冒号替换为横线
	formattedDeviceID := strings.ReplaceAll(deviceID, ":", "-")

	// 生成路径: audio/2024/01/15/AA-BB-CC-DD-EE-FF/session_001/chunk_001.wav
	path := fmt.Sprintf("audio/%s/%s/%s/%s/session_%s/chunk_%03d.%s",
		year, month, day, formattedDeviceID, sessionID, chunkIndex, format)

	return path
}

// GenerateSessionMetadataPath 生成会话元数据文件路径
func GenerateSessionMetadataPath(deviceID string, sessionID string) string {
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")

	// 格式化设备ID
	formattedDeviceID := strings.ReplaceAll(deviceID, ":", "-")

	// 生成路径: audio/2024/01/15/AA-BB-CC-DD-EE-FF/session_001/session_metadata.json
	path := fmt.Sprintf("audio/%s/%s/%s/%s/session_%s/session_metadata.json",
		year, month, day, formattedDeviceID, sessionID)

	return path
}

// GetContentType 根据文件扩展名获取Content-Type
func GetContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".wav":
		return "audio/wav"
	case ".mp3":
		return "audio/mpeg"
	case ".pcm":
		return "audio/pcm"
	case ".json":
		return "application/json"
	default:
		return "application/octet-stream"
	}
}
