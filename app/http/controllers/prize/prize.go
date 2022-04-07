package prize

import (
	"github.com/gin-gonic/gin"
	"idea-go/app/http/controllers"
	"idea-go/app/logic/prize"
	"strconv"
)

type PrizeController struct {
	controllers.BaseController
}

func NewPrizeController(ctx *gin.Context) *PrizeController {
	return &PrizeController{
		controllers.NewBaseBaseController(ctx),
	}
}

func (p *PrizeController) GetList() {
	prizeLogic := prize.NewPrizeLogic(p.GCtx)
	rid, err := strconv.ParseInt(p.GetQueryDefault("rid", "0").Val, 10, 32)
	if err != nil {
		rid = 0
	}

	ranking := prizeLogic.GetPrizeList(uint32(rid))
	p.Success(ranking)
}
