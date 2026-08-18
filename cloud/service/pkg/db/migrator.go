package db

import (
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"

	"gorm.io/gorm"
)

type migrator struct {
	db *gorm.DB
}

// AutoMigrate 自动创建指定模型的数据库表结构
func (m *migrator) AutoMigrate() error {
	return m.CreateTables(model.GetMigrationModels()...)
}

const defaultTableOptions = "AUTO_INCREMENT=20220801 DEFAULT CHARSET=utf8"

// tableOptioner 允许 model 覆盖默认建表选项（例如需要 utf8mb4 的表）
type tableOptioner interface {
	TableOptions() string
}

func (m *migrator) CreateTables(dst ...interface{}) error {
	for _, d := range dst {
		options := defaultTableOptions
		if to, ok := d.(tableOptioner); ok {
			options = to.TableOptions()
		}
		db := m.db.Set("gorm:table_options", options)
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
