package audio_recordings

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/httputils"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/types"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/util/errors"
)

type AudioRecordingMeta struct {
	ID string `uri:"id" binding:"required"`
}

type AudioRecordingSessionAudioMeta struct {
	SessionID string `uri:"session_id" binding:"required"`
	AudioID   string `uri:"audio_id" binding:"required"`
}

func (a *audioRecordingsRouter) uploadAudioRecording(c *gin.Context) {
	r := httputils.NewResponse()

	// 设置请求体大小限制，防止恶意大文件上传
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10<<20) // 10MB

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		klog.Errorf("Failed to get uploaded file: %v", err)
		if err.Error() == "http: request body too large" {
			httputils.SetFailed(c, r, errors.ErrFileSizeExceeded)
			return
		}
		httputils.SetFailed(c, r, errors.ErrInvalidRequest)
		return
	}

	// 获取上传参数
	var req types.AudioRecordingUploadRequest
	if err := c.ShouldBind(&req); err != nil {
		klog.Errorf("Failed to bind upload request: %v", err)
		httputils.SetFailed(c, r, err)
		return
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		klog.Errorf("Failed to open uploaded file: %v", err)
		httputils.SetFailed(c, r, err)
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

	klog.Infof("Upload audio recording request: filename=%s, size=%d, session_id=%s, audio_id=%s, mac_address=%s",
		file.Filename, file.Size, req.SessionID, req.AudioID, req.MacAddress)

	if r.Result, err = a.c.AudioRecording().UploadAudioRecording(c, fileHeader, &req); err != nil {
		klog.Errorf("Failed to upload audio recording: %v", err)
		httputils.SetFailed(c, r, err)
		return
	}

	httputils.SetSuccess(c, r)
}

func (a *audioRecordingsRouter) getAudioRecording(c *gin.Context) {
	r := httputils.NewResponse()

	var (
		opt AudioRecordingMeta
		err error
	)
	if err = c.ShouldBindUri(&opt); err != nil {
		httputils.SetFailed(c, r, err)
		return
	}

	klog.Infof("Get audio recording request: id=%s", opt.ID)

	if r.Result, err = a.c.AudioRecording().GetAudioRecording(c, opt.ID); err != nil {
		klog.Errorf("Failed to get audio recording: %v", err)
		httputils.SetFailed(c, r, err)
		return
	}

	httputils.SetSuccess(c, r)
}

func (a *audioRecordingsRouter) getAudioRecordingBySessionAndAudio(c *gin.Context) {
	r := httputils.NewResponse()

	var (
		opt AudioRecordingSessionAudioMeta
		err error
	)
	if err = c.ShouldBindUri(&opt); err != nil {
		httputils.SetFailed(c, r, err)
		return
	}

	klog.Infof("Get audio recording by session and audio request: session_id=%s, audio_id=%s", opt.SessionID, opt.AudioID)

	if r.Result, err = a.c.AudioRecording().GetAudioRecordingBySessionAndAudio(c, opt.SessionID, opt.AudioID); err != nil {
		klog.Errorf("Failed to get audio recording by session and audio: %v", err)
		httputils.SetFailed(c, r, err)
		return
	}

	httputils.SetSuccess(c, r)
}

func (a *audioRecordingsRouter) listAudioRecordings(c *gin.Context) {
	r := httputils.NewResponse()

	var (
		req types.AudioRecordingListRequest
		err error
	)
	if err = c.ShouldBindQuery(&req); err != nil {
		klog.Errorf("Failed to bind audio recording list request: %v", err)
		httputils.SetFailed(c, r, err)
		return
	}

	// 设置默认分页参数
	if req.Offset < 0 {
		req.Offset = 0
	}
	if req.Limit <= 0 || req.Limit > 1000 {
		req.Limit = 50
	}

	klog.Infof("List audio recordings request: session_id=%s, mac_address=%s, offset=%d, limit=%d",
		req.SessionID, req.MacAddress, req.Offset, req.Limit)

	if r.Result, err = a.c.AudioRecording().ListAudioRecordings(c, &req); err != nil {
		klog.Errorf("Failed to list audio recordings: %v", err)
		httputils.SetFailed(c, r, err)
		return
	}

	httputils.SetSuccess(c, r)
}

