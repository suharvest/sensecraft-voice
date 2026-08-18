package types

// AudioRecordingUploadRequest 音频录音上传请求
type AudioRecordingUploadRequest struct {
	SessionID  string `form:"session_id" binding:"required"`
	AudioID    string `form:"audio_id" binding:"required"`
	MacAddress string `form:"mac_address" binding:"required"`
}

// AudioRecordingUploadResponse 音频录音上传响应
type AudioRecordingUploadResponse struct {
	ID         string `json:"id"`
	SessionID  string `json:"session_id"`
	AudioID    string `json:"audio_id"`
	FileSize   int64  `json:"file_size"`
	UploadTime int64  `json:"upload_time"`
}

// AudioRecordingInfo 音频录音信息
type AudioRecordingInfo struct {
	ID         string `json:"id"`
	SessionID  string `json:"session_id"`
	AudioID    string `json:"audio_id"`
	MacAddress string `json:"mac_address"`
	FilePath   string `json:"file_path"`
	FileSize   int64  `json:"file_size"`
	UploadTime int64  `json:"upload_time"`
	Status     int8   `json:"status"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  int64  `json:"updated_at"`
	// 播放链接
	PlayURL string `json:"play_url"`
}

// AudioRecordingListRequest 音频录音列表请求
type AudioRecordingListRequest struct {
	SessionID  string `form:"session_id"`
	MacAddress string `form:"mac_address"`
	StartTime  int64  `form:"start_time"`
	EndTime    int64  `form:"end_time"`
	Status     *int8  `form:"status"`
	Offset     int    `form:"offset"`
	Limit      int    `form:"limit"`
}

// AudioRecordingListResponse 音频录音列表响应
type AudioRecordingListResponse struct {
	Total int64                 `json:"total"`
	Items []*AudioRecordingInfo `json:"items"`
}

// AudioRecordingDownloadRequest 音频录音下载请求
type AudioRecordingDownloadRequest struct {
	ID string `form:"id" binding:"required"`
}

// AudioRecordingDownloadResponse 音频录音下载响应
type AudioRecordingDownloadResponse struct {
	DownloadURL string `json:"download_url"`
	ExpiresAt   int64  `json:"expires_at"`
}
