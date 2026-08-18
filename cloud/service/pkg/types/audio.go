package types

// AudioChunkRequest 音频块上传请求
type AudioChunkRequest struct {
	DeviceID  string    `json:"device_id" binding:"required"`
	SessionID string    `json:"session_id"` // 可选，如果为空则自动生成
	ChunkInfo ChunkInfo `json:"chunk_info" binding:"required"`
	AudioData string    `json:"audio_data" binding:"required"` // base64编码的音频数据
	Checksum  string    `json:"checksum" binding:"required"`   // MD5校验和
}

// ChunkInfo 音频块信息
type ChunkInfo struct {
	ChunkIndex int    `json:"chunk_index" binding:"required,min=1"`
	StartTime  int64  `json:"start_time" binding:"required"`
	EndTime    int64  `json:"end_time" binding:"required"`
	Duration   int    `json:"duration" binding:"required,min=1"`
	Format     string `json:"format" binding:"required"`
	SampleRate int    `json:"sample_rate" binding:"required,min=1"`
	Channels   int    `json:"channels" binding:"required,min=1"`
	BitRate    int    `json:"bit_rate,omitempty"`
}

// AudioChunkResponse 音频块上传响应
type AudioChunkResponse struct {
	ChunkID    string `json:"chunk_id"`
	SessionID  string `json:"session_id"`
	MinIOPath  string `json:"minio_path"`
	UploadedAt int64  `json:"uploaded_at"`
}

// AudioSessionRequest 创建音频会话请求
type AudioSessionRequest struct {
	DeviceID  string `json:"device_id" binding:"required"`
	SessionID string `json:"session_id,omitempty"` // 可选，如果为空则自动生成
	StartTime int64  `json:"start_time" binding:"required"`
}

// AudioSessionResponse 音频会话响应
type AudioSessionResponse struct {
	SessionID string `json:"session_id"`
	DeviceID  string `json:"device_id"`
	StartTime int64  `json:"start_time"`
	Status    int8   `json:"status"`
	CreatedAt int64  `json:"created_at"`
}

// AudioSessionListRequest 音频会话列表请求
type AudioSessionListRequest struct {
	DeviceID  string `form:"device_id"`
	StartTime int64  `form:"start_time"`
	EndTime   int64  `form:"end_time"`
	Status    *int8  `form:"status"`
	Offset    int    `form:"offset"`
	Limit     int    `form:"limit"`
}

// AudioSessionListResponse 音频会话列表响应
type AudioSessionListResponse struct {
	Total int64               `json:"total"`
	Items []*AudioSessionInfo `json:"items"`
}

// AudioSessionInfo 音频会话信息
type AudioSessionInfo struct {
	ID          int64  `json:"id"`
	SessionID   string `json:"session_id"`
	DeviceID    string `json:"device_id"`
	StartTime   int64  `json:"start_time"`
	EndTime     *int64 `json:"end_time"`
	TotalChunks int    `json:"total_chunks"`
	Status      int8   `json:"status"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

// AudioChunkListRequest 音频块列表请求
type AudioChunkListRequest struct {
	SessionID string `form:"session_id" binding:"required"`
	StartTime int64  `form:"start_time"`
	EndTime   int64  `form:"end_time"`
	Offset    int    `form:"offset"`
	Limit     int    `form:"limit"`
}

// AudioChunkListResponse 音频块列表响应
type AudioChunkListResponse struct {
	Total int64             `json:"total"`
	Items []*AudioChunkInfo `json:"items"`
}

// AudioChunkInfo 音频块信息
type AudioChunkInfo struct {
	ID         int64  `json:"id"`
	SessionID  string `json:"session_id"`
	ChunkIndex int    `json:"chunk_index"`
	DeviceID   string `json:"device_id"`
	StartTime  int64  `json:"start_time"`
	EndTime    int64  `json:"end_time"`
	Duration   int    `json:"duration"`
	FileSize   int64  `json:"file_size"`
	Format     string `json:"format"`
	SampleRate int    `json:"sample_rate"`
	Channels   int8   `json:"channels"`
	MinIOPath  string `json:"minio_path"`
	Checksum   string `json:"checksum"`
	Status     int8   `json:"status"`
	CreatedAt  int64  `json:"created_at"`
}

// TimeSyncRequest 时间同步请求
type TimeSyncRequest struct {
	DeviceID   string `json:"device_id" binding:"required"`
	DeviceTime int64  `json:"device_time" binding:"required"`
	ServerTime int64  `json:"server_time" binding:"required"`
}

// TimeSyncResponse 时间同步响应
type TimeSyncResponse struct {
	DeviceID   string `json:"device_id"`
	DeviceTime int64  `json:"device_time"`
	ServerTime int64  `json:"server_time"`
	Offset     int64  `json:"offset"`
	LastSync   int64  `json:"last_sync"`
}

// AudioDownloadRequest 音频下载请求
type AudioDownloadRequest struct {
	SessionID  string `form:"session_id" binding:"required"`
	ChunkIndex *int   `form:"chunk_index"` // 可选，指定下载某个块
	Format     string `form:"format"`      // 可选，指定格式
}

// AudioDownloadResponse 音频下载响应
type AudioDownloadResponse struct {
	DownloadURL string `json:"download_url"`
	ExpiresAt   int64  `json:"expires_at"`
}
