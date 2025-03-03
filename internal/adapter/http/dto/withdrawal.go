package dto

import "exchange-provider/internal/entity"

type Withdrawal struct {
	Id       uint64 `json:"id"`
	WId      string `json:"exchange_withdrawal_id"`
	OrderId  int64  `json:"order_id,omitempty"`
	UserId   int64  `json:"user_id,omitempty"`
	Exchange string `json:"exchange,omitempty"`

	Address string
	Tag     string

	Coin  string
	Chain string

	Total       string
	Fee         string
	ExchangeFee string
	Executed    string

	TxId   string
	Status string
}

func WFromEntity(w *entity.Withdrawal) *Withdrawal {
	return &Withdrawal{
		Id:  w.Id,
		WId: w.WId,

		Address: w.Addr,
		Tag:     w.Tag,

		Coin:        w.CoinId,
		Chain:       w.ChainId,
		Total:       w.Total,
		Fee:         w.Fee,
		ExchangeFee: w.ExchangeFee,
		Executed:    w.Executed,

		TxId:   w.TxId,
		Status: string(w.Status),
	}
}
