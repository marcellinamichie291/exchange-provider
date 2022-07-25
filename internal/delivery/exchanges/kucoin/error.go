package kucoin

import (
	"fmt"
	"order_service/pkg/errors"

	"github.com/Kucoin/kucoin-go-sdk"
)

func handleSDKErr(err error, res *kucoin.ApiResponse) error {

	if err != nil {
		return errors.Wrap(err, "kucoin-sdk", errors.ErrInternal)
	}

	if res != nil && res.Code != "200000" {
		return errors.Wrap(errors.New(fmt.Sprintf("%s:%s:%s", res.Message, res.Code, err)), "kucoin-sdk", errors.ErrInternal)
	}

	return nil

}
