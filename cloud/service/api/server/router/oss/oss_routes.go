package oss

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/httputils"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/types"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/util/errors"
)

type ossRouter struct {
	c controller.SensecraftVoiceInterface
}

func newOSSRouter(c controller.SensecraftVoiceInterface) *ossRouter {
	return &ossRouter{c: c}
}

// RegisterRoutes 注册OSS相关路由
func (r *ossRouter) RegisterRoutes(router *gin.RouterGroup) {
	oss := router.Group("/oss")
	{
		// 文件上传
		oss.POST("/upload", r.uploadFile)

		// 文件管理
		oss.GET("/files", r.listFiles)
		oss.GET("/file/:file_id", r.getFile)
		oss.DELETE("/file/:file_id", r.deleteFile)

		// 文件下载
		oss.GET("/download/:file_id", r.downloadFile)
	}
}

// uploadFile 上传文件
func (r *ossRouter) uploadFile(c *gin.Context) {
	resp := httputils.NewResponse()

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		klog.Errorf("Failed to get uploaded file: %v", err)
		httputils.SetFailed(c, resp, errors.ErrInvalidRequest)
		return
	}

	// 获取上传者信息
	uploader := c.PostForm("uploader")

	// 打开文件
	src, err := file.Open()
	if err != nil {
		klog.Errorf("Failed to open uploaded file: %v", err)
		httputils.SetFailed(c, resp, errors.ErrInvalidRequest)
		return
	}
	defer src.Close()

	// 创建文件头信息
	fileHeader := &types.FileHeader{
		Filename: file.Filename,
		Size:     file.Size,
		File:     src,
		Header:   make(map[string][]string),
	}

	klog.Infof("Upload file request: filename=%s, size=%d, uploader=%s",
		file.Filename, file.Size, uploader)

	result, err := r.c.FileUpload().UploadFile(c, fileHeader, uploader)
	if err != nil {
		klog.Errorf("Failed to upload file: %v", err)
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = result
	httputils.SetSuccess(c, resp)
}

// getFile 获取文件信息
func (r *ossRouter) getFile(c *gin.Context) {
	resp := httputils.NewResponse()

	fileIDStr := c.Param("file_id")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		httputils.SetFailed(c, resp, errors.ErrInvalidRequest)
		return
	}

	klog.Infof("Get file request: file_id=%d", fileID)

	result, err := r.c.FileUpload().GetFile(c, fileID)
	if err != nil {
		klog.Errorf("Failed to get file: %v", err)
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = result
	httputils.SetSuccess(c, resp)
}

// listFiles 获取文件列表
func (r *ossRouter) listFiles(c *gin.Context) {
	resp := httputils.NewResponse()

	var req types.FileListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		klog.Errorf("Failed to bind file list request: %v", err)
		httputils.SetFailed(c, resp, err)
		return
	}

	// 设置默认分页参数
	if req.Offset < 0 {
		req.Offset = 0
	}
	if req.Limit <= 0 || req.Limit > 1000 {
		req.Limit = 50
	}

	klog.Infof("List files request: uploader=%s, offset=%d, limit=%d",
		req.Uploader, req.Offset, req.Limit)

	result, err := r.c.FileUpload().ListFiles(c, &req)
	if err != nil {
		klog.Errorf("Failed to list files: %v", err)
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = result
	httputils.SetSuccess(c, resp)
}

// downloadFile 下载文件
func (r *ossRouter) downloadFile(c *gin.Context) {
	fileIDStr := c.Param("file_id")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file id"})
		return
	}

	klog.Infof("Download file request: file_id=%d", fileID)

	reader, fileInfo, err := r.c.FileUpload().DownloadFile(c, fileID)
	if err != nil {
		klog.Errorf("Failed to download file: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}
	defer reader.Close()

	// 设置响应头
	c.Header("Content-Disposition", "attachment; filename="+fileInfo.OriginalName)
	c.Header("Content-Type", fileInfo.ContentType)
	c.Header("Content-Length", strconv.FormatInt(fileInfo.FileSize, 10))

	// 流式传输文件内容
	_, err = io.Copy(c.Writer, reader)
	if err != nil {
		klog.Errorf("Failed to stream file content: %v", err)
		return
	}
}

// deleteFile 删除文件
func (r *ossRouter) deleteFile(c *gin.Context) {
	resp := httputils.NewResponse()

	fileIDStr := c.Param("file_id")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		httputils.SetFailed(c, resp, errors.ErrInvalidRequest)
		return
	}

	klog.Infof("Delete file request: file_id=%d", fileID)

	err = r.c.FileUpload().DeleteFile(c, fileID)
	if err != nil {
		klog.Errorf("Failed to delete file: %v", err)
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = gin.H{"message": "file deleted successfully"}
	httputils.SetSuccess(c, resp)
}
