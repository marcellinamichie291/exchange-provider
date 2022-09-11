package uniswapv3

import (
	"fmt"
	"math/big"
	"order_service/internal/delivery/exchanges/uniswap/v3/contracts"
	"order_service/internal/entity"
	"order_service/pkg/errors"
	"order_service/pkg/utils/numbers"

	"github.com/ethereum/go-ethereum/common"
)

func (u *UniSwapV3) Withdrawal(o *entity.UserOrder, coin *entity.Coin, a *entity.Address, vol string) (string, error) {
	agent := u.agent("Withdrawal")

	t, err := u.tokens.get(coin.CoinId)
	if err != nil {
		return "", err
	}

	value, err := numbers.FloatStringToBigInt(vol, t.Decimals)
	if err != nil {
		return "", err
	}

	contract, err := contracts.NewMain(t.Address, u.dp)
	if err != nil {
		return "", err
	}

	sender := common.HexToAddress(o.Deposit.Addr)
	reciever := common.HexToAddress(o.Withdrawal.Addr)

	// unwrap
	if t.isNative() {

		u.l.Debug(agent, fmt.Sprintf("unwrapping `%s` WETH", vol))
		// TODO: check if it's wrapped before.
		opts, err := u.newKeyedTransactorWithChainID(sender, big.NewInt(0))
		if err != nil {
			return "", err
		}

		tx, err := contract.Withdraw(opts, value)
		if err != nil {
			u.wallet.ReleaseNonce(t.Address, opts.Nonce.Uint64())
			return "", err
		}
		u.wallet.BurnNonce(t.Address, tx.Nonce())

		o.MetaData["unwrapWETH"] = tx.Hash().String()

		done := make(chan struct{})
		tf := &ttFeed{
			txHash:   tx.Hash(),
			receiver: &t.Address,
			needTx:   true,

			doneCh: done,
		}
		u.tt.push(tf)
		<-done

		switch tf.status {
		case txFailed:
			return "", errors.Wrap(errors.NewMesssage(fmt.Sprintf("unwrapWETH tx `%s` failed (%s)", tx.Hash(), tf.faildesc)))
		case txSuccess:
			u.l.Debug(agent, fmt.Sprintf("unwrapping `%v` WETH was successful", vol))
			o.Withdrawal.ExchangeFee = computeTxFee(tf.tx.GasPrice(), tf.Receipt.GasUsed)
			o.Withdrawal.ExchangeFeeCurrency = ether
		}

		tx, err = u.transferEth(sender, reciever, value)
		if err != nil {
			return "", err
		}
		o.Withdrawal.Executed = numbers.BigIntToFloatString(tx.Value(), t.Decimals)
		return tx.Hash().String(), nil
	}

	opts, err := u.newKeyedTransactorWithChainID(sender, big.NewInt(0))
	if err != nil {
		return "", err
	}
	tx, err := contract.Transfer(opts, reciever, value)
	if err != nil {
		u.wallet.ReleaseNonce(sender, opts.Nonce.Uint64())
		return "", err
	}

	u.wallet.BurnNonce(sender, tx.Nonce())
	o.Withdrawal.TxId = tx.Hash().String()
	return tx.Hash().String(), nil
}

func (u *UniSwapV3) TrackWithdrawal(w *entity.Withdrawal, done chan<- struct{},
	proccessedCh <-chan bool) {

	t, err := u.tokens.get(w.CoinId)
	if err != nil {
		w.Status = entity.WithdrawalFailed
		w.FailedDesc = err.Error()
		done <- struct{}{}
		<-proccessedCh
		return
	}

	var r common.Address
	if t.isNative() {
		r = common.HexToAddress(w.Addr)
	} else {
		r = t.Address
	}

	doneCh := make(chan struct{})
	tf := &ttFeed{
		txHash:   common.HexToHash(w.WId),
		receiver: &r,
		needTx:   true,
		doneCh:   doneCh,
	}
	u.tt.push(tf)
	<-doneCh

	switch tf.status {
	case txSuccess:
		f := computeTxFee(tf.tx.GasPrice(), tf.Receipt.GasUsed)
		fee, _ := numbers.FloatStringToBigInt(f, ethDecimals)

		unwrapFee := new(big.Int)
		var err error
		if w.ExchangeFee != "" {
			unwrapFee, err = numbers.FloatStringToBigInt(w.ExchangeFee, ethDecimals)
			if err != nil {
				unwrapFee = big.NewInt(0)
			}
		}
		w.ExchangeFee = numbers.BigIntToFloatString(new(big.Int).Add(fee, unwrapFee), ethDecimals)
		w.ExchangeFeeCurrency = ether
		w.Status = entity.WithdrawalSucceed

	default:
		w.Status = entity.WithdrawalFailed
		w.FailedDesc = tf.faildesc
	}

	done <- struct{}{}
	<-proccessedCh

}
