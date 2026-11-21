package interfaceshttprouter

import (
	infraconfig "GCFeed-sy/internal/infra/config"
	"GCFeed-sy/internal/infra/migration"
	"database/sql"

	"github.com/gin-gonic/gin"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Register 负责后端依赖装配：数据库模型、仓储、Service、Handler、中间件和路由。
func Register(g *gin.Engine, cfg *infraconfig.Config, db *sql.DB) error {

	// database/sql 连接池交给 GORM 复用，避免维护两套数据库连接。
	gormDB, err := gorm.Open(gormmysql.New(gormmysql.Config{
		Conn: db,
	}))
	if err != nil {
		return err
	}

	// AutoMigrate 根据模型创建或补齐表结构，适合教学项目快速启动。
	err = migration.AutoMigrate(gormDB)
	if err != nil {
		return err
	}

	g.GET("/health", HealthCheck)

	return nil
}

// HealthCheck 提供基础健康检查接口，方便本地调试和容器探活。
func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "All is ok",
	})
}
