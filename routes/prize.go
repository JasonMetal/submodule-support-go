package router

import (
	"github.com/gin-gonic/gin"
	"idea-go/app/http/controller/prize"
	"idea-go/app/http/middleware"
)

func RegisterPrize(router *gin.Engine) {

	prizeRoute := router.Group("/v2/prize")
	{
		// 使用中间件
		prizeRoute.Use(middleware.ParseRoundId())
		prizeRoute.GET("/getPrizeList", func(ctx *gin.Context) {
			prize.NewPrizeController(ctx).GetList()
		})

		// 不使用中间件
		prizeRoute.GET("/detail", func(ctx *gin.Context) {
			prize.NewPrizeController(ctx).Detail()
		})

	}

}
