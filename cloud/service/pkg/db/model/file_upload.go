package model

func init() {
	register(&FileUpload{})
}

// FileUpload 文件上传记录表
type FileUpload struct {
	ID           int64  `gorm:"column:id;primaryKey;autoIncrement;not null" json:"id"`
	FileName     string `gorm:"column:file_name;type:varchar(255);not null" json:"file_name"`
	OriginalName string `gorm:"column:original_name;type:varchar(255);not null" json:"original_name"`
	FileSize     int64  `gorm:"column:file_size;type:bigint;not null" json:"file_size"`
	ContentType  string `gorm:"column:content_type;type:varchar(100);not null" json:"content_type"`
	MinIOPath    string `gorm:"column:minio_path;type:varchar(500);not null" json:"minio_path"`
	Checksum     string `gorm:"column:checksum;type:varchar(32);not null" json:"checksum"`
	Uploader     string `gorm:"column:uploader;type:varchar(64);default:''" json:"uploader"`
	Status       int8   `gorm:"column:status;type:tinyint;default:1;not null" json:"status"` // 1=正常,0=已删除
	CreatedAt    int64  `gorm:"column:created_at;type:bigint;not null" json:"created_at"`
	UpdatedAt    int64  `gorm:"column:updated_at;type:bigint;not null" json:"updated_at"`
}

func (f *FileUpload) TableName() string { return "file_uploads" }

// 文件状态常量
const (
	FileStatusNormal  = 1 // 正常
	FileStatusDeleted = 0 // 已删除
)
