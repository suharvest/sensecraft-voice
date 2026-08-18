package db

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
)

// 上海时区常量
var shanghaiLocation *time.Location

func init() {
	var err error
	shanghaiLocation, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		// 如果无法加载Asia/Shanghai时区，使用UTC+8
		shanghaiLocation = time.FixedZone("CST", 8*60*60)
	}
}

type StatsInterface interface {
	GetTotalRecords(ctx context.Context, storeID *int64) (int64, error)
	GetTotalDevices(ctx context.Context, storeID *int64) (int64, error)
	GetTotalStores(ctx context.Context) (int64, error)
	GetTotalUsers(ctx context.Context, storeID *int64) (int64, error)
	GetTotalLocations(ctx context.Context, storeID *int64) (int64, error)
	GetWeeklyRecordTrend(ctx context.Context, storeID *int64) ([]WeeklyRecordCount, error)
	GetTodayActiveDevices(ctx context.Context, storeID *int64) (int64, error)
	GetTodayHourlyDistribution(ctx context.Context, storeID *int64) ([]HourlyRecordCount, error)
	GetTodayKeywordTriggers(ctx context.Context, storeID *int64) (int64, error)
	GetTodayKeywordMatches(ctx context.Context, storeID *int64) ([]KeywordMatchStats, error)
}

type WeeklyRecordCount struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type HourlyRecordCount struct {
	Hour  int   `json:"hour"` // 0-23
	Count int64 `json:"count"`
}

type KeywordMatchStats struct {
	KeywordID   int64  `json:"keyword_id"`
	Keyword     string `json:"keyword"`
	MarkColor   string `json:"mark_color"`
	MatchCount  int64  `json:"match_count"`
	RecordCount int64  `json:"record_count"` // 涉及到的记录数
}

type stats struct {
	db *gorm.DB
}

func newStats(db *gorm.DB) StatsInterface { return &stats{db: db} }

// 获取录音记录总数
func (s *stats) GetTotalRecords(ctx context.Context, storeID *int64) (int64, error) {
	var count int64
	query := s.db.WithContext(ctx).Model(&model.Recording{})

	if storeID != nil {
		query = query.Joins("JOIN devices ON recordings.mac_address = devices.mac_address").
			Where("devices.store_id = ?", *storeID)
	}

	err := query.Count(&count).Error
	return count, err
}

// 获取设备总数
func (s *stats) GetTotalDevices(ctx context.Context, storeID *int64) (int64, error) {
	var count int64
	query := s.db.WithContext(ctx).Model(&model.Device{})

	if storeID != nil {
		query = query.Where("store_id = ?", *storeID)
	}

	err := query.Count(&count).Error
	return count, err
}

// 获取门店总数
func (s *stats) GetTotalStores(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&model.Store{}).Count(&count).Error
	return count, err
}

// 获取用户总数（通过录音记录中的 speaker_id 统计）
func (s *stats) GetTotalUsers(ctx context.Context, storeID *int64) (int64, error) {
	var count int64
	query := s.db.WithContext(ctx).Model(&model.User{})

	if storeID != nil {
		// 通过录音记录关联设备，再关联门店
		query = query.Joins("JOIN recordings ON users.id = recordings.speaker_id").
			Joins("JOIN devices ON recordings.mac_address = devices.mac_address").
			Where("devices.store_id = ?", *storeID).
			Distinct("users.id")
	}

	err := query.Count(&count).Error
	return count, err
}

// 获取点位总数
func (s *stats) GetTotalLocations(ctx context.Context, storeID *int64) (int64, error) {
	var count int64
	query := s.db.WithContext(ctx).Model(&model.Location{})

	if storeID != nil {
		query = query.Where("store_id = ?", *storeID)
	}

	err := query.Count(&count).Error
	return count, err
}

// 获取本周录音记录增长趋势
func (s *stats) GetWeeklyRecordTrend(ctx context.Context, storeID *int64) ([]WeeklyRecordCount, error) {
	var results []WeeklyRecordCount

	// 获取本周开始时间 - 使用上海时区
	now := time.Now().In(shanghaiLocation)
	weekStart := now.AddDate(0, 0, -int(now.Weekday())+1) // 本周一
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, shanghaiLocation)

	// 查询本周每天的录音记录数量 - 使用上海时区
	query := s.db.WithContext(ctx).Model(&model.Recording{}).
		Select("DATE(CONVERT_TZ(FROM_UNIXTIME(recordings.created_at/1000), '+00:00', '+08:00')) as date, COUNT(*) as count").
		Where("recordings.created_at >= ?", weekStart.UnixMilli())

	if storeID != nil {
		query = query.Joins("JOIN devices ON recordings.mac_address = devices.mac_address").
			Where("devices.store_id = ?", *storeID)
	}

	rows, err := query.Group("DATE(CONVERT_TZ(FROM_UNIXTIME(recordings.created_at/1000), '+00:00', '+08:00'))").
		Order("date ASC").
		Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var result WeeklyRecordCount
		if err := rows.Scan(&result.Date, &result.Count); err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

