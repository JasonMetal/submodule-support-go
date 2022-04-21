package prize

import (
	"context"
	"idea-go/app/entity"
	"idea-go/app/model/manyideacloud/prize"
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

func (ps *prizeService) GetPlayRecord(rid uint32) *entity.HudongOnlineGamePlayRecords {

	return prize.NewPrize(ps.ctx).GetPlayRecord(rid)
}
