package dto

import (
	"encoding/json"
	"exchange-provider/internal/entity"
)

type deposite struct {
	Id       int64
	UserId   int64
	OrderId  int64
	Status   string
	Exchange string

	CoinId  string
	ChainId string

	TxId   string
	Volume string

	Address    string
	Tag        string
	FailedDesc string
}

func DToDto(d *entity.Deposit) *deposite {
	return &deposite{
		Id:       d.Id,
		UserId:   d.UserId,
		OrderId:  d.OrderId,
		Status:   d.Status,
		Exchange: d.Exchange,

		CoinId:  d.CoinId,
		ChainId: d.ChainId,

		TxId:   d.TxId,
		Volume: d.Volume,

		Address: d.Addr,
		Tag:     d.Tag,

		FailedDesc: d.FailedDesc,
	}
}

func (d *deposite) ToEntity() *entity.Deposit {
	return &entity.Deposit{
		Id:      d.Id,
		UserId:  d.UserId,
		OrderId: d.OrderId,

		Status:   d.Status,
		Exchange: d.Exchange,

		Coin: &entity.Coin{CoinId: d.CoinId, ChainId: d.ChainId},

		TxId:   d.TxId,
		Volume: d.Volume,

		Address: &entity.Address{Addr: d.Address, Tag: d.Tag},

		FailedDesc: d.FailedDesc,
	}
}

func (d *deposite) MarshalBinary() ([]byte, error) {
	return json.Marshal(d)
}
