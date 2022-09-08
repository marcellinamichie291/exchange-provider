package uniswapv3

import (
	"context"
	"math/big"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func (u *UniSwapV3) TransferEth(from, to common.Address, value *big.Int) (*types.Transaction, error) {

	head, err := u.dp.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	gasPrice, err := u.dp.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	gasTipCap, err := u.dp.SuggestGasTipCap(context.Background())
	if err != nil {
		return nil, err
	}

	gasFeeCap := new(big.Int).Add(
		gasTipCap,
		new(big.Int).Mul(head.BaseFee, big.NewInt(2)),
	)

	gas, err := u.dp.EstimateGas(context.Background(), ethereum.CallMsg{
		From:      from,
		To:        &to,
		GasPrice:  gasPrice,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Value:     value,
		Data:      []byte{},
	})
	if err != nil {
		return nil, err
	}

	nonce, err := u.wallet.Nonce(from)
	if err != nil {
		return nil, err
	}

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   u.chainId,
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gas,
		To:        &to,
		Value:     value,
		Data:      []byte{},
	})

	prvKey, err := u.wallet.PrivateKey(from)
	if err != nil {
		return nil, err
	}
	tx, err = types.SignTx(tx, types.NewLondonSigner(u.chainId), prvKey)
	err = u.dp.SendTransaction(context.Background(), tx)
	if err != nil {
		u.wallet.ReleaseNonce(from, nonce)
		return nil, err
	}

	u.wallet.BurnNonce(from, nonce)
	return tx, nil
}
