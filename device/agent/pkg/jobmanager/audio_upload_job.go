package jobmanager

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// AudioFile 音频文件信息
type AudioFile struct {
	Path      string    // 文件完整路径
	SessionID string    // 从文件名解析的session_id
	AudioID   string    // 从文件名解析的audio_id
	Size      int64     // 文件大小
	ModTime   time.Time // 文件修改时间
}

// AudioUploadConfig 音频上传配置
type AudioUploadConfig struct {
	Enabled       bool   `yaml:"enabled"`
	ScanDir       string `yaml:"scan_dir"`
	MacAddress    string `yaml:"mac_address"`
	Timeout       string `yaml:"timeout"`
	MaxFileSize   int64  `yaml:"max_file_size"`
	MaxConcurrent int    `yaml:"max_concurrent"`
}

// AudioUploadJob 音频文件上传任务
type AudioUploadJob struct {
	config     AudioUploadConfig
	client     *http.Client
	macAddress string
	getBaseURL func() string
}

// NewAudioUploadJob 创建新的音频上传任务
func NewAudioUploadJob(config AudioUploadConfig) *AudioUploadJob {
	timeout, err := time.ParseDuration(config.Timeout)
	if err != nil {
		timeout = 30 * time.Second
	}

	// 自动获取MAC地址
	macAddress := config.MacAddress
	if macAddress == "" {
		macAddress = getMacAddress()
	}

	return &AudioUploadJob{
		config:     config,
		client:     &http.Client{Timeout: timeout},
		macAddress: macAddress,
		getBaseURL: func() string { return "" }, // 默认返回空字符串
	}
}

// NewAudioUploadJobWithBaseURL 创建带baseURL获取函数的音频上传任务
func NewAudioUploadJobWithBaseURL(config AudioUploadConfig, getBaseURL func() string) *AudioUploadJob {
	timeout, err := time.ParseDuration(config.Timeout)
	if err != nil {
		timeout = 30 * time.Second
	}

	// 自动获取MAC地址
	macAddress := config.MacAddress
	if macAddress == "" {
		macAddress = getMacAddress()
	}

	return &AudioUploadJob{
		config:     config,
		client:     &http.Client{Timeout: timeout},
		macAddress: macAddress,
		getBaseURL: getBaseURL,
	}
}

// Name 返回任务名称
func (j *AudioUploadJob) Name() string {
	return "audio-upload-job"
}

// CronSpec 返回cron表达式
func (j *AudioUploadJob) CronSpec() string {
	return "@every 60s"
}

// Do 执行音频上传任务
func (j *AudioUploadJob) Do(ctx *JobContext) error {
	if !j.config.Enabled {
		return nil
	}

	// 扫描音频文件
	files := j.scanAudioFiles(ctx)
	if len(files) == 0 {
		return nil
	}

	ctx.WithLogFields(map[string]interface{}{
		"files_found": len(files),
	})

	// 处理文件上传
	successCount := 0
	failedCount := 0

	for _, file := range files {
		// 检查文件大小限制
		if file.Size > j.config.MaxFileSize*1024*1024 {
			ctx.WithLogFields(map[string]interface{}{
				"file":  file.Path,
				"size":  file.Size,
				"error": "file too large",
			})
			failedCount++
			continue
		}

		// 尝试上传文件
		if j.uploadFile(ctx, file) {
			// 成功：删除文件
			if err := os.Remove(file.Path); err != nil {
				ctx.WithLogFields(map[string]interface{}{
					"file":  file.Path,
					"error": "failed to delete file after upload",
				})
			} else {
				ctx.WithLogFields(map[string]interface{}{
					"file":   file.Path,
					"action": "uploaded and deleted",
				})
			}
			successCount++
		} else {
			// 失败：保留文件
			ctx.WithLogFields(map[string]interface{}{
				"file":   file.Path,
				"action": "upload failed, file preserved",
			})
			failedCount++
		}
	}

	ctx.WithLogFields(map[string]interface{}{
		"success_count": successCount,
		"failed_count":  failedCount,
	})

	return nil
}

