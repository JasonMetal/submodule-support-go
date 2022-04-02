package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"idea-go/app/http/controllers/testOne"
)

func RegisterOther(router *gin.Engine) {
	fmt.Println("Registered other router")
	v2 := router.Group("/v2")
	{
		v2.GET("/test1/detail", func(ctx *gin.Context) {
			testOne.NewTestOne(ctx).GetTest1()
		})

		v2.POST("/test1/update", func(ctx *gin.Context) {
			testOne.NewTestOne(ctx).UpdateTest1()
		})
	}
}
