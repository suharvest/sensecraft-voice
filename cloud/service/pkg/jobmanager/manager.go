package jobmanager

import (
	"github.com/robfig/cron/v3"

	logutil "github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/util/log"
)

type Job interface {
	// Name returns the job name
	Name() string
	// CronSpec returns the cron expression of the job
	// e.g. "* * * * *"
	CronSpec() string
	// Do is the job handler
	Do(ctx *JobContext) error
}

type Manager struct {
	cron *cron.Cron
}

func NewManager(lc *logutil.LogOptions, jobs ...Job) *Manager {
	c := cron.New()
	for _, job := range jobs {
		c.AddFunc(job.CronSpec(), func() {
			ctx := NewJobContext(job.Name(), lc)
			ctx.Log(job.Do(ctx))
		})
	}
	return &Manager{
		c,
	}
}

func (m *Manager) Run() {
	m.cron.Start()
}

func (m *Manager) Stop() {
	ctx := m.cron.Stop()
	<-ctx.Done()
}
