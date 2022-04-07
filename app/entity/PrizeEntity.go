package entity

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