// scanAudioFiles 扫描音频文件
func (j *AudioUploadJob) scanAudioFiles(ctx *JobContext) []AudioFile {
	var files []AudioFile

	// 确保扫描目录存在
	if _, err := os.Stat(j.config.ScanDir); os.IsNotExist(err) {
		ctx.WithLogFields(map[string]interface{}{
			"scan_dir": j.config.ScanDir,
			"error":    "directory does not exist",
		})
		return files
	}

	// 递归扫描目录
	err := filepath.Walk(j.config.ScanDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// 忽略单个文件的错误，继续扫描
			ctx.WithLogFields(map[string]interface{}{
				"file":  path,
				"error": err.Error(),
			})
			return nil
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 只处理 .wav 文件
		if !strings.HasSuffix(strings.ToLower(path), ".wav") {
			return nil
		}

		// 解析文件路径获取 session_id 和 audio_id
		sessionID, audioID := j.parseFilePath(path)
		if sessionID == "" || audioID == "" {
			ctx.WithLogFields(map[string]interface{}{
				"file":        path,
				"error":       "invalid filename format",
				"expected":    "sessionid-audioid.wav or sessionid/audioid.wav",
				"actual_name": info.Name(),
			})
			return nil
		}

		files = append(files, AudioFile{
			Path:      path,
			SessionID: sessionID,
			AudioID:   audioID,
			Size:      info.Size(),
			ModTime:   info.ModTime(),
		})

		return nil
	})

	if err != nil {
		ctx.WithLogFields(map[string]interface{}{
			"scan_dir": j.config.ScanDir,
			"error":    err.Error(),
		})
	}

	return files
}

// parseFilePath 解析文件路径获取session_id和audio_id
func (j *AudioUploadJob) parseFilePath(filePath string) (sessionID, audioID string) {
	// 获取相对于扫描目录的路径
	relPath, err := filepath.Rel(j.config.ScanDir, filePath)
	if err != nil {
		return "", ""
	}

	// 移除 .wav 扩展名
	relPath = strings.TrimSuffix(strings.ToLower(relPath), ".wav")

	// 分割路径
	pathParts := strings.Split(relPath, string(filepath.Separator))

	if len(pathParts) == 1 {
		// 格式1: sessionid-audioid.wav
		parts := strings.Split(pathParts[0], "-")
		if len(parts) >= 2 {
			sessionID = parts[0]
			audioID = strings.Join(parts[1:], "-")
		}
	} else if len(pathParts) == 2 {
		// 格式2: sessionid/audioid.wav
		sessionID = pathParts[0]
		audioID = pathParts[1]
	}

	return sessionID, audioID
}

// parseFileName 解析文件名获取session_id和audio_id（保留向后兼容）
func (j *AudioUploadJob) parseFileName(filename string) (sessionID, audioID string) {
	// 移除 .wav 扩展名
	name := strings.TrimSuffix(strings.ToLower(filename), ".wav")

	// 按 - 分割，期望格式：sessionid-audioid
	parts := strings.Split(name, "-")
	if len(parts) >= 2 {
		sessionID = parts[0]
		audioID = strings.Join(parts[1:], "-") // 支持audio_id中包含-
	}

	return sessionID, audioID
}

