package jobmanager

import (
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	logutil "github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/util/log"
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
	cron   *cron.Cron
	jobs   []Job
	logCfg *logutil.LogOptions
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
		cron:   c,
		jobs:   jobs,
		logCfg: lc,
	}
}

func (m *Manager) Run() {
	// 启动时立即执行所有任务
	m.runJobsImmediately()

	// 启动 cron 调度
	m.cron.Start()
}

func (m *Manager) Stop() {
	ctx := m.cron.Stop()
	<-ctx.Done()
}

// runJobsImmediately 启动时立即执行所有任务
func (m *Manager) runJobsImmediately() {
	logrus.Info("启动时立即执行所有任务")

	for _, job := range m.jobs {
		go func(job Job) {
			logrus.Infof("立即执行任务: %s", job.Name())
			ctx := NewJobContext(job.Name(), m.logCfg)
			if err := job.Do(ctx); err != nil {
				logrus.Errorf("启动时执行任务 %s 失败: %v", job.Name(), err)
			} else {
				logrus.Infof("启动时执行任务 %s 成功", job.Name())
			}
		}(job)
	}
}

// GetJob 根据名称获取任务
func (m *Manager) GetJob(name string) Job {
	for _, job := range m.jobs {
		if job.Name() == name {
			return job
		}
	}
	return nil
}

// NewJobContext 创建任务上下文
func (m *Manager) NewJobContext(name string) *JobContext {
	return NewJobContext(name, m.logCfg)
}
