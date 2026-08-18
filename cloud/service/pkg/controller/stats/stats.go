package stats

import (
	"context"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/errors"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/config"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/util/log"
)

type StatsGetter interface {
	Stats() Interface
}

type Interface interface {
	GetDashboardStats(ctx context.Context, storeID *int64) (*DashboardStats, error)
}

type DashboardStats struct {
	TotalRecords            int64               `json:"total_records"`
	TotalDevices            int64               `json:"total_devices"`
	TotalStores             int64               `json:"total_stores"`
	TotalUsers              int64               `json:"total_users"`
	TotalLocation           int64               `json:"total_location"`
	WeeklyRecordTrend       []WeeklyRecordCount `json:"weekly_record_trend"`
	TodayActiveDevices      int64               `json:"today_active_devices"`
	TodayHourlyDistribution []HourlyRecordCount `json:"today_hourly_distribution"`
	TodayKeywordTriggers    int64               `json:"today_keyword_triggers"`
	TodayKeywordMatches     []KeywordMatchStats `json:"today_keyword_matches"`
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
	cc      config.Config
	factory db.ShareDaoFactory
}

func (s *stats) GetDashboardStats(ctx context.Context, storeID *int64) (*DashboardStats, error) {
	// 并行获取各种统计数据
	type result struct {
		totalRecords            int64
		totalDevices            int64
		totalStores             int64
		totalUsers              int64
		totalLocation           int64
		weeklyTrend             []db.WeeklyRecordCount
		todayActiveDevices      int64
		todayHourlyDistribution []db.HourlyRecordCount
		todayKeywordTriggers    int64
		todayKeywordMatches     []db.KeywordMatchStats
		err                     error
	}

	ch := make(chan result, 1)

	go func() {
		var res result

		// 获取录音记录总数
		res.totalRecords, res.err = s.factory.Stats().GetTotalRecords(ctx, storeID)
		if res.err != nil {
			log.Errorf("获取录音记录总数失败: %v, storeID: %v", res.err, storeID)
			ch <- res
			return
		}

		// 获取设备总数
		res.totalDevices, res.err = s.factory.Stats().GetTotalDevices(ctx, storeID)
		if res.err != nil {
			log.Errorf("获取设备总数失败: %v, storeID: %v", res.err, storeID)
			ch <- res
			return
		}

		// 获取门店总数
		res.totalStores, res.err = s.factory.Stats().GetTotalStores(ctx)
		if res.err != nil {
			log.Errorf("获取门店总数失败: %v", res.err)
			ch <- res
			return
		}

		// 获取用户总数
		res.totalUsers, res.err = s.factory.Stats().GetTotalUsers(ctx, storeID)
		if res.err != nil {
			log.Errorf("获取用户总数失败: %v, storeID: %v", res.err, storeID)
			ch <- res
			return
		}

		// 获取点位总数
		res.totalLocation, res.err = s.factory.Stats().GetTotalLocations(ctx, storeID)
		if res.err != nil {
			log.Errorf("获取点位总数失败: %v, storeID: %v", res.err, storeID)
			ch <- res
			return
		}

		// 获取本周录音记录增长趋势
		res.weeklyTrend, res.err = s.factory.Stats().GetWeeklyRecordTrend(ctx, storeID)
		if res.err != nil {
			log.Errorf("获取本周录音记录增长趋势失败: %v, storeID: %v", res.err, storeID)
			ch <- res
			return
		}

		// 获取今日活跃设备数量
		res.todayActiveDevices, res.err = s.factory.Stats().GetTodayActiveDevices(ctx, storeID)
		if res.err != nil {
			log.Errorf("获取今日活跃设备数量失败: %v, storeID: %v", res.err, storeID)
			ch <- res
			return
		}

		// 获取今日每小时分布
		res.todayHourlyDistribution, res.err = s.factory.Stats().GetTodayHourlyDistribution(ctx, storeID)
		if res.err != nil {
			log.Errorf("获取今日每小时分布失败: %v, storeID: %v", res.err, storeID)
			ch <- res
			return
		}

		// 获取今日关键词触发次数
		res.todayKeywordTriggers, res.err = s.factory.Stats().GetTodayKeywordTriggers(ctx, storeID)
		if res.err != nil {
			log.Errorf("获取今日关键词触发次数失败: %v, storeID: %v", res.err, storeID)
			ch <- res
			return
		}

		// 获取今日关键词匹配详情
		res.todayKeywordMatches, res.err = s.factory.Stats().GetTodayKeywordMatches(ctx, storeID)
		if res.err != nil {
			log.Errorf("获取今日关键词匹配详情失败: %v, storeID: %v", res.err, storeID)
		}

		ch <- res
	}()

	res := <-ch
	if res.err != nil {
		return nil, errors.ErrServerInternal
	}

	// 转换 WeeklyRecordCount 格式
	var weeklyTrend []WeeklyRecordCount
	for _, item := range res.weeklyTrend {
		weeklyTrend = append(weeklyTrend, WeeklyRecordCount{
			Date:  item.Date,
			Count: item.Count,
		})
	}

	// 转换 HourlyRecordCount 格式
	var todayHourlyDistribution []HourlyRecordCount
	for _, item := range res.todayHourlyDistribution {
		todayHourlyDistribution = append(todayHourlyDistribution, HourlyRecordCount{
			Hour:  item.Hour,
			Count: item.Count,
		})
	}

	// 转换 KeywordMatchStats 格式
	var todayKeywordMatches []KeywordMatchStats
	for _, item := range res.todayKeywordMatches {
		todayKeywordMatches = append(todayKeywordMatches, KeywordMatchStats{
			KeywordID:   item.KeywordID,
			Keyword:     item.Keyword,
			MarkColor:   item.MarkColor,
			MatchCount:  item.MatchCount,
			RecordCount: item.RecordCount,
		})
	}

	return &DashboardStats{
		TotalRecords:            res.totalRecords,
		TotalDevices:            res.totalDevices,
		TotalStores:             res.totalStores,
		TotalUsers:              res.totalUsers,
		TotalLocation:           res.totalLocation,
		WeeklyRecordTrend:       weeklyTrend,
		TodayActiveDevices:      res.todayActiveDevices,
		TodayHourlyDistribution: todayHourlyDistribution,
		TodayKeywordTriggers:    res.todayKeywordTriggers,
		TodayKeywordMatches:     todayKeywordMatches,
	}, nil
}

func NewStats(cfg config.Config, f db.ShareDaoFactory) *stats {
	return &stats{cc: cfg, factory: f}
}
