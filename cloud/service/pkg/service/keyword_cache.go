package service

import (
	"context"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/types"
)

// CompiledKeyword 表示一个预编译的关键词及其近义词正则
type CompiledKeyword struct {
	KeywordID    int64
	Keyword      string
	ExactRegex   *regexp.Regexp
	SynonymRegex []*regexp.Regexp
}

// KeywordCache 维护预编译的关键词缓存
type KeywordCache struct {
	dao     db.ShareDaoFactory
	mu      sync.RWMutex
	items   []*CompiledKeyword
	updated time.Time
}

func NewKeywordCache(dao db.ShareDaoFactory) *KeywordCache {
	return &KeywordCache{dao: dao}
}

// 全局缓存（可选）：用于在不便显式注入时复用
var (
	globalKeywordCache     *KeywordCache
	globalKeywordCacheOnce sync.Once
)

func SetGlobalKeywordCache(cache *KeywordCache) { globalKeywordCache = cache }
func GetGlobalKeywordCache() *KeywordCache      { return globalKeywordCache }

// Refresh 全量刷新并预编译关键词缓存
func (kc *KeywordCache) Refresh(ctx context.Context) error {
	// 拉取所有关键词（放宽上限，按需可配置）
	req := &types.ListKeywordsRequest{Offset: 0, Limit: 10000}
	keywords, _, err := kc.dao.Keyword().List(ctx, req)
	if err != nil {
		return err
	}

	var compiled []*CompiledKeyword
	for _, kw := range keywords {
		ck := &CompiledKeyword{
			KeywordID:  kw.ID,
			Keyword:    kw.Keyword,
			ExactRegex: regexp.MustCompile(regexp.QuoteMeta(strings.ToLower(kw.Keyword))),
		}

		if kw.Synonyms != "" {
			for _, syn := range strings.Split(kw.Synonyms, ",") {
				s := strings.TrimSpace(syn)
				if s == "" {
					continue
				}
				ck.SynonymRegex = append(ck.SynonymRegex, regexp.MustCompile(regexp.QuoteMeta(strings.ToLower(s))))
			}
		}

		compiled = append(compiled, ck)
	}

	kc.mu.Lock()
	kc.items = compiled
	kc.updated = time.Now()
	kc.mu.Unlock()
	return nil
}

// GetCompiled 读取当前编译缓存快照
func (kc *KeywordCache) GetCompiled() ([]*CompiledKeyword, time.Time) {
	kc.mu.RLock()
	defer kc.mu.RUnlock()
	return kc.items, kc.updated
}
