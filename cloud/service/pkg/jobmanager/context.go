package jobmanager

import (
	"context"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db"
	logutil "github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/util/log"
)

type JobContext struct {
	context.Context
	*logutil.Logger
}

func NewJobContext(name string, cfg *logutil.LogOptions) *JobContext {
	jc := &JobContext{
		Context: db.WithDBContext(context.Background()),
		Logger:  logutil.NewLogger(cfg),
	}
	jc.WithLogField("job", name)
	return jc
}

func (c *JobContext) Log(err error) {
	c.Logger.Log(c.Context, err)
}
