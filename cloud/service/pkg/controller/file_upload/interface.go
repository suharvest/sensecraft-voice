package file_upload

import (
	"context"
	"io"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/types"
)

// FileUploadGetter 文件上传控制器获取器
type FileUploadGetter interface {
	FileUpload() Interface
}

// Interface 文件上传控制器接口
type Interface interface {
	// 上传文件
	UploadFile(ctx context.Context, fileHeader *types.FileHeader, uploader string) (*types.FileUploadResponse, error)
	// 获取文件信息
	GetFile(ctx context.Context, fileID int64) (*types.FileInfo, error)
	// 获取文件列表
	ListFiles(ctx context.Context, req *types.FileListRequest) (*types.FileListResponse, error)
	// 下载文件
	DownloadFile(ctx context.Context, fileID int64) (io.ReadCloser, *types.FileInfo, error)
	// 删除文件
	DeleteFile(ctx context.Context, fileID int64) error
}