// uploadFile 上传单个文件
func (j *AudioUploadJob) uploadFile(ctx *JobContext, file AudioFile) bool {
	// 打开文件
	audioFile, err := os.Open(file.Path)
	if err != nil {
		ctx.WithLogFields(map[string]interface{}{
			"file":  file.Path,
			"error": "failed to open file",
		})
		return false
	}
	defer audioFile.Close()

	// 创建 multipart 请求体
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加文件
	fileWriter, err := writer.CreateFormFile("file", filepath.Base(file.Path))
	if err != nil {
		ctx.WithLogFields(map[string]interface{}{
			"file":  file.Path,
			"error": "failed to create form file",
		})
		return false
	}

	// 复制文件内容
	_, err = io.Copy(fileWriter, audioFile)
	if err != nil {
		ctx.WithLogFields(map[string]interface{}{
			"file":  file.Path,
			"error": "failed to copy file content",
		})
		return false
	}

	// 添加元数据字段
	writer.WriteField("session_id", file.SessionID)
	writer.WriteField("audio_id", file.AudioID)
	writer.WriteField("mac_address", j.macAddress)

	// 关闭writer
	err = writer.Close()
	if err != nil {
		ctx.WithLogFields(map[string]interface{}{
			"file":  file.Path,
			"error": "failed to close multipart writer",
		})
		return false
	}

	// 获取当前配置的baseURL
	baseURL := j.getBaseURL()
	if baseURL == "" {
		ctx.WithLogFields(map[string]interface{}{
			"file":  file.Path,
			"error": "远程服务地址未配置",
		})
		return false
	}

	// 创建HTTP请求
	uploadURL := baseURL + "/api/v1/audio/upload"
	req, err := http.NewRequestWithContext(context.Background(), "POST", uploadURL, body)
	if err != nil {
		ctx.WithLogFields(map[string]interface{}{
			"file":  file.Path,
			"error": "failed to create HTTP request",
		})
		return false
	}

	// 设置请求头
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", "SenseCraft-Voice-Client/1.0")

	// 发送请求
	resp, err := j.client.Do(req)
	if err != nil {
		ctx.WithLogFields(map[string]interface{}{
			"file":  file.Path,
			"error": "failed to send HTTP request",
		})
		return false
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		ctx.WithLogFields(map[string]interface{}{
			"file":        file.Path,
			"status_code": resp.StatusCode,
			"error":       "upload failed",
		})
		return false
	}

	// 读取响应体用于调试
	responseBody, _ := io.ReadAll(resp.Body)
	ctx.WithLogFields(map[string]interface{}{
		"file":          file.Path,
		"status_code":   resp.StatusCode,
		"response_body": string(responseBody),
	})

	return true
}

// GetStatus 获取任务状态
func (j *AudioUploadJob) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"job_name":       j.Name(),
		"cron_spec":      j.CronSpec(),
		"enabled":        j.config.Enabled,
		"scan_dir":       j.config.ScanDir,
		"base_url":       j.getBaseURL(),
		"max_file_size":  j.config.MaxFileSize,
		"max_concurrent": j.config.MaxConcurrent,
		"timeout":        j.config.Timeout,
	}
}

// getMacAddress 获取设备MAC地址
func getMacAddress() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "unknown"
	}

	for _, iface := range interfaces {
		// 跳过回环接口和无效接口
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		// 跳过虚拟接口
		if strings.Contains(iface.Name, "docker") ||
			strings.Contains(iface.Name, "veth") ||
			strings.Contains(iface.Name, "br-") {
			continue
		}

		// 获取MAC地址
		if len(iface.HardwareAddr) > 0 {
			return iface.HardwareAddr.String()
		}
	}

	return "unknown"
}

// UpdateConfig 更新配置
func (j *AudioUploadJob) UpdateConfig(config AudioUploadConfig) {
	j.config = config

	// 更新MAC地址
	if config.MacAddress != "" {
		j.macAddress = config.MacAddress
	} else {
		j.macAddress = getMacAddress()
	}

	// 更新HTTP客户端超时
	timeout, err := time.ParseDuration(config.Timeout)
	if err != nil {
		timeout = 30 * time.Second
	}
	j.client.Timeout = timeout
}

// UpdateBaseURLGetter 更新baseURL获取函数
func (j *AudioUploadJob) UpdateBaseURLGetter(getBaseURL func() string) {
	j.getBaseURL = getBaseURL
}