func (a *audioRecordingsRouter) downloadAudioRecording(c *gin.Context) {
	var (
		opt AudioRecordingMeta
		err error
	)
	if err = c.ShouldBindUri(&opt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	klog.Infof("Download audio recording request: id=%s", opt.ID)

	// 获取录音信息
	recordingInfo, err := a.c.AudioRecording().GetAudioRecording(c, opt.ID)
	if err != nil {
		klog.Errorf("Failed to get audio recording: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "audio recording not found"})
		return
	}

	// 直接从MinIO获取文件流
	reader, err := a.c.AudioRecording().GetMinIOClient().DownloadFile(c, "", recordingInfo.FilePath)
	if err != nil {
		klog.Errorf("Failed to download audio file from MinIO: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to download audio file"})
		return
	}
	defer reader.Close()

	// 设置响应头
	filename := recordingInfo.SessionID + "-" + recordingInfo.AudioID + ".wav"
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "audio/wav")
	c.Header("Content-Length", strconv.FormatInt(recordingInfo.FileSize, 10))
	c.Header("Accept-Ranges", "bytes")

	// 直接代理文件内容
	_, err = io.Copy(c.Writer, reader)
	if err != nil {
		// 只记录真正的错误，客户端取消连接是正常行为
		if err != context.Canceled && err != context.DeadlineExceeded {
			klog.Errorf("Failed to proxy audio recording content: %v", err)
		}
		return
	}
}

func (a *audioRecordingsRouter) deleteAudioRecording(c *gin.Context) {
	r := httputils.NewResponse()

	var (
		opt AudioRecordingMeta
		err error
	)
	if err = c.ShouldBindUri(&opt); err != nil {
		httputils.SetFailed(c, r, err)
		return
	}

	klog.Infof("Delete audio recording request: id=%s", opt.ID)

	if err = a.c.AudioRecording().DeleteAudioRecording(c, opt.ID); err != nil {
		klog.Errorf("Failed to delete audio recording: %v", err)
		httputils.SetFailed(c, r, err)
		return
	}

	r.Result = gin.H{"message": "audio recording deleted successfully"}
	httputils.SetSuccess(c, r)
}

func (a *audioRecordingsRouter) playAudioRecording(c *gin.Context) {
	var (
		opt AudioRecordingSessionAudioMeta
		err error
	)
	if err = c.ShouldBindUri(&opt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session_id or audio_id"})
		return
	}

	klog.Infof("Play audio recording request: session_id=%s, audio_id=%s", opt.SessionID, opt.AudioID)

	// 获取录音信息
	recordingInfo, err := a.c.AudioRecording().GetAudioRecordingBySessionAndAudio(c, opt.SessionID, opt.AudioID)
	if err != nil {
		klog.Errorf("Failed to get audio recording: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "audio recording not found"})
		return
	}

	// 直接从MinIO获取文件流
	reader, err := a.c.AudioRecording().GetMinIOClient().DownloadFile(c, "", recordingInfo.FilePath)
	if err != nil {
		klog.Errorf("Failed to download audio file from MinIO: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to download audio file"})
		return
	}
	defer reader.Close()

	// 设置响应头用于直接播放
	filename := recordingInfo.SessionID + "-" + recordingInfo.AudioID + ".wav"
	c.Header("Content-Disposition", "inline; filename="+filename)
	c.Header("Content-Type", "audio/wav")
	c.Header("Content-Length", strconv.FormatInt(recordingInfo.FileSize, 10))
	c.Header("Accept-Ranges", "bytes")
	c.Header("Cache-Control", "public, max-age=3600") // 缓存1小时

	// 直接代理文件内容，让gin框架处理所有细节
	_, err = io.Copy(c.Writer, reader)
	if err != nil {
		// 只记录真正的错误，客户端取消连接是正常行为
		if err != context.Canceled && err != context.DeadlineExceeded {
			klog.Errorf("Failed to proxy audio recording content: %v", err)
		}
		return
	}
}
