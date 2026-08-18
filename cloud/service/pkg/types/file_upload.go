package types

import (
	"io"
	"net/http"
)

// FileHeader 文件头信息
type FileHeader struct {
	Filename string
	Size     int64
	File     io.Reader
	Header   http.Header
}

// FileUploadRequest 文件上传请求
type FileUploadRequest struct {
	Uploader string `form:"uploader"` // 上传者，可选
}

// FileUploadResponse 文件上传响应
type FileUploadResponse struct {
	FileID       int64  `json:"file_id"`
	FileName     string `json:"file_name"`
	OriginalName string `json:"original_name"`
	FileSize     int64  `json:"file_size"`
	ContentType  string `json:"content_type"`
	MinIOPath    string `json:"minio_path"`
	DownloadURL  string `json:"download_url"`
	UploadedAt   int64  `json:"uploaded_at"`
}

// FileListRequest 文件列表请求
type FileListRequest struct {
	Uploader string `form:"uploader"`
	Status   *int8  `form:"status"`
	Offset   int    `form:"offset"`
	Limit    int    `form:"limit"`
}

// FileListResponse 文件列表响应
type FileListResponse struct {
	Total int64       `json:"total"`
	Items []*FileInfo `json:"items"`
}

// FileInfo 文件信息
type FileInfo struct {
	ID           int64  `json:"id"`
	FileName     string `json:"file_name"`
	OriginalName string `json:"original_name"`
	FileSize     int64  `json:"file_size"`
	ContentType  string `json:"content_type"`
	MinIOPath    string `json:"minio_path"`
	Checksum     string `json:"checksum"`
	Uploader     string `json:"uploader"`
	Status       int8   `json:"status"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}

// FileDownloadRequest 文件下载请求
type FileDownloadRequest struct {
	FileID int64 `uri:"file_id" binding:"required"`
}

// FileDeleteRequest 文件删除请求
type FileDeleteRequest struct {
	FileID int64 `uri:"file_id" binding:"required"`
}
