package main

import (
	"time"

	"github.com/gofiber/fiber/v2"
	sqlite "github.com/ormasia/swiftstream/internal/common/db"
	osshandlers "github.com/ormasia/swiftstream/internal/oss/handlers"
	ossrepo "github.com/ormasia/swiftstream/internal/oss/repo"
	ossrouters "github.com/ormasia/swiftstream/internal/oss/router"
)

func main() {
	app := fiber.New(fiber.Config{
		AppName:      "OSS Service",
		BodyLimit:    50 * 1024 * 1024, // 50MB 请求体限制
		ReadTimeout:  60 * time.Second, // 60秒读取超时
		WriteTimeout: 60 * time.Second, // 60秒写入超时
	})
	// 初始化SQLite数据库
	cfg := sqlite.SQLiteCfg{
		Path:         "data/swift.db",
		MaxOpenConns: 10,
		MaxIdleConns: 5,
		ConnMaxLife:  2 * 60 * 60, // 2 hours
	}
	db, err := sqlite.ConnectDB(cfg)
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}

	// 自动迁移数据库表
	if err := ossrepo.CreateTable(db); err != nil {
		panic("Failed to migrate database: " + err.Error())
	}

	handlers := osshandlers.NewHandlers(db)
	// 注册OSS路由
	ossrouters.RegisterRoutes(app, *handlers)

	// 启动服务器
	if err := app.Listen(":8080"); err != nil {
		panic("Failed to start server: " + err.Error())
	}
}
