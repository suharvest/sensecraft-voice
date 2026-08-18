package db

import (
	"context"
	"strings"
	"time"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
	"gorm.io/gorm"
)

type KeywordMatchInterface interface {
	Create(ctx context.Context, match *model.KeywordMatch) error
	GetByRecordingID(ctx context.Context, recordingID int64) ([]*model.KeywordMatch, error)
	GetByMacAndKeyword(ctx context.Context, macAddress string, keywordID int64) ([]*model.KeywordMatch, error)
	GetByMacAddress(ctx context.Context, macAddress string, limit int) ([]*model.KeywordMatch, error)
	GetByKeywordID(ctx context.Context, keywordID int64, limit int) ([]*model.KeywordMatch, error)
	GetAll(ctx context.Context, limit int) ([]*model.KeywordMatch, error)
	GetByTimeRange(ctx context.Context, startTime, endTime int64, limit int) ([]*model.KeywordMatch, error)
	GetByConditions(ctx context.Context, macAddress string, keywordID int64, startTime, endTime int64, limit int) ([]*model.KeywordMatch, error)
	GetByStoreID(ctx context.Context, storeID int64, startTime, endTime int64, limit int) ([]*model.KeywordMatch, error)
	DeleteByRecordingID(ctx context.Context, recordingID int64) error
	BatchCreate(ctx context.Context, matches []*model.KeywordMatch) error
	// Count 方法用于统计满足条件的总数
	CountByRecordingID(ctx context.Context, recordingID int64) (int64, error)
	CountByStoreID(ctx context.Context, storeID int64, startTime, endTime int64) (int64, error)
	CountByConditions(ctx context.Context, macAddress string, keywordID int64, startTime, endTime int64) (int64, error)
	CountAll(ctx context.Context) (int64, error)
}

type keywordMatch struct {
	db *gorm.DB
}

func newKeywordMatch(db *gorm.DB) KeywordMatchInterface {
	return &keywordMatch{db: db}
}

func (km *keywordMatch) Create(ctx context.Context, match *model.KeywordMatch) error {
	match.CreatedAt = time.Now().UnixMilli()
	return km.db.WithContext(ctx).Create(match).Error
}

func (km *keywordMatch) GetByRecordingID(ctx context.Context, recordingID int64) ([]*model.KeywordMatch, error) {
	var matches []*model.KeywordMatchQuery
	err := km.db.WithContext(ctx).
		Table("keyword_matches").
		Select(`
			keyword_matches.id,
			keyword_matches.recording_id,
			keyword_matches.mac_address,
			keyword_matches.keyword_id,
			keyword_matches.keyword,
			keyword_matches.matched_text,
			keyword_matches.match_type,
			keyword_matches.confidence,
			keyword_matches.position,
			keyword_matches.length,
			keyword_matches.created_at,
			recordings.session_id,
			recordings.audio_id,
			recordings.speaker_id,
			recordings.speaker_name,
			recordings.text,
			recordings.device_time,
			recordings.status
		`).
		Joins("JOIN recordings ON keyword_matches.recording_id = recordings.id").
		Where("keyword_matches.recording_id = ?", recordingID).
		Order("keyword_matches.created_at desc").
		Find(&matches).Error

	if err != nil {
		return nil, err
	}

	// 转换为KeywordMatch格式
	result := make([]*model.KeywordMatch, len(matches))
	for i, match := range matches {
		result[i] = &model.KeywordMatch{
			ID:          match.ID,
			RecordingID: match.RecordingID,
			MacAddress:  match.MacAddress,
			KeywordID:   match.KeywordID,
			Keyword:     match.Keyword,
			MatchedText: match.MatchedText,
			MatchType:   match.MatchType,
			Confidence:  match.Confidence,
			Position:    match.Position,
			Length:      match.Length,
			CreatedAt:   match.CreatedAt,
			SessionID:   match.SessionID,
			AudioID:     match.AudioID,
			SpeakerID:   match.SpeakerID,
			SpeakerName: match.SpeakerName,
			Text:        match.Text,
			DeviceTime:  match.DeviceTime,
			Status:      match.Status,
		}
	}

	return result, nil
}

