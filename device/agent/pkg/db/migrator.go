package db

import (
	"github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/db/model"

	"gorm.io/gorm"
)

type migrator struct {
	db *gorm.DB
}

// AutoMigrate 自动创建指定模型的数据库表结构
func (m *migrator) AutoMigrate() error {
	return m.CreateTables(model.GetMigrationModels()...)
}

func (m *migrator) CreateTables(dst ...interface{}) error {
	db := m.db.Set("gorm:table_options", "AUTO_INCREMENT=20220801 DEFAULT CHARSET=utf8")

	for _, d := range dst {
		if db.Migrator().HasTable(d) {
			continue
		}
		if err := db.Migrator().CreateTable(d); err != nil {
			return err
		}
	}
	return nil
}

func newMigrator(db *gorm.DB) *migrator {
	return &migrator{db}
}
