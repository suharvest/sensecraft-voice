package db

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
)

// AudioRecordingInterface 音频录音数据库接口
type AudioRecordingInterface interface {
	Create(ctx context.Context, object *model.AudioRecording) (*model.AudioRecording, error)
	GetByID(ctx context.Context, id string) (*model.AudioRecording, error)
	GetBySessionIDAndAudioID(ctx context.Context, sessionID, audioID string) (*model.AudioRecording, error)
	GetBySessionID(ctx context.Context, sessionID string) ([]*model.AudioRecording, error)
	GetByMacAddress(ctx context.Context, macAddress string) ([]*model.AudioRecording, error)
	List(ctx context.Context, in AudioRecordingListRequest) ([]*model.AudioRecording, error)
	Count(ctx context.Context, in AudioRecordingListRequest) (int64, error)
	Update(ctx context.Context, object *model.AudioRecording) error
	Delete(ctx context.Context, id string) error
	// 批量查询方法
	GetBySessionIDAndAudioIDBatch(ctx context.Context, sessionAudioPairs []SessionAudioPair) ([]*model.AudioRecording, error)
}

// SessionAudioPair 会话和音频ID对
type SessionAudioPair struct {
	SessionID string
	AudioID   string
}

// AudioRecordingListRequest 音频录音列表请求
type AudioRecordingListRequest struct {
	SessionID  string
	MacAddress string
	StartTime  int64
	EndTime    int64
	Status     *int8
	Offset     int
	Limit      int
}

// audioRecording 音频录音数据库实现
type audioRecording struct {
	db *gorm.DB
}

func newAudioRecording(db *gorm.DB) AudioRecordingInterface {
	return &audioRecording{db: db}
}

func (a *audioRecording) Create(ctx context.Context, object *model.AudioRecording) (*model.AudioRecording, error) {
	// 设置创建和更新时间
	now := model.GetCurrentTimestamp()
	object.CreatedAt = now
	object.UpdatedAt = now

	if err := a.db.WithContext(ctx).Create(object).Error; err != nil {
		return nil, err
	}
	return object, nil
}

func (a *audioRecording) GetByID(ctx context.Context, id string) (*model.AudioRecording, error) {
	var recording model.AudioRecording
	err := a.db.WithContext(ctx).Where("id = ?", id).First(&recording).Error
	if err != nil {
		return nil, err
	}
	return &recording, nil
}

func (a *audioRecording) GetBySessionIDAndAudioID(ctx context.Context, sessionID, audioID string) (*model.AudioRecording, error) {
	var recording model.AudioRecording
	err := a.db.WithContext(ctx).Where("session_id = ? AND audio_id = ?", sessionID, audioID).First(&recording).Error
	if err != nil {
		return nil, err
	}
	return &recording, nil
}

func (a *audioRecording) GetBySessionID(ctx context.Context, sessionID string) ([]*model.AudioRecording, error) {
	var recordings []*model.AudioRecording
	err := a.db.WithContext(ctx).Where("session_id = ?", sessionID).Order("upload_time ASC").Find(&recordings).Error
	return recordings, err
}

func (a *audioRecording) GetByMacAddress(ctx context.Context, macAddress string) ([]*model.AudioRecording, error) {
	var recordings []*model.AudioRecording
	err := a.db.WithContext(ctx).Where("mac_address = ?", macAddress).Order("upload_time DESC").Find(&recordings).Error
	return recordings, err
}

func (a *audioRecording) List(ctx context.Context, in AudioRecordingListRequest) ([]*model.AudioRecording, error) {
	var recordings []*model.AudioRecording
	query := a.db.WithContext(ctx).Model(&model.AudioRecording{})

	// 添加查询条件
	if in.SessionID != "" {
		query = query.Where("session_id = ?", in.SessionID)
	}
	if in.MacAddress != "" {
		query = query.Where("mac_address = ?", in.MacAddress)
	}
	if in.StartTime > 0 {
		query = query.Where("upload_time >= ?", in.StartTime)
	}
	if in.EndTime > 0 {
		query = query.Where("upload_time <= ?", in.EndTime)
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
	query = query.Order("upload_time DESC")

	err := query.Find(&recordings).Error
	return recordings, err
}

func (a *audioRecording) Count(ctx context.Context, in AudioRecordingListRequest) (int64, error) {
	var count int64
	query := a.db.WithContext(ctx).Model(&model.AudioRecording{})

	// 添加查询条件
	if in.SessionID != "" {
		query = query.Where("session_id = ?", in.SessionID)
	}
	if in.MacAddress != "" {
		query = query.Where("mac_address = ?", in.MacAddress)
	}
	if in.StartTime > 0 {
		query = query.Where("upload_time >= ?", in.StartTime)
	}
	if in.EndTime > 0 {
		query = query.Where("upload_time <= ?", in.EndTime)
	}
	if in.Status != nil {
		query = query.Where("status = ?", *in.Status)
	}

	err := query.Count(&count).Error
	return count, err
}

func (a *audioRecording) Update(ctx context.Context, object *model.AudioRecording) error {
	object.UpdatedAt = model.GetCurrentTimestamp()
	return a.db.WithContext(ctx).Save(object).Error
}

func (a *audioRecording) Delete(ctx context.Context, id string) error {
	// 软删除，更新状态为已删除
	return a.db.WithContext(ctx).Model(&model.AudioRecording{}).Where("id = ?", id).Update("status", model.AudioRecordingStatusDeleted).Error
}

func (a *audioRecording) GetBySessionIDAndAudioIDBatch(ctx context.Context, sessionAudioPairs []SessionAudioPair) ([]*model.AudioRecording, error) {
	if len(sessionAudioPairs) == 0 {
		return []*model.AudioRecording{}, nil
	}

	var results []*model.AudioRecording

	// 构建批量查询条件
	var conditions []string
	var args []interface{}

	for _, pair := range sessionAudioPairs {
		conditions = append(conditions, "(session_id = ? AND audio_id = ?)")
		args = append(args, pair.SessionID, pair.AudioID)
	}

	// 执行批量查询
	query := a.db.WithContext(ctx).Model(&model.AudioRecording{}).
		Where("status != ?", model.AudioRecordingStatusDeleted).
		Where("("+strings.Join(conditions, " OR ")+")", args...)

	err := query.Find(&results).Error
	return results, err
}
