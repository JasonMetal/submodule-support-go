package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"idea-go/app/http/controllers/prize"
)

func RegisterPrize(router *gin.Engine) {
	fmt.Println("Registered other router")
	v2 := router.Group("/v2")
	{
		v2.GET("/prize/getPrizeList", func(ctx *gin.Context) {
			prize.NewPrizeController(ctx).GetList()
		})

	}
}