func (km *keywordMatch) GetByMacAndKeyword(ctx context.Context, macAddress string, keywordID int64) ([]*model.KeywordMatch, error) {
	var matches []*model.KeywordMatchQuery
	err := km.db.WithContext(ctx).
		Table("keyword_matches").
		Select(`
			keyword_matches.id,
			keyword_matches.recording_id,
			keyword_matches.mac_address,
			keyword_matches.keyword_id,
			keyword_matches.keyword,
			keyword_matches.matched_text,
			keyword_matches.match_type,
			keyword_matches.confidence,
			keyword_matches.position,
			keyword_matches.length,
			keyword_matches.created_at,
			recordings.session_id,
			recordings.audio_id,
			recordings.speaker_id,
			recordings.speaker_name,
			recordings.text,
			recordings.device_time,
			recordings.status
		`).
		Joins("JOIN recordings ON keyword_matches.recording_id = recordings.id").
		Where("keyword_matches.mac_address = ? AND keyword_matches.keyword_id = ?", macAddress, keywordID).
		Order("keyword_matches.created_at desc").
		Find(&matches).Error

	if err != nil {
		return nil, err
	}

	// 转换为KeywordMatch格式
	result := make([]*model.KeywordMatch, len(matches))
	for i, match := range matches {
		result[i] = &model.KeywordMatch{
			ID:          match.ID,
			RecordingID: match.RecordingID,
			MacAddress:  match.MacAddress,
			KeywordID:   match.KeywordID,
			Keyword:     match.Keyword,
			MatchedText: match.MatchedText,
			MatchType:   match.MatchType,
			Confidence:  match.Confidence,
			Position:    match.Position,
			Length:      match.Length,
			CreatedAt:   match.CreatedAt,
			SessionID:   match.SessionID,
			AudioID:     match.AudioID,
			SpeakerID:   match.SpeakerID,
			SpeakerName: match.SpeakerName,
			Text:        match.Text,
			DeviceTime:  match.DeviceTime,
			Status:      match.Status,
		}
	}

	return result, nil
}

func (km *keywordMatch) GetByMacAddress(ctx context.Context, macAddress string, limit int) ([]*model.KeywordMatch, error) {
	var matches []*model.KeywordMatchQuery
	query := km.db.WithContext(ctx).
		Table("keyword_matches").
		Select(`
			keyword_matches.id,
			keyword_matches.recording_id,
			keyword_matches.mac_address,
			keyword_matches.keyword_id,
			keyword_matches.keyword,
			keyword_matches.matched_text,
			keyword_matches.match_type,
			keyword_matches.confidence,
			keyword_matches.position,
			keyword_matches.length,
			keyword_matches.created_at,
			recordings.session_id,
			recordings.audio_id,
			recordings.speaker_id,
			recordings.speaker_name,
			recordings.text,
			recordings.device_time,
			recordings.status
		`).
		Joins("JOIN recordings ON keyword_matches.recording_id = recordings.id").
		Where("keyword_matches.mac_address = ?", macAddress)
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Order("keyword_matches.created_at desc").Find(&matches).Error

	if err != nil {
		return nil, err
	}

	// 转换为KeywordMatch格式
	result := make([]*model.KeywordMatch, len(matches))
	for i, match := range matches {
		result[i] = &model.KeywordMatch{
			ID:          match.ID,
			RecordingID: match.RecordingID,
			MacAddress:  match.MacAddress,
			KeywordID:   match.KeywordID,
			Keyword:     match.Keyword,
			MatchedText: match.MatchedText,
			MatchType:   match.MatchType,
			Confidence:  match.Confidence,
			Position:    match.Position,
			Length:      match.Length,
			CreatedAt:   match.CreatedAt,
			SessionID:   match.SessionID,
			AudioID:     match.AudioID,
			SpeakerID:   match.SpeakerID,
			SpeakerName: match.SpeakerName,
			Text:        match.Text,
			DeviceTime:  match.DeviceTime,
			Status:      match.Status,
		}
	}

	return result, nil
}

