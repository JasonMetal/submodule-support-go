package prize

import (
	"context"
	"idea-go/app/entity"
	"idea-go/app/services/prize"
)

type Prize struct {
	ctx context.Context
}

func NewPrizeLogic(ctx context.Context) *Prize {
	return &Prize{ctx}
}

func (p *Prize) GetPrizeList(rid uint32) *entity.PrizeData {

	return prize.NewPrizeService(p.ctx).GetByRid(rid)

}
