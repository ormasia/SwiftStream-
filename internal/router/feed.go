package router

/*
================= Feed Routes==================

在短视频／社交产品里，客户端主界面通常是一条向下（或左右）滑动的内容流。
这条流会持续根据用户 ID、兴趣画像、时间戳等动态刷新，所以后端负责生成这组
「下一批视频条目」的接口、服务，行业里普遍叫 Feed;

===============================================
*/

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ormasia/swiftstream/internal/handlers"
)

func RegFeedRoutes(app fiber.Router) {
	app.Get("/feed", handlers.GetFeed)
}
