package types

import "github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"

// CreateKeywordRequest 创建关键词请求
type CreateKeywordRequest struct {
	Keyword   string `json:"keyword" binding:"required,min=1,max=50"`
	Synonyms  string `json:"synonyms" binding:"required,min=1,max=500"`
	MarkColor string `json:"mark_color" binding:"required,len=7"`
}

// UpdateKeywordRequest 更新关键词请求
type UpdateKeywordRequest struct {
	Keyword   string `json:"keyword,omitempty" binding:"omitempty,min=1,max=50"`
	Synonyms  string `json:"synonyms,omitempty" binding:"omitempty,min=1,max=500"`
	MarkColor string `json:"mark_color,omitempty" binding:"omitempty,len=7"`
}

// ListKeywordsRequest 获取关键词列表请求
type ListKeywordsRequest struct {
	Offset    int    `form:"offset" binding:"min=0"`
	Limit     int    `form:"limit" binding:"min=1,max=100"`
	Keyword   string `form:"keyword"`
	MarkColor string `form:"mark_color"`
}

// ListKeywordsResponse 获取关键词列表响应
type ListKeywordsResponse struct {
	Total  int64            `json:"total"`
	Items  []*model.Keyword `json:"items"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
}

// BatchDeleteKeywordsRequest 批量删除关键词请求
type BatchDeleteKeywordsRequest struct {
	IDs []int64 `json:"ids" binding:"required,min=1"`
}

// BatchDeleteKeywordsResponse 批量删除关键词响应
type BatchDeleteKeywordsResponse struct {
	DeletedCount int64   `json:"deleted_count"`
	DeletedIDs   []int64 `json:"deleted_ids"`
}