// 获取今日产生录音记录的设备数量
func (s *stats) GetTodayActiveDevices(ctx context.Context, storeID *int64) (int64, error) {
	var count int64

	// 获取今日开始时间 - 使用上海时区
	now := time.Now().In(shanghaiLocation)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, shanghaiLocation)

	query := s.db.WithContext(ctx).Model(&model.Recording{}).
		Distinct("recordings.mac_address").
		Where("recordings.created_at >= ?", todayStart.UnixMilli())

	if storeID != nil {
		query = query.Joins("JOIN devices ON recordings.mac_address = devices.mac_address").
			Where("devices.store_id = ?", *storeID)
	}

	err := query.Count(&count).Error
	return count, err
}

// 获取今日每小时分布统计
func (s *stats) GetTodayHourlyDistribution(ctx context.Context, storeID *int64) ([]HourlyRecordCount, error) {
	var results []HourlyRecordCount

	// 获取今日开始时间 - 使用上海时区
	now := time.Now().In(shanghaiLocation)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, shanghaiLocation)
	todayEnd := todayStart.AddDate(0, 0, 1) // 明天开始时间

	// 查询今日每小时的录音记录数量 - 使用上海时区
	query := s.db.WithContext(ctx).Model(&model.Recording{}).
		Select("HOUR(CONVERT_TZ(FROM_UNIXTIME(recordings.created_at/1000), '+00:00', '+08:00')) as hour, COUNT(*) as count").
		Where("recordings.created_at >= ? AND recordings.created_at < ?", todayStart.UnixMilli(), todayEnd.UnixMilli())

	if storeID != nil {
		query = query.Joins("JOIN devices ON recordings.mac_address = devices.mac_address").
			Where("devices.store_id = ?", *storeID)
	}

	rows, err := query.Group("HOUR(CONVERT_TZ(FROM_UNIXTIME(recordings.created_at/1000), '+00:00', '+08:00'))").
		Order("hour ASC").
		Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 创建24小时的映射，确保所有小时都有数据
	hourMap := make(map[int]int64)
	for i := 0; i < 24; i++ {
		hourMap[i] = 0
	}

	// 填充实际数据
	for rows.Next() {
		var hour int
		var count int64
		if err := rows.Scan(&hour, &count); err != nil {
			return nil, err
		}
		hourMap[hour] = count
	}

	// 转换为结果数组
	for i := 0; i < 24; i++ {
		results = append(results, HourlyRecordCount{
			Hour:  i,
			Count: hourMap[i],
		})
	}

	return results, nil
}

// 获取今日关键词触发次数
// 根据storeID过滤关键词触发记录
func (s *stats) GetTodayKeywordTriggers(ctx context.Context, storeID *int64) (int64, error) {
	var count int64

	// 获取今日开始时间 - 使用上海时区
	now := time.Now().In(shanghaiLocation)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, shanghaiLocation)
	todayEnd := todayStart.AddDate(0, 0, 1) // 明天开始时间

	// 查询今日关键词触发次数
	query := s.db.WithContext(ctx).Model(&model.KeywordMatch{}).
		Joins("JOIN recordings ON keyword_matches.recording_id = recordings.id").
		Joins("JOIN devices ON recordings.mac_address = devices.mac_address").
		Where("keyword_matches.created_at >= ? AND keyword_matches.created_at < ?", todayStart.UnixMilli(), todayEnd.UnixMilli())

	if storeID != nil && *storeID != 0 {
		query = query.Where("devices.store_id = ?", *storeID)
	}

	err := query.Count(&count).Error
	return count, err
}

// 获取今日关键词匹配详情
// 返回所有关键词列表，不进行统计和门店过滤
func (s *stats) GetTodayKeywordMatches(ctx context.Context, storeID *int64) ([]KeywordMatchStats, error) {
	var results []KeywordMatchStats

	// 获取所有关键词
	var keywords []model.Keyword
	if err := s.db.WithContext(ctx).Model(&model.Keyword{}).Find(&keywords).Error; err != nil {
		return nil, err
	}

	// 为所有关键词创建结果
	for _, keyword := range keywords {
		results = append(results, KeywordMatchStats{
			KeywordID:   keyword.ID,
			Keyword:     keyword.Keyword,
			MarkColor:   keyword.MarkColor,
			MatchCount:  0,
			RecordCount: 0,
		})
	}

	return results, nil
}
