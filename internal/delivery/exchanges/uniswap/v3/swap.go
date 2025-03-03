package uniswapv3

import (
	"exchange-provider/internal/delivery/exchanges/uniswap/v3/contracts"
	"exchange-provider/pkg/utils/numbers"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func (u *dex) swap(tIn, tOut Token, value string, source, dest common.Address) (*types.Transaction, *pair, error) {

	var err error
	pool, err := u.setBestPrice(tIn, tOut)
	if err != nil {
		return nil, nil, err
	}

	amount, err := numbers.FloatStringToBigInt(value, tIn.Decimals)
	if err != nil {
		return nil, nil, err
	}

	val := big.NewInt(0)
	if tIn.isNative() {
		val = amount
	}
	opts, err := u.newKeyedTransactorWithChainID(source, val)
	if err != nil {
		return nil, nil, err
	}

	defer func() {
		if err != nil {
			u.wallet.ReleaseNonce(source, opts.Nonce.Uint64())
		} else {
			u.wallet.BurnNonce(source, opts.Nonce.Uint64())

		}
	}()

	data := [][]byte{}
	abi, err := contracts.RouteMetaData.GetAbi()
	if err != nil {
		return nil, nil, err
	}

	route, err := contracts.NewRoute(u.cfg.Router, u.provider())
	if err != nil {
		return nil, nil, err
	}

	params := contracts.IV3SwapRouterExactInputSingleParams{
		TokenIn:           tIn.Address,
		TokenOut:          tOut.Address,
		Fee:               pool.feeTier,
		Recipient:         dest,
		AmountIn:          amount,
		AmountOutMinimum:  big.NewInt(0),
		SqrtPriceLimitX96: big.NewInt(0),
	}
	es, err := abi.Pack("exactInputSingle", params)
	if err != nil {
		return nil, nil, err
	}

	data = append(data, es)

	deadline := big.NewInt(time.Now().Add(time.Minute * time.Duration(30)).Unix())
	tx, err := route.Multicall0(opts, deadline, data)
	if err != nil {
		return nil, nil, err
	}

	return tx, pool, err
}
