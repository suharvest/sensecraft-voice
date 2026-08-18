package db

import (
	"context"

	"gorm.io/gorm"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
)

// AudioSessionInterface 音频会话数据库接口
type AudioSessionInterface interface {
	Create(ctx context.Context, object *model.AudioSession) (*model.AudioSession, error)
	GetBySessionID(ctx context.Context, sessionID string) (*model.AudioSession, error)
	Update(ctx context.Context, object *model.AudioSession) error
	List(ctx context.Context, in AudioSessionListRequest) ([]*model.AudioSession, error)
	Count(ctx context.Context, in AudioSessionListRequest) (int64, error)
}

// AudioChunkInterface 音频块数据库接口
type AudioChunkInterface interface {
	Create(ctx context.Context, object *model.AudioChunk) (*model.AudioChunk, error)
	GetBySessionID(ctx context.Context, sessionID string) ([]*model.AudioChunk, error)
	GetBySessionIDAndIndex(ctx context.Context, sessionID string, chunkIndex int) (*model.AudioChunk, error)
	List(ctx context.Context, in AudioChunkListRequest) ([]*model.AudioChunk, error)
	Count(ctx context.Context, in AudioChunkListRequest) (int64, error)
	Update(ctx context.Context, object *model.AudioChunk) error
}

// TimeSyncInterface 时间同步数据库接口
type TimeSyncInterface interface {
	CreateOrUpdate(ctx context.Context, object *model.TimeSync) (*model.TimeSync, error)
	GetByDeviceID(ctx context.Context, deviceID string) (*model.TimeSync, error)
	List(ctx context.Context, deviceIDs []string) ([]*model.TimeSync, error)
}

// AudioSessionListRequest 音频会话列表请求
type AudioSessionListRequest struct {
	DeviceID  string
	StartTime int64
	EndTime   int64
	Status    *int8
	Offset    int
	Limit     int
}

// AudioChunkListRequest 音频块列表请求
type AudioChunkListRequest struct {
	SessionID string
	StartTime int64
	EndTime   int64
	Offset    int
	Limit     int
}

// audioSession 音频会话数据库实现
type audioSession struct {
	db *gorm.DB
}

func newAudioSession(db *gorm.DB) AudioSessionInterface {
	return &audioSession{db: db}
}

func (a *audioSession) Create(ctx context.Context, object *model.AudioSession) (*model.AudioSession, error) {
	// 设置创建和更新时间
	now := model.GetCurrentTimestamp()
	object.CreatedAt = now
	object.UpdatedAt = now

	if err := a.db.WithContext(ctx).Create(object).Error; err != nil {
		return nil, err
	}
	return object, nil
}

