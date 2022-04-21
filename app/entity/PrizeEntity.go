package entity

import "time"

type PrizeData struct {
	Id     uint32  `json:"id"`
	Name   string  `json:"name"`
	Img    string  `json:"img"`
	Claim  string  `json:"claim"`
	Level  uint32  `json:"level"`
	Rule   string  `json:"rule"`
	Rate   float32 `json:"rate"`
	Number uint32  `json:"number"`
}

type HudongOnlineGamePlayRecords struct {
	ID          uint32
	RoundID     uint32
	Openid      string
	CreateTime  time.Time
	score       uint32
	UserPrizeID uint32 `gorm:"column:user_prize_id"`
}

type HudongOnlineGameUserPrize struct {
	ID                          uint32                      `gorm:"column:id" json:"id"`
	OpenID                      string                      `gorm:"column:open_id" json:"openid"`
	HudongOnlineGamePlayRecords HudongOnlineGamePlayRecords `gorm:"foreignKey:user_prize_id"`
}
