package prize

import (
	"context"
	"idea-go/app/entity"
	"idea-go/app/model/common"
	"idea-go/bootstrap"
)

type Prize struct {
	Ctx context.Context
}

func NewPrize(ctx context.Context) *Prize {
	return &Prize{ctx}
}

func (p Prize) getTable() string {
	return "aikehou_hudong_online_game_prize"
}

// GetByRid 根据rid获取prize数据
func (p *Prize) GetByRid(rid uint32) *entity.PrizeData {
	data := &entity.PrizeData{}
	err := common.Manyideacloud(p.Ctx).DB.
		Table(p.getTable()).Select("*").
		Where("round_id=? and number>0 and status=1", rid).
		Find(&data).Error

	if err != nil {

	}

	return data
}

func (p *Prize) GetPlayRecord(prizeID uint32) *entity.HudongOnlineGamePlayRecords {
	record := &entity.HudongOnlineGamePlayRecords{}
	err := common.Manyideacloud(p.Ctx).DB.
		Preload("PrizeData").
		Where("user_prize_id=?", prizeID).
		First(&record).
		Error

	if err != nil {
		bootstrap.CheckError(err)
	}

	return record
}
