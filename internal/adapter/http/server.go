package http

import (
	"exchange-provider/internal/adapter/http/dto"
	"exchange-provider/internal/app"
	"exchange-provider/internal/entity"
	"exchange-provider/pkg/logger"
	"net/http"

	"exchange-provider/pkg/errors"

	"github.com/go-redis/redis/v9"
	"github.com/spf13/viper"
)

type Server struct {
	app *app.OrderUseCase

	l  logger.Logger
	v  *viper.Viper
	rc *redis.Client
}

func NewServer(app *app.OrderUseCase, v *viper.Viper, rc *redis.Client, l logger.Logger) *Server {
	return &Server{
		app: app,

		l:  l,
		v:  v,
		rc: rc,
	}
}

func (s *Server) NewUserOrder(ctx Context) {

	userId, _ := ctx.GetKey("user_id")
	req := dto.CreateOrderRequest{}
	if err := ctx.Bind(&req); err != nil {
		handlerErr(ctx, errors.Wrap(errors.ErrBadRequest, errors.NewMesssage("invalid request")))
		return
	}

	if err := req.Validate(); err != nil {
		handlerErr(ctx, errors.Wrap(errors.ErrBadRequest, errors.NewMesssage(err.Error())))
		return
	}

	bc, err := dto.ParseCoin(req.BC)
	if err != nil {
		handlerErr(ctx, err)
		return
	}
	qc, err := dto.ParseCoin(req.QC)
	if err != nil {
		handlerErr(ctx, err)
		return
	}

	ex, err := s.app.SelectExchangeByPair(bc, qc)
	if err != nil {
		handlerErr(ctx, err)
		return
	}

	o, err := s.app.NewUserOrder(userId.(int64), &entity.Address{Addr: req.Address, Tag: req.Tag}, bc, qc, req.Side, ex)
	if err != nil {
		handlerErr(ctx, err)
		return
	}

	var dc string
	var minD float64
	if o.Side == "buy" {
		dc = o.QC.String()
		_, minD = s.app.GetMinPairDeposit(o.BC, o.QC)
	} else {
		dc = o.BC.String()
		minD, _ = s.app.GetMinPairDeposit(o.BC, o.QC)
	}

	ctx.JSON(http.StatusOK, &dto.CreateOrderResponse{
		OrderId:         o.Seq,
		DC:              dc,
		MinDeposit:      minD,
		DepositeAddress: o.Deposit.Addr,
		AddressTag:      o.Deposit.Tag,
	})
}

func (s *Server) GetPaginatedForUser(ctx Context) {
	userId, _ := ctx.GetKey("user_id")

	pa := &dto.PaginatedUserOrdersRequest{}
	if err := ctx.Bind(pa); err != nil {
		handlerErr(ctx, errors.Wrap(errors.ErrBadRequest, err.Error()))
		return
	}

	if err := pa.Validate(userId.(int64)); err != nil {
		handlerErr(ctx, err)
		return
	}

	pao := pa.Map()
	if err := s.app.GetPaginated(pao); err != nil {
		handlerErr(ctx, err)
		return
	}
	r := &dto.PaginatedUserOrdersResponse{}
	r.Map(pao, false)
	ctx.JSON(http.StatusOK, r)
}

func (s *Server) GetPaginatedForAdmin(ctx Context) {

	pa := &dto.PaginatedUserOrdersRequest{}
	if err := ctx.Bind(pa); err != nil {
		handlerErr(ctx, errors.Wrap(errors.ErrBadRequest, errors.NewMesssage("invalid request")))
		return
	}

	if err := pa.Validate(0); err != nil {
		handlerErr(ctx, errors.Wrap(errors.ErrBadRequest, errors.NewMesssage(err.Error())))
		return
	}

	pao := pa.Map()
	if err := s.app.GetPaginated(pao); err != nil {
		handlerErr(ctx, err)
		return
	}
	r := &dto.PaginatedUserOrdersResponse{}
	r.Map(pao, true)
	ctx.JSON(http.StatusOK, r)
}

func (s *Server) SetTxId(ctx Context) {
	userId, _ := ctx.GetKey("user_id")
	r := &dto.SetTxIdRequest{}
	if err := ctx.Bind(r); err != nil {
		handlerErr(ctx, errors.Wrap(errors.ErrBadRequest, errors.NewMesssage("invalid request")))
		return
	}

	if err := r.Validate(); err != nil {
		r.Msg = errors.ErrorMsg(err)
		ctx.JSON(http.StatusBadRequest, r)
		return
	}

	if err := s.app.SetTxId(userId.(int64), r.Seq, r.TxId); err != nil {
		r.Msg = errors.ErrorMsg(err)
		ctx.JSON(http.StatusBadRequest, r)
		return
	}

	r.Msg = "transaction id setted successfully"
	ctx.JSON(http.StatusOK, r)
}
