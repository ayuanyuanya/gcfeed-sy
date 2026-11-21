package migration

import (
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	err := autoMigrateModels(db)
	if err != nil {
		return err
	}

	return nil
}

func autoMigrateModels(db *gorm.DB) error {
	var err error
	for attempt := range 4 {
		err = db.AutoMigrate()
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
