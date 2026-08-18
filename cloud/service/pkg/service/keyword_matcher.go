package service

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/types"
	"k8s.io/klog/v2"
)

// KeywordMatcher 关键词匹配服务
type KeywordMatcher struct {
	factory db.ShareDaoFactory
	cache   *KeywordCache
}

// MatchResult 匹配结果
type MatchResult struct {
	KeywordID   int64   `json:"keyword_id"`
	Keyword     string  `json:"keyword"`
	MatchedText string  `json:"matched_text"`
	MatchType   string  `json:"match_type"`
	Position    int     `json:"position"`
	Length      int     `json:"length"`
	Confidence  float64 `json:"confidence"`
}

// NewKeywordMatcher 创建关键词匹配服务
func NewKeywordMatcher(factory db.ShareDaoFactory) *KeywordMatcher {
	return &KeywordMatcher{factory: factory}
}

// NewKeywordMatcherWithCache 支持注入关键词缓存
func NewKeywordMatcherWithCache(factory db.ShareDaoFactory, cache *KeywordCache) *KeywordMatcher {
	return &KeywordMatcher{factory: factory, cache: cache}
}

// MatchKeywords 匹配文本中的关键词
func (km *KeywordMatcher) MatchKeywords(ctx context.Context, text string, macAddress string) ([]MatchResult, error) {
	var results []MatchResult
	textLower := strings.ToLower(text)

	// 优先使用缓存
	if km.cache != nil {
		compiled, _ := km.cache.GetCompiled()
		if len(compiled) == 0 {
			// 冷启动兜底
			if err := km.cache.Refresh(ctx); err != nil {
				klog.Errorf("Failed to refresh keyword cache: %v", err)
			}
			compiled, _ = km.cache.GetCompiled()
		}

		for _, ck := range compiled {
			if ck.ExactRegex != nil {
				if indices := ck.ExactRegex.FindAllStringIndex(textLower, -1); len(indices) > 0 {
					for _, index := range indices {
						results = append(results, MatchResult{
							KeywordID:   ck.KeywordID,
							Keyword:     ck.Keyword,
							MatchedText: textLower[index[0]:index[1]],
							MatchType:   model.MatchTypeExact,
							Position:    index[0],
							Length:      index[1] - index[0],
							Confidence:  1.0,
						})
					}
				}
			}
			for _, r := range ck.SynonymRegex {
				if indices := r.FindAllStringIndex(textLower, -1); len(indices) > 0 {
					for _, index := range indices {
						results = append(results, MatchResult{
							KeywordID:   ck.KeywordID,
							Keyword:     ck.Keyword,
							MatchedText: textLower[index[0]:index[1]],
							MatchType:   model.MatchTypeSynonym,
							Position:    index[0],
							Length:      index[1] - index[0],
							Confidence:  0.8,
						})
					}
				}
			}
		}
		return results, nil
	}

	// 兼容无缓存：回退到原有逻辑
	keywords, err := km.getAllKeywords(ctx)
	if err != nil {
		klog.Errorf("Failed to get keywords: %v", err)
		return nil, err
	}
	for _, keyword := range keywords {
		if matches := km.exactMatch(textLower, keyword.Keyword); len(matches) > 0 {
			for _, match := range matches {
				results = append(results, MatchResult{
					KeywordID:   keyword.ID,
					Keyword:     keyword.Keyword,
					MatchedText: match.Text,
					MatchType:   model.MatchTypeExact,
					Position:    match.Position,
					Length:      match.Length,
					Confidence:  1.0,
				})
			}
		}
		if matches := km.synonymMatch(textLower, keyword.Synonyms, keyword.Keyword); len(matches) > 0 {
			for _, match := range matches {
				results = append(results, MatchResult{
					KeywordID:   keyword.ID,
					Keyword:     keyword.Keyword,
					MatchedText: match.Text,
					MatchType:   model.MatchTypeSynonym,
					Position:    match.Position,
					Length:      match.Length,
					Confidence:  0.8,
				})
			}
		}
	}
	return results, nil
}

