package router

import "github.com/gofiber/fiber/v2"

type Deps struct {
}

func RegisterRoutes(app *fiber.App, deps Deps) {
	// Register your routes here

	// TODO: 全局中间件

	v1 := app.Group("/v1")
	// 注册 Feed 路由
	RegFeedRoutes(v1)
	// 注册 Media 路由
	RegMediaRoutes(app)
}
