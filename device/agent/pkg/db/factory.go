package db

import (
	"gorm.io/gorm"
)

type ShareDaoFactory interface {
	Audit() AuditInterface
}

type shareDaoFactory struct {
	db *gorm.DB
}

func (f *shareDaoFactory) Audit() AuditInterface { return newAudit(f.db) }

func NewDaoFactory(db *gorm.DB, migrate bool) (ShareDaoFactory, error) {
	if migrate {
		// 自动创建指定模型的数据库表结构
		if err := newMigrator(db).AutoMigrate(); err != nil {
			return nil, err
		}
	}

	return &shareDaoFactory{
		db: db,
	}, nil
}
