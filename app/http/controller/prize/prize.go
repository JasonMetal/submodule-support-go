package prize

import (
	"github.com/gin-gonic/gin"
	"idea-go/app/http/controller"
	"idea-go/app/logic/prize"
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
	rid := p.GetRoundId()

	ranking := prizeLogic.GetPrizeList(rid)
	p.Success(ranking)
}

func (p *prizeController) Detail() {

}
