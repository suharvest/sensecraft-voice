package db

import (
	"context"

	"gorm.io/gorm"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
)

// FileUploadInterface 文件上传数据库接口
type FileUploadInterface interface {
	Create(ctx context.Context, object *model.FileUpload) (*model.FileUpload, error)
	GetByID(ctx context.Context, id int64) (*model.FileUpload, error)
	GetByMinIOPath(ctx context.Context, minioPath string) (*model.FileUpload, error)
	List(ctx context.Context, in FileUploadListRequest) ([]*model.FileUpload, error)
	Count(ctx context.Context, in FileUploadListRequest) (int64, error)
	Update(ctx context.Context, object *model.FileUpload) error
	Delete(ctx context.Context, id int64) error
}

// FileUploadListRequest 文件上传列表请求
type FileUploadListRequest struct {
	Uploader string
	Status   *int8
	Offset   int
	Limit    int
}

// fileUpload 文件上传数据库实现
type fileUpload struct {
	db *gorm.DB
}

func newFileUpload(db *gorm.DB) FileUploadInterface {
	return &fileUpload{db: db}
}

func (f *fileUpload) Create(ctx context.Context, object *model.FileUpload) (*model.FileUpload, error) {
	// 设置创建和更新时间
	now := model.GetCurrentTimestamp()
	object.CreatedAt = now
	object.UpdatedAt = now

	if err := f.db.WithContext(ctx).Create(object).Error; err != nil {
		return nil, err
	}
	return object, nil
}

func (f *fileUpload) GetByID(ctx context.Context, id int64) (*model.FileUpload, error) {
	var file model.FileUpload
	err := f.db.WithContext(ctx).Where("id = ? AND status = ?", id, model.FileStatusNormal).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (f *fileUpload) GetByMinIOPath(ctx context.Context, minioPath string) (*model.FileUpload, error) {
	var file model.FileUpload
	err := f.db.WithContext(ctx).Where("minio_path = ? AND status = ?", minioPath, model.FileStatusNormal).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (f *fileUpload) List(ctx context.Context, in FileUploadListRequest) ([]*model.FileUpload, error) {
	var files []*model.FileUpload
	query := f.db.WithContext(ctx).Model(&model.FileUpload{}).Where("status = ?", model.FileStatusNormal)

	// 添加查询条件
	if in.Uploader != "" {
		query = query.Where("uploader = ?", in.Uploader)
	}
	if in.Status != nil {
		query = query.Where("status = ?", *in.Status)
	}

	// 分页
	if in.Offset > 0 {
		query = query.Offset(in.Offset)
	}
	if in.Limit > 0 {
		query = query.Limit(in.Limit)
	}

	// 排序
	query = query.Order("created_at DESC")

	err := query.Find(&files).Error
	return files, err
}

func (f *fileUpload) Count(ctx context.Context, in FileUploadListRequest) (int64, error) {
	var count int64
	query := f.db.WithContext(ctx).Model(&model.FileUpload{}).Where("status = ?", model.FileStatusNormal)

	// 添加查询条件
	if in.Uploader != "" {
		query = query.Where("uploader = ?", in.Uploader)
	}
	if in.Status != nil {
		query = query.Where("status = ?", *in.Status)
	}

	err := query.Count(&count).Error
	return count, err
}

func (f *fileUpload) Update(ctx context.Context, object *model.FileUpload) error {
	object.UpdatedAt = model.GetCurrentTimestamp()
	return f.db.WithContext(ctx).Save(object).Error
}

func (f *fileUpload) Delete(ctx context.Context, id int64) error {
	// 软删除，将状态设为已删除
	return f.db.WithContext(ctx).Model(&model.FileUpload{}).Where("id = ?", id).Update("status", model.FileStatusDeleted).Error
}
