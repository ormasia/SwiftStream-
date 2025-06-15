package db

import (
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type SQLiteCfg struct {
	Path         string        // "data/swift.db"
	MaxOpenConns int           // 10
	MaxIdleConns int           // 5
	ConnMaxLife  time.Duration // "2h"
}

func ConnectDB(cfg SQLiteCfg) (*gorm.DB, error) {
	gcfg := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(sqlite.Open(
		fmt.Sprintf("file:%s?cache=shared&mode=rwc&_journal_mode=WAL&_foreign_keys=on",
			cfg.Path)),
		gcfg)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLife)
	return db, nil
}
