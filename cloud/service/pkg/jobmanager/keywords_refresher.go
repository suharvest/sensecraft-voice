package jobmanager

import (
	"time"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/service"
)

// KeywordsRefresher 基于 cron 的关键词缓存刷新任务
type KeywordsRefresher struct {
	schedule string
	cache    *service.KeywordCache
}

func NewKeywordsRefresher(schedule string, cache *service.KeywordCache) *KeywordsRefresher {
	return &KeywordsRefresher{schedule: schedule, cache: cache}
}

func (j *KeywordsRefresher) Name() string     { return "keywords-refresher" }
func (j *KeywordsRefresher) CronSpec() string { return j.schedule }

func (j *KeywordsRefresher) Do(ctx *JobContext) error {
	start := time.Now()
	compiledBefore, last := j.cache.GetCompiled()
	ctx.WithLogFields(map[string]interface{}{
		"compiled_count_before": len(compiledBefore),
		"last_updated":          last,
	})

	if err := j.cache.Refresh(ctx); err != nil {
		return err
	}

	compiledAfter, updated := j.cache.GetCompiled()
	ctx.WithLogFields(map[string]interface{}{
		"compiled_count_after": len(compiledAfter),
		"updated":              updated,
		"elapsed_ms":           time.Since(start).Milliseconds(),
	})
	return nil
}