func (km *keywordMatch) GetByKeywordID(ctx context.Context, keywordID int64, limit int) ([]*model.KeywordMatch, error) {
	var matches []*model.KeywordMatchQuery
	query := km.db.WithContext(ctx).
		Table("keyword_matches").
		Select(`
			keyword_matches.id,
			keyword_matches.recording_id,
			keyword_matches.mac_address,
			keyword_matches.keyword_id,
			keyword_matches.keyword,
			keyword_matches.matched_text,
			keyword_matches.match_type,
			keyword_matches.confidence,
			keyword_matches.position,
			keyword_matches.length,
			keyword_matches.created_at,
			recordings.session_id,
			recordings.audio_id,
			recordings.speaker_id,
			recordings.speaker_name,
			recordings.text,
			recordings.device_time,
			recordings.status
		`).
		Joins("JOIN recordings ON keyword_matches.recording_id = recordings.id").
		Where("keyword_matches.keyword_id = ?", keywordID)
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Order("keyword_matches.created_at desc").Find(&matches).Error

	if err != nil {
		return nil, err
	}

	// 转换为KeywordMatch格式
	result := make([]*model.KeywordMatch, len(matches))
	for i, match := range matches {
		result[i] = &model.KeywordMatch{
			ID:          match.ID,
			RecordingID: match.RecordingID,
			MacAddress:  match.MacAddress,
			KeywordID:   match.KeywordID,
			Keyword:     match.Keyword,
			MatchedText: match.MatchedText,
			MatchType:   match.MatchType,
			Confidence:  match.Confidence,
			Position:    match.Position,
			Length:      match.Length,
			CreatedAt:   match.CreatedAt,
			SessionID:   match.SessionID,
			AudioID:     match.AudioID,
			SpeakerID:   match.SpeakerID,
			SpeakerName: match.SpeakerName,
			Text:        match.Text,
			DeviceTime:  match.DeviceTime,
			Status:      match.Status,
		}
	}

	return result, nil
}

func (km *keywordMatch) GetAll(ctx context.Context, limit int) ([]*model.KeywordMatch, error) {
	var matches []*model.KeywordMatchQuery
	query := km.db.WithContext(ctx).
		Table("keyword_matches").
		Select(`
			keyword_matches.id,
			keyword_matches.recording_id,
			keyword_matches.mac_address,
			keyword_matches.keyword_id,
			keyword_matches.keyword,
			keyword_matches.matched_text,
			keyword_matches.match_type,
			keyword_matches.confidence,
			keyword_matches.position,
			keyword_matches.length,
			keyword_matches.created_at,
			recordings.session_id,
			recordings.audio_id,
			recordings.speaker_id,
			recordings.speaker_name,
			recordings.text,
			recordings.device_time,
			recordings.status
		`).
		Joins("JOIN recordings ON keyword_matches.recording_id = recordings.id")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Order("keyword_matches.created_at desc").Find(&matches).Error

	if err != nil {
		return nil, err
	}

	// 转换为KeywordMatch格式
	result := make([]*model.KeywordMatch, len(matches))
	for i, match := range matches {
		result[i] = &model.KeywordMatch{
			ID:          match.ID,
			RecordingID: match.RecordingID,
			MacAddress:  match.MacAddress,
			KeywordID:   match.KeywordID,
			Keyword:     match.Keyword,
			MatchedText: match.MatchedText,
			MatchType:   match.MatchType,
			Confidence:  match.Confidence,
			Position:    match.Position,
			Length:      match.Length,
			CreatedAt:   match.CreatedAt,
			SessionID:   match.SessionID,
			AudioID:     match.AudioID,
			SpeakerID:   match.SpeakerID,
			SpeakerName: match.SpeakerName,
			Text:        match.Text,
			DeviceTime:  match.DeviceTime,
			Status:      match.Status,
		}
	}

	return result, nil
}

func (km *keywordMatch) GetByTimeRange(ctx context.Context, startTime, endTime int64, limit int) ([]*model.KeywordMatch, error) {
	var matches []*model.KeywordMatchQuery
	query := km.db.WithContext(ctx).
		Table("keyword_matches").
		Select(`
			keyword_matches.id,
			keyword_matches.recording_id,
			keyword_matches.mac_address,
			keyword_matches.keyword_id,
			keyword_matches.keyword,
			keyword_matches.matched_text,
			keyword_matches.match_type,
			keyword_matches.confidence,
			keyword_matches.position,
			keyword_matches.length,
			keyword_matches.created_at,
			recordings.session_id,
			recordings.audio_id,
			recordings.speaker_id,
			recordings.speaker_name,
			recordings.text,
			recordings.device_time,
			recordings.status
		`).
		Joins("JOIN recordings ON keyword_matches.recording_id = recordings.id").
		Where("recordings.device_time >= ? AND recordings.device_time <= ?", startTime, endTime)
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Order("keyword_matches.created_at desc").Find(&matches).Error

	if err != nil {
		return nil, err
	}

	// 转换为KeywordMatch格式
	result := make([]*model.KeywordMatch, len(matches))
	for i, match := range matches {
		result[i] = &model.KeywordMatch{
			ID:          match.ID,
			RecordingID: match.RecordingID,
			MacAddress:  match.MacAddress,
			KeywordID:   match.KeywordID,
			Keyword:     match.Keyword,
			MatchedText: match.MatchedText,
			MatchType:   match.MatchType,
			Confidence:  match.Confidence,
			Position:    match.Position,
			Length:      match.Length,
			CreatedAt:   match.CreatedAt,
			SessionID:   match.SessionID,
			AudioID:     match.AudioID,
			SpeakerID:   match.SpeakerID,
			SpeakerName: match.SpeakerName,
			Text:        match.Text,
			DeviceTime:  match.DeviceTime,
			Status:      match.Status,
		}
	}

	return result, nil
}

