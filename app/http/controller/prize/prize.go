package prize

import (
	"github.com/gin-gonic/gin"
	"idea-go/app/http/controller"
	"idea-go/app/logic/prize"
	"strconv"
)

type prizeController struct {
	controller.BaseController
}

func NewPrizeController(ctx *gin.Context) *prizeController {
	return &prizeController{
		controller.NewBaseBaseController(ctx),
	}
}

func (p *prizeController) GetList() {
	prizeLogic := prize.NewPrizeLogic(p.GCtx)
	rid, err := strconv.ParseInt(p.GetQueryDefault("rid", "0").Val, 10, 32)
	if err != nil {
		rid = 0
	}

	ranking := prizeLogic.GetPrizeList(uint32(rid))
	p.Success(ranking)
}

func (p *prizeController) Detail() {

}
