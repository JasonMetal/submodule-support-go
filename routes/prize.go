package router

import (
	"github.com/gin-gonic/gin"
	"idea-go/app/http/controller/prize"
	"idea-go/app/http/middleware"
)

func RegisterPrize(router *gin.Engine) {

	prizeRoute := router.Group("/v2")
	{
		// 使用中间件
		prizeRouteWithMiddleware := prizeRoute.Use(middleware.ParseRoundId())
		prizeRouteWithMiddleware.GET("/getPrizeList", func(ctx *gin.Context) {
			prize.NewPrizeController(ctx).GetList()
		})

		// 不使用中间件
		prizeRoute.GET("/detail", func(ctx *gin.Context) {
			prize.NewPrizeController(ctx).Detail()
		})

	}

}
