package entity

import (
	"fmt"
)

type WithdrawalStatus string

const (
	WithdrawalSucceed WithdrawalStatus = "succeed"
	WithdrawalFailed  WithdrawalStatus = "failed"
	WithdrawalPending WithdrawalStatus = "pending"
)

type Withdrawal struct {
	Id      uint64
	WId     string
	OrderId int64
	UserId  int64

	Status WithdrawalStatus

	*Address

	Exchange string

	*Coin

	Total       string
	Fee         string
	FeeCurrency string

	ExchangeFee         string
	ExchangeFeeCurrency string

	Executed string

	TxId       string
	FailedDesc string
}

func (w *Withdrawal) String() string {
	return fmt.Sprintf("%+v", *w)
}
