package testOne

import (
	"github.com/gin-gonic/gin"
	"idea-go/app/http/controller"
)

type TestController struct {
	controller.BaseController
}

func NewTestOne(ctx *gin.Context) *TestController {
	return &TestController{
		controller.NewBaseBaseController(ctx),
	}
}

func (tc *TestController) GetTest1() {

	tc.Success("GetTest1")
	return
}

func (tc *TestController) UpdateTest1() {

	return
}
