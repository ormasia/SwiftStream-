package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ormasia/swiftstream/internal/oss/handlers"
	"github.com/ormasia/swiftstream/internal/oss/middleware"
)

func RegisterRoutes(app *fiber.App, handlers handlers.Handlers) {
	// OSS 分片上传路由
	oss := app.Group("/api/oss", middleware.Cors())

	// 初始化分片上传
	oss.Post("/upload/init", handlers.Init)

	// 上传分片
	oss.Post("/upload/:uploadid/chunk/:chunkIndex", handlers.Chunk)

	// 完成上传
	oss.Post("/upload/:uploadid/complete", handlers.Complete)

	// 查询上传状态
	oss.Get("/upload/:uploadid/status", handlers.Status)
}
