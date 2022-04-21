package prize

import (
	"context"
	"fmt"
	"idea-go/app/entity"
	"idea-go/app/service/prize"
)

type Prize struct {
	ctx context.Context
}

func NewPrizeLogic(ctx context.Context) *Prize {
	return &Prize{ctx}
}

func (p *Prize) GetPrizeList(rid uint32) *entity.PrizeData {

	record := prize.NewPrizeService(p.ctx).GetPlayRecord(21841)
	fmt.Println(record)
	return prize.NewPrizeService(p.ctx).GetByRid(rid)

}
