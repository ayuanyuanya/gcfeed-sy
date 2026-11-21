package main

import (
	infraconfig "GCFeed-sy/internal/infra/config"
	infradatabase "GCFeed-sy/internal/infra/database"
	infrahttpgin "GCFeed-sy/internal/infra/httpgin"
	interfaceshttprouter "GCFeed-sy/internal/interface/router/http"
	"log"
)

const configPath string = "./configs/config.yaml"

func main() {
	// 启动顺序保持简单：配置 -> 数据库 -> Gin -> 路由 -> 启动服务。
	config, err := infraconfig.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("load config.yaml failed: %v", err)
	}
	log.Printf("load config.yaml success")
	log.Printf("config loaded: port=%d database=%s:%d/%s",
		config.Port,
		config.Database.Host,
		config.Database.Port,
		config.Database.Name)

	// 数据库连接使用 database/sql 连接池，后续会被 GORM 复用。
	db, err := infradatabase.New(config.Database)
	if err != nil {
		log.Fatalf("create db failed: %v", err)
	}

	// Gin 引擎只负责 HTTP 入口，业务依赖在 router.Register 中装配。
	g := infrahttpgin.Init()

	// router.Register 会完成仓储、Service、Handler 和中间件的组装。
	err = interfaceshttprouter.Register(g, config, db)
	if err != nil {
		log.Fatalf("register failed: %v", err)
	}
	log.Printf("register success...")

	// Run 会阻塞当前进程，直到服务器停止或启动失败。
	log.Printf("server is running...")
	err = infrahttpgin.Run(config, g)
	if err != nil {
		log.Fatalf("start server failed: %v", err)
	}
}