func (a *audioSession) GetBySessionID(ctx context.Context, sessionID string) (*model.AudioSession, error) {
	var session model.AudioSession
	err := a.db.WithContext(ctx).Where("session_id = ?", sessionID).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (a *audioSession) Update(ctx context.Context, object *model.AudioSession) error {
	object.UpdatedAt = model.GetCurrentTimestamp()
	return a.db.WithContext(ctx).Save(object).Error
}

func (a *audioSession) List(ctx context.Context, in AudioSessionListRequest) ([]*model.AudioSession, error) {
	var sessions []*model.AudioSession
	query := a.db.WithContext(ctx).Model(&model.AudioSession{})

	// 添加查询条件
	if in.DeviceID != "" {
		query = query.Where("device_id = ?", in.DeviceID)
	}
	if in.StartTime > 0 {
		query = query.Where("start_time >= ?", in.StartTime)
	}
	if in.EndTime > 0 {
		query = query.Where("start_time <= ?", in.EndTime)
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

	err := query.Find(&sessions).Error
	return sessions, err
}

func (a *audioSession) Count(ctx context.Context, in AudioSessionListRequest) (int64, error) {
	var count int64
	query := a.db.WithContext(ctx).Model(&model.AudioSession{})

	// 添加查询条件
	if in.DeviceID != "" {
		query = query.Where("device_id = ?", in.DeviceID)
	}
	if in.StartTime > 0 {
		query = query.Where("start_time >= ?", in.StartTime)
	}
	if in.EndTime > 0 {
		query = query.Where("start_time <= ?", in.EndTime)
	}
	if in.Status != nil {
		query = query.Where("status = ?", *in.Status)
	}

	err := query.Count(&count).Error
	return count, err
}

// audioChunk 音频块数据库实现
type audioChunk struct {
	db *gorm.DB
}

func newAudioChunk(db *gorm.DB) AudioChunkInterface {
	return &audioChunk{db: db}
}

func (a *audioChunk) Create(ctx context.Context, object *model.AudioChunk) (*model.AudioChunk, error) {
	object.CreatedAt = model.GetCurrentTimestamp()
	if err := a.db.WithContext(ctx).Create(object).Error; err != nil {
		return nil, err
	}
	return object, nil
}

func (a *audioChunk) GetBySessionID(ctx context.Context, sessionID string) ([]*model.AudioChunk, error) {
	var chunks []*model.AudioChunk
	err := a.db.WithContext(ctx).Where("session_id = ?", sessionID).Order("chunk_index ASC").Find(&chunks).Error
	return chunks, err
}

func (a *audioChunk) GetBySessionIDAndIndex(ctx context.Context, sessionID string, chunkIndex int) (*model.AudioChunk, error) {
	var chunk model.AudioChunk
	err := a.db.WithContext(ctx).Where("session_id = ? AND chunk_index = ?", sessionID, chunkIndex).First(&chunk).Error
	if err != nil {
		return nil, err
	}
	return &chunk, nil
}

func (a *audioChunk) List(ctx context.Context, in AudioChunkListRequest) ([]*model.AudioChunk, error) {
	var chunks []*model.AudioChunk
	query := a.db.WithContext(ctx).Model(&model.AudioChunk{})

	// 添加查询条件
	if in.SessionID != "" {
		query = query.Where("session_id = ?", in.SessionID)
	}
	if in.StartTime > 0 {
		query = query.Where("start_time >= ?", in.StartTime)
	}
	if in.EndTime > 0 {
		query = query.Where("end_time <= ?", in.EndTime)
	}

	// 分页
	if in.Offset > 0 {
		query = query.Offset(in.Offset)
	}
	if in.Limit > 0 {
		query = query.Limit(in.Limit)
	}

	// 排序
	query = query.Order("chunk_index ASC")

	err := query.Find(&chunks).Error
	return chunks, err
}

func (a *audioChunk) Count(ctx context.Context, in AudioChunkListRequest) (int64, error) {
	var count int64
	query := a.db.WithContext(ctx).Model(&model.AudioChunk{})

	// 添加查询条件
	if in.SessionID != "" {
		query = query.Where("session_id = ?", in.SessionID)
	}
	if in.StartTime > 0 {
		query = query.Where("start_time >= ?", in.StartTime)
	}
	if in.EndTime > 0 {
		query = query.Where("end_time <= ?", in.EndTime)
	}

	err := query.Count(&count).Error
	return count, err
}

func (a *audioChunk) Update(ctx context.Context, object *model.AudioChunk) error {
	return a.db.WithContext(ctx).Save(object).Error
}

// timeSync 时间同步数据库实现
type timeSync struct {
	db *gorm.DB
}

func newTimeSync(db *gorm.DB) TimeSyncInterface {
	return &timeSync{db: db}
}

func (t *timeSync) CreateOrUpdate(ctx context.Context, object *model.TimeSync) (*model.TimeSync, error) {
	now := model.GetCurrentTimestamp()

	// 先尝试查找现有记录
	var existing model.TimeSync
	err := t.db.WithContext(ctx).Where("device_id = ?", object.DeviceID).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// 记录不存在，创建新记录
		object.CreatedAt = now
		object.UpdatedAt = now
		object.LastSync = now
		if err := t.db.WithContext(ctx).Create(object).Error; err != nil {
			return nil, err
		}
		return object, nil
	} else if err != nil {
		return nil, err
	}

	// 记录存在，更新
	object.ID = existing.ID
	object.CreatedAt = existing.CreatedAt
	object.UpdatedAt = now
	object.LastSync = now
	if err := t.db.WithContext(ctx).Save(object).Error; err != nil {
		return nil, err
	}

	return object, nil
}

func (t *timeSync) GetByDeviceID(ctx context.Context, deviceID string) (*model.TimeSync, error) {
	var sync model.TimeSync
	err := t.db.WithContext(ctx).Where("device_id = ?", deviceID).First(&sync).Error
	if err != nil {
		return nil, err
	}
	return &sync, nil
}

func (t *timeSync) List(ctx context.Context, deviceIDs []string) ([]*model.TimeSync, error) {
	var syncs []*model.TimeSync
	query := t.db.WithContext(ctx).Model(&model.TimeSync{})

	if len(deviceIDs) > 0 {
		query = query.Where("device_id IN ?", deviceIDs)
	}

	err := query.Find(&syncs).Error
	return syncs, err
}
