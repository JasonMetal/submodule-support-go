package testOne

import (
	"github.com/gin-gonic/gin"
	"idea-go/app/http/controllers"
)

type TestController struct {
	controllers.BaseController
}

func NewTestOne(ctx *gin.Context) *TestController {
	return &TestController{
		controllers.NewBaseBaseController(ctx),
	}
}

func (tc *TestController) GetTest1() {
	tc.Success("GetTest1")
	return
}

func (tc *TestController) UpdateTest1() {

	return
}
