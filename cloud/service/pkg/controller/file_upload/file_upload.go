package file_upload

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/config"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/plugins/minio"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/types"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/util/errors"
	"github.com/google/uuid"
	"k8s.io/klog/v2"
)

// FileUploadController 文件上传控制器
type FileUploadController struct {
	config      config.Config
	factory     db.ShareDaoFactory
	minioClient minio.Client
}

// NewFileUploadController 创建文件上传控制器
func NewFileUploadController(cfg config.Config, factory db.ShareDaoFactory, minioClient minio.Client) *FileUploadController {
	return &FileUploadController{
		config:      cfg,
		factory:     factory,
		minioClient: minioClient,
	}
}

// UploadFile 上传文件
func (f *FileUploadController) UploadFile(ctx context.Context, fileHeader *types.FileHeader, uploader string) (*types.FileUploadResponse, error) {
	// 1. 验证文件
	if err := f.validateFile(fileHeader); err != nil {
		return nil, err
	}

	// 2. 生成文件信息
	fileID := uuid.New().String()
	fileName := fmt.Sprintf("%s%s", fileID, filepath.Ext(fileHeader.Filename))
	originalName := fileHeader.Filename
	fileSize := fileHeader.Size
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// 3. 生成MinIO存储路径
	minioPath := f.generateMinIOPath(fileName)

	// 4. 计算文件校验和
	checksum, err := f.calculateChecksum(fileHeader.File)
	if err != nil {
		klog.Errorf("Failed to calculate file checksum: %v", err)
		return nil, errors.ErrServerInternal
	}

	// 5. 上传到MinIO
	if err := f.minioClient.UploadFile(ctx, "", minioPath, fileHeader.File, fileSize, contentType); err != nil {
		klog.Errorf("Failed to upload file to MinIO: %v", err)
		return nil, errors.ErrMinIOUploadFailed
	}

	// 6. 保存到数据库
	fileRecord := &model.FileUpload{
		FileName:     fileName,
		OriginalName: originalName,
		FileSize:     fileSize,
		ContentType:  contentType,
		MinIOPath:    minioPath,
		Checksum:     checksum,
		Uploader:     uploader,
		Status:       model.FileStatusNormal,
	}

	createdFile, err := f.factory.FileUpload().Create(ctx, fileRecord)
	if err != nil {
		// 回滚MinIO文件
		if deleteErr := f.minioClient.DeleteFile(ctx, "", minioPath); deleteErr != nil {
			klog.Errorf("Failed to rollback MinIO file %s: %v", minioPath, deleteErr)
		}
		klog.Errorf("Failed to save file record to database: %v", err)
		return nil, errors.ErrServerInternal
	}

	// 7. 生成下载URL
	downloadURL := fmt.Sprintf("/api/v1/oss/download/%d", createdFile.ID)

	return &types.FileUploadResponse{
		FileID:       createdFile.ID,
		FileName:     createdFile.FileName,
		OriginalName: createdFile.OriginalName,
		FileSize:     createdFile.FileSize,
		ContentType:  createdFile.ContentType,
		MinIOPath:    createdFile.MinIOPath,
		DownloadURL:  downloadURL,
		UploadedAt:   createdFile.CreatedAt,
	}, nil
}

// GetFile 获取文件信息
func (f *FileUploadController) GetFile(ctx context.Context, fileID int64) (*types.FileInfo, error) {
	file, err := f.factory.FileUpload().GetByID(ctx, fileID)
	if err != nil {
		klog.Errorf("Failed to get file %d: %v", fileID, err)
		return nil, errors.ErrNotFound
	}

	return &types.FileInfo{
		ID:           file.ID,
		FileName:     file.FileName,
		OriginalName: file.OriginalName,
		FileSize:     file.FileSize,
		ContentType:  file.ContentType,
		MinIOPath:    file.MinIOPath,
		Checksum:     file.Checksum,
		Uploader:     file.Uploader,
		Status:       file.Status,
		CreatedAt:    file.CreatedAt,
		UpdatedAt:    file.UpdatedAt,
	}, nil
}

// ListFiles 获取文件列表
func (f *FileUploadController) ListFiles(ctx context.Context, req *types.FileListRequest) (*types.FileListResponse, error) {
	// 构建数据库查询请求
	daoReq := db.FileUploadListRequest{
		Uploader: req.Uploader,
		Status:   req.Status,
		Offset:   req.Offset,
		Limit:    req.Limit,
	}

	// 获取总数
	total, err := f.factory.FileUpload().Count(ctx, daoReq)
	if err != nil {
		return nil, errors.ErrServerInternal
	}

	// 获取列表数据
	files, err := f.factory.FileUpload().List(ctx, daoReq)
	if err != nil {
		return nil, errors.ErrServerInternal
	}

	// 转换为响应格式
	items := make([]*types.FileInfo, 0, len(files))
	for _, file := range files {
		items = append(items, &types.FileInfo{
			ID:           file.ID,
			FileName:     file.FileName,
			OriginalName: file.OriginalName,
			FileSize:     file.FileSize,
			ContentType:  file.ContentType,
			MinIOPath:    file.MinIOPath,
			Checksum:     file.Checksum,
			Uploader:     file.Uploader,
			Status:       file.Status,
			CreatedAt:    file.CreatedAt,
			UpdatedAt:    file.UpdatedAt,
		})
	}

	return &types.FileListResponse{
		Total: total,
		Items: items,
	}, nil
}

// DownloadFile 下载文件
func (f *FileUploadController) DownloadFile(ctx context.Context, fileID int64) (io.ReadCloser, *types.FileInfo, error) {
	// 获取文件信息
	fileInfo, err := f.GetFile(ctx, fileID)
	if err != nil {
		return nil, nil, err
	}

	// 从MinIO下载文件
	reader, err := f.minioClient.DownloadFile(ctx, "", fileInfo.MinIOPath)
	if err != nil {
		klog.Errorf("Failed to download file from MinIO: %v", err)
		return nil, nil, errors.ErrServerInternal
	}

	return reader, fileInfo, nil
}

// DeleteFile 删除文件
func (f *FileUploadController) DeleteFile(ctx context.Context, fileID int64) error {
	// 获取文件信息
	file, err := f.factory.FileUpload().GetByID(ctx, fileID)
	if err != nil {
		return errors.ErrNotFound
	}

	// 从MinIO删除文件
	if err := f.minioClient.DeleteFile(ctx, "", file.MinIOPath); err != nil {
		klog.Errorf("Failed to delete file from MinIO: %v", err)
		// 继续执行数据库删除，避免数据不一致
	}

	// 从数据库软删除
	if err := f.factory.FileUpload().Delete(ctx, fileID); err != nil {
		klog.Errorf("Failed to delete file record from database: %v", err)
		return errors.ErrServerInternal
	}

	return nil
}

// validateFile 验证文件
func (f *FileUploadController) validateFile(fileHeader *types.FileHeader) error {
	// 检查文件大小
	if fileHeader.Size > f.config.OSS.Processing.MaxFileSize {
		return errors.ErrFileSizeExceeded
	}

	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext == "" {
		return errors.ErrInvalidRequest
	}

	return nil
}

// generateMinIOPath 生成MinIO存储路径
func (f *FileUploadController) generateMinIOPath(fileName string) string {
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")

	// 生成路径: files/2024/01/15/uuid-filename.ext
	path := fmt.Sprintf("files/%s/%s/%s/%s", year, month, day, fileName)

	return path
}

// calculateChecksum 计算文件校验和
func (f *FileUploadController) calculateChecksum(file io.Reader) (string, error) {
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
