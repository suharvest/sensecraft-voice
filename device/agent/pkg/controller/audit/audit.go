package audit

import (
	"context"

	"k8s.io/klog/v2"

	"github.com/YOUR-ORG/sensecraft-voice/device/agent/api/server/errors"
	"github.com/YOUR-ORG/sensecraft-voice/device/agent/cmd/app/config"
	"github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/db"
	"github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/db/model"
	"github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/types"
)

type AuditGetter interface {
	Audit() Interface
}

type Interface interface {
	List(ctx context.Context) ([]types.Audit, error)
	Get(ctx context.Context, aid int64) (*types.Audit, error)
}

type audit struct {
	cc      config.Config
	factory db.ShareDaoFactory
}

func (a *audit) Get(ctx context.Context, aid int64) (*types.Audit, error) {
	object, err := a.factory.Audit().Get(ctx, aid)
	if err != nil {
		klog.Errorf("failed to get audit %d: %v", aid, err)
		return nil, errors.ErrServerInternal
	}
	if object == nil {
		return nil, errors.ErrAuditNotFound
	}
	return a.model2Type(object), nil
}

func (a *audit) List(ctx context.Context) ([]types.Audit, error) {
	objects, err := a.factory.Audit().List(ctx, db.WithOrderByDesc())
	if err != nil {
		klog.Errorf("failed to get tenants: %v", err)
		return nil, errors.ErrServerInternal
	}

	var ts []types.Audit
	for _, object := range objects {
		ts = append(ts, *a.model2Type(&object))
	}
	return ts, nil
}

func (a *audit) model2Type(o *model.Audit) *types.Audit {
	return &types.Audit{
		SensecraftVoiceMeta: types.SensecraftVoiceMeta{
			Id:              o.Id,
			ResourceVersion: o.ResourceVersion,
		},
		TimeMeta: types.TimeMeta{
			GmtCreate:   o.GmtCreate,
			GmtModified: o.GmtModified,
		},
		Ip:       o.Ip,
		Action:   o.Action,
		Status:   o.Status,
		Operator: o.Operator,
		Path:     o.Path,
	}
}

func NewAudit(cfg config.Config, f db.ShareDaoFactory) *audit {
	return &audit{
		cc:      cfg,
		factory: f,
	}
}
