package kucoin

import (
	"order_service/internal/entity"
	"order_service/pkg/logger"
	"sync"

	"order_service/pkg/errors"

	"github.com/go-redis/redis/v9"
)

type wtFeed struct {
	w            *entity.Withdrawal
	done         chan<- struct{}
	err          chan<- error
	proccessedCh <-chan bool
}

type withdrawalTracker struct {
	feedCh chan *wtFeed
	l      logger.Logger
	c      *withdrawalCache
}

func newWithdrawalTracker(r *redis.Client, l logger.Logger) *withdrawalTracker {
	return &withdrawalTracker{
		feedCh: make(chan *wtFeed, 1024),
		l:      l,
		c:      newWithdrawalCache(r, l),
	}
}

func (t *withdrawalTracker) run(wg *sync.WaitGroup) {
	const op = errors.Op("Kucoin.WithdrawalTracker.run")
	t.l.Debug(string(op), "started")

	defer wg.Done()
	for {
		select {
		case feed := <-t.feedCh:
			func(f *wtFeed) {
				wd, err := t.c.getWithdrawal(f.w.Id)
				if err != nil {
					f.err <- errors.Wrap(err, op)
					return
				}
				switch wd.Status {
				case "SUCCESS":
					f.w.Status = entity.WithdrawalSucceed
					f.w.ExchangeFee = wd.Fee
					f.w.Executed = wd.Amount
					f.w.TxId = wd.FixTxId()
					f.done <- struct{}{}
				case "FAILURE":
					f.w.Status = entity.WithdrawalFailed
					f.done <- struct{}{}
				default:
					f.w.Status = entity.WithdrawalPending
				}

				if <-f.proccessedCh {
					if err := t.c.delWithdrawal(f.w.Id); err != nil {
						t.l.Error(string(op), errors.Wrap(err, op).Error())
					}
					if err := t.c.proccessedWithdrawal(f.w.Id); err != nil {
						t.l.Error(string(op), errors.Wrap(err, op).Error())
					}
					return
				}

			}(feed)

		}
	}
}

func (t *withdrawalTracker) track(f *wtFeed) {
	t.feedCh <- f
}