func (km *keywordMatch) GetByConditions(ctx context.Context, macAddress string, keywordID int64, startTime, endTime int64, limit int) ([]*model.KeywordMatch, error) {
	var matches []*model.KeywordMatchQuery
	query := km.db.WithContext(ctx).
		Table("keyword_matches").
		Select(`
			keyword_matches.id,
			keyword_matches.recording_id,
			keyword_matches.mac_address,
			keyword_matches.keyword_id,
			keyword_matches.keyword,
			keyword_matches.matched_text,
			keyword_matches.match_type,
			keyword_matches.confidence,
			keyword_matches.position,
			keyword_matches.length,
			keyword_matches.created_at,
			recordings.session_id,
			recordings.audio_id,
			recordings.speaker_id,
			recordings.speaker_name,
			recordings.text,
			recordings.device_time,
			recordings.status
		`).
		Joins("JOIN recordings ON keyword_matches.recording_id = recordings.id")

	// 构建动态查询条件
	conditions := []string{}
	args := []interface{}{}

	if macAddress != "" {
		conditions = append(conditions, "keyword_matches.mac_address = ?")
		args = append(args, macAddress)
	}

	if keywordID > 0 {
		conditions = append(conditions, "keyword_matches.keyword_id = ?")
		args = append(args, keywordID)
	}

	if startTime > 0 && endTime > 0 {
		conditions = append(conditions, "recordings.device_time >= ? AND recordings.device_time <= ?")
		args = append(args, startTime, endTime)
	}

	if len(conditions) > 0 {
		whereClause := strings.Join(conditions, " AND ")
		query = query.Where(whereClause, args...)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Order("keyword_matches.created_at desc").Find(&matches).Error

	if err != nil {
		return nil, err
	}

	// 转换为KeywordMatch格式
	result := make([]*model.KeywordMatch, len(matches))
	for i, match := range matches {
		result[i] = &model.KeywordMatch{
			ID:          match.ID,
			RecordingID: match.RecordingID,
			MacAddress:  match.MacAddress,
			KeywordID:   match.KeywordID,
			Keyword:     match.Keyword,
			MatchedText: match.MatchedText,
			MatchType:   match.MatchType,
			Confidence:  match.Confidence,
			Position:    match.Position,
			Length:      match.Length,
			CreatedAt:   match.CreatedAt,
			SessionID:   match.SessionID,
			AudioID:     match.AudioID,
			SpeakerID:   match.SpeakerID,
			SpeakerName: match.SpeakerName,
			Text:        match.Text,
			DeviceTime:  match.DeviceTime,
			Status:      match.Status,
		}
	}

	return result, nil
}

func (km *keywordMatch) GetByStoreID(ctx context.Context, storeID int64, startTime, endTime int64, limit int) ([]*model.KeywordMatch, error) {
	var matches []*model.KeywordMatchQuery
	query := km.db.WithContext(ctx).
		Table("keyword_matches").
		Select(`
			keyword_matches.id,
			keyword_matches.recording_id,
			keyword_matches.mac_address,
			keyword_matches.keyword_id,
			keyword_matches.keyword,
			keyword_matches.matched_text,
			keyword_matches.match_type,
			keyword_matches.confidence,
			keyword_matches.position,
			keyword_matches.length,
			keyword_matches.created_at,
			recordings.session_id,
			recordings.audio_id,
			recordings.speaker_id,
			recordings.speaker_name,
			recordings.text,
			recordings.device_time,
			recordings.status
		`).
		Joins("JOIN recordings ON keyword_matches.recording_id = recordings.id").
		Joins("JOIN devices ON recordings.mac_address = devices.mac_address").
		Where("devices.store_id = ?", storeID)

	// 添加时间范围过滤
	if startTime > 0 && endTime > 0 {
		query = query.Where("recordings.device_time >= ? AND recordings.device_time <= ?", startTime, endTime)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Order("keyword_matches.created_at desc").Find(&matches).Error

	if err != nil {
		return nil, err
	}

	// 转换为KeywordMatch格式
	result := make([]*model.KeywordMatch, len(matches))
	for i, match := range matches {
		result[i] = &model.KeywordMatch{
			ID:          match.ID,
			RecordingID: match.RecordingID,
			MacAddress:  match.MacAddress,
			KeywordID:   match.KeywordID,
			Keyword:     match.Keyword,
			MatchedText: match.MatchedText,
			MatchType:   match.MatchType,
			Confidence:  match.Confidence,
			Position:    match.Position,
			Length:      match.Length,
			CreatedAt:   match.CreatedAt,
			SessionID:   match.SessionID,
			AudioID:     match.AudioID,
			SpeakerID:   match.SpeakerID,
			SpeakerName: match.SpeakerName,
			Text:        match.Text,
			DeviceTime:  match.DeviceTime,
			Status:      match.Status,
		}
	}

	return result, nil
}

func (km *keywordMatch) DeleteByRecordingID(ctx context.Context, recordingID int64) error {
	return km.db.WithContext(ctx).Where("recording_id = ?", recordingID).Delete(&model.KeywordMatch{}).Error
}

func (km *keywordMatch) BatchCreate(ctx context.Context, matches []*model.KeywordMatch) error {
	if len(matches) == 0 {
		return nil
	}

	now := time.Now().UnixMilli()
	for _, match := range matches {
		match.CreatedAt = now
	}

	return km.db.WithContext(ctx).CreateInBatches(matches, 100).Error
}

// CountByRecordingID 统计指定录音ID的匹配记录总数
func (km *keywordMatch) CountByRecordingID(ctx context.Context, recordingID int64) (int64, error) {
	var count int64
	err := km.db.WithContext(ctx).
		Table("keyword_matches").
		Joins("JOIN recordings ON keyword_matches.recording_id = recordings.id").
		Where("keyword_matches.recording_id = ?", recordingID).
		Count(&count).Error
	return count, err
}

// CountByStoreID 统计指定门店ID的匹配记录总数
func (km *keywordMatch) CountByStoreID(ctx context.Context, storeID int64, startTime, endTime int64) (int64, error) {
	var count int64
	query := km.db.WithContext(ctx).
		Table("keyword_matches").
		Joins("JOIN recordings ON keyword_matches.recording_id = recordings.id").
		Joins("JOIN devices ON recordings.mac_address = devices.mac_address").
		Where("devices.store_id = ?", storeID)

	// 添加时间范围过滤
	if startTime > 0 && endTime > 0 {
		query = query.Where("recordings.device_time >= ? AND recordings.device_time <= ?", startTime, endTime)
	}

	err := query.Count(&count).Error
	return count, err
}

// CountByConditions 统计满足组合条件的匹配记录总数
func (km *keywordMatch) CountByConditions(ctx context.Context, macAddress string, keywordID int64, startTime, endTime int64) (int64, error) {
	var count int64
	query := km.db.WithContext(ctx).
		Table("keyword_matches").
		Joins("JOIN recordings ON keyword_matches.recording_id = recordings.id")

	// 构建动态查询条件
	conditions := []string{}
	args := []interface{}{}

	if macAddress != "" {
		conditions = append(conditions, "keyword_matches.mac_address = ?")
		args = append(args, macAddress)
	}

	if keywordID > 0 {
		conditions = append(conditions, "keyword_matches.keyword_id = ?")
		args = append(args, keywordID)
	}

	if startTime > 0 && endTime > 0 {
		conditions = append(conditions, "recordings.device_time >= ? AND recordings.device_time <= ?")
		args = append(args, startTime, endTime)
	}

	if len(conditions) > 0 {
		whereClause := strings.Join(conditions, " AND ")
		query = query.Where(whereClause, args...)
	}

	err := query.Count(&count).Error
	return count, err
}

// CountAll 统计所有匹配记录的总数
func (km *keywordMatch) CountAll(ctx context.Context) (int64, error) {
	var count int64
	err := km.db.WithContext(ctx).
		Table("keyword_matches").
		Joins("JOIN recordings ON keyword_matches.recording_id = recordings.id").
		Count(&count).Error
	return count, err
}
