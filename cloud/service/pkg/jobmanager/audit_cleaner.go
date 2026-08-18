package jobmanager

import (
	"time"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db"
)

const (
	DefaultSchedule     = "0 0 * * 6" // 每周六 0 点执行
	DefaultDaysReserved = 30          // 保留 30 天的审计日志
)

type AuditsCleaner struct {
	cfg AuditOptions
	dao db.ShareDaoFactory
}

type AuditOptions struct {
	Schedule     string `yaml:"schedule"`
	DaysReserved int    `yaml:"days_reserved"`
}

func DefaultOptions() AuditOptions {
	return AuditOptions{
		Schedule:     DefaultSchedule,
		DaysReserved: DefaultDaysReserved,
	}
}

func NewAuditsCleaner(cfg AuditOptions, dao db.ShareDaoFactory) *AuditsCleaner {
	return &AuditsCleaner{
		cfg: cfg,
		dao: dao,
	}
}

func (ac *AuditsCleaner) Name() string {
	return "audits-cleaner"
}

func (ac *AuditsCleaner) CronSpec() string {
	return ac.cfg.Schedule
}

func (ac *AuditsCleaner) Do(ctx *JobContext) (err error) {
	resv := ac.cfg.DaysReserved
	before := time.Now().AddDate(0, 0, -resv)
	entries := map[string]interface{}{
		"days_reserved": resv,
		"deadline":      before,
	}
	entries["records_deleted"], err = ac.dao.Audit().BatchDelete(ctx, db.WithCreatedBefore(before))
	ctx.WithLogFields(entries)

	return
}

func (a *AuditOptions) Valid() error {
	// TODO
	return nil
}
