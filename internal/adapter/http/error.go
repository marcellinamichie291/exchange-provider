package http

import (
	"net/http"
	"order_service/pkg/errors"
)

func handlerErr(ctx Context, err error) {
	switch errors.ErrorCode(err) {
	case errors.ErrNotFound:
		ctx.JSON(http.StatusNotFound, errors.ErrorMsg(err))
		return
	case errors.ErrBadRequest:
		ctx.JSON(http.StatusBadRequest, errors.ErrorMsg(err))
	default:
		ctx.JSON(http.StatusInternalServerError, errors.ErrorMsg(err))
		return
	}
}