// SaveMatches 保存匹配结果到数据库
func (km *KeywordMatcher) SaveMatches(ctx context.Context, recordingID int64, macAddress string, results []MatchResult) error {
	if len(results) == 0 {
		return nil
	}

	var matches []*model.KeywordMatch
	for _, result := range results {
		match := &model.KeywordMatch{
			RecordingID: recordingID,
			MacAddress:  macAddress,
			KeywordID:   result.KeywordID,
			Keyword:     result.Keyword,
			MatchedText: result.MatchedText,
			MatchType:   result.MatchType,
			Confidence:  result.Confidence,
			Position:    result.Position,
			Length:      result.Length,
		}
		matches = append(matches, match)
	}

	return km.factory.KeywordMatch().BatchCreate(ctx, matches)
}

// GetMatchesByMac 根据MAC地址获取匹配记录
func (km *KeywordMatcher) GetMatchesByMac(ctx context.Context, macAddress string, limit int) ([]*model.KeywordMatch, error) {
	return km.factory.KeywordMatch().GetByMacAddress(ctx, macAddress, limit)
}

// GetMatchesByKeyword 根据关键词ID获取匹配记录
func (km *KeywordMatcher) GetMatchesByKeyword(ctx context.Context, keywordID int64, limit int) ([]*model.KeywordMatch, error) {
	return km.factory.KeywordMatch().GetByKeywordID(ctx, keywordID, limit)
}

// getAllKeywords 获取所有关键词
func (km *KeywordMatcher) getAllKeywords(ctx context.Context) ([]*model.Keyword, error) {
	// 这里可以添加缓存逻辑，避免每次都查询数据库
	req := &types.ListKeywordsRequest{
		Offset: 0,
		Limit:  1000, // 假设关键词数量不会超过1000
	}

	keywords, _, err := km.factory.Keyword().List(ctx, req)
	if err != nil {
		return nil, err
	}

	return keywords, nil
}

// TextMatch 文本匹配结果
type TextMatch struct {
	Text     string
	Position int
	Length   int
}

// exactMatch 精确匹配
func (km *KeywordMatcher) exactMatch(text, keyword string) []TextMatch {
	var matches []TextMatch
	keywordLower := strings.ToLower(keyword)

	// 使用正则表达式进行精确匹配，支持中文和英文
	pattern := regexp.MustCompile(regexp.QuoteMeta(keywordLower))
	indices := pattern.FindAllStringIndex(text, -1)

	for _, index := range indices {
		matches = append(matches, TextMatch{
			Text:     text[index[0]:index[1]],
			Position: index[0],
			Length:   index[1] - index[0],
		})
	}

	return matches
}

// synonymMatch 近义词匹配
func (km *KeywordMatcher) synonymMatch(text, synonyms, originalKeyword string) []TextMatch {
	var matches []TextMatch

	// 分割近义词
	synonymList := strings.Split(synonyms, ",")

	for _, synonym := range synonymList {
		synonym = strings.TrimSpace(synonym)
		if synonym == "" {
			continue
		}

		// 对每个近义词进行精确匹配
		synonymMatches := km.exactMatch(text, synonym)
		matches = append(matches, synonymMatches...)
	}

	return matches
}

// GetKeywordStats 获取关键词统计信息
func (km *KeywordMatcher) GetKeywordStats(ctx context.Context, macAddress string, days int) (map[int64]int, error) {
	// 计算时间范围
	startTime := time.Now().AddDate(0, 0, -days).UnixMilli()

	// 这里可以实现更复杂的统计查询
	// 暂时返回简单的计数
	matches, err := km.factory.KeywordMatch().GetByMacAddress(ctx, macAddress, 0)
	if err != nil {
		return nil, err
	}

	stats := make(map[int64]int)
	for _, match := range matches {
		if match.CreatedAt >= startTime {
			stats[match.KeywordID]++
		}
	}

	return stats, nil
}
