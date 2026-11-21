package migration

import (
	infraaccount "GCFeed-sy/internal/infra/persistence/account"
	infrafeed "GCFeed-sy/internal/infra/persistence/feed"
	infravideo "GCFeed-sy/internal/infra/persistence/video"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	// 1.自动迁移表结构
	err := autoMigrateModels(db)
	if err != nil {
		return err
	}
	// 2.为所有视频表补充一条统计数据
	if err := infravideo.EnsureStats(db); err != nil {
		return err
	}
	// 3.补充一条索引
	err = infrafeed.EnsureTimelineIndex(db)
	return err
}

func autoMigrateModels(db *gorm.DB) error {
	var err error
	for attempt := range 4 {
		err = db.AutoMigrate(
			&infraaccount.UserModel{},
		)
		if err == nil {
			return nil
		}
		if !isConcurrentMigrationError(err) {
			return err
		}
		time.Sleep(time.Duration(attempt+1) * 150 * time.Millisecond)
	}
	return err
}

// 判断错误是否为并发迁移相关错误
func isConcurrentMigrationError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if !errors.As(err, &mysqlErr) {
		return false // 不是 MySQL 错误，直接返回 false
	}
	switch mysqlErr.Number { // 1060列名重复,1061 索引重复
	case 1060, 1061:
		return true
	default:
		return false
	}
}
