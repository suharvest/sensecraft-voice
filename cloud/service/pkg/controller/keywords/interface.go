package keywords

import (
	"context"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/types"
)

// KeywordGetter 关键词控制器获取器
type KeywordGetter interface {
	Keyword() Interface
}

// Interface 关键词控制器接口
type Interface interface {
	// 创建关键词
	Create(ctx context.Context, req *types.CreateKeywordRequest) (*model.Keyword, error)
	// 根据ID获取关键词
	GetById(ctx context.Context, id int64) (*model.Keyword, error)
	// 获取关键词列表
	List(ctx context.Context, req *types.ListKeywordsRequest) (*types.ListKeywordsResponse, error)
	// 更新关键词
	Update(ctx context.Context, id int64, req *types.UpdateKeywordRequest) (*model.Keyword, error)
	// 删除关键词
	Delete(ctx context.Context, id int64) error
	// 批量删除关键词
	BatchDelete(ctx context.Context, req *types.BatchDeleteKeywordsRequest) (*types.BatchDeleteKeywordsResponse, error)
}
