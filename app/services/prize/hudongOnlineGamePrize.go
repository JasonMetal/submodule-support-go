package prize

import (
	"context"
	"idea-go/app/entity"
	"idea-go/app/models/manyideacloud/prize"
)

type prizeService struct {
	ctx context.Context
}

func NewPrizeService(ctx context.Context) *prizeService {
	return &prizeService{ctx}
}

func (ps *prizeService) GetByRid(rid uint32) *entity.PrizeData {

	return prize.NewPrize(ps.ctx).GetByRid(rid)
}
