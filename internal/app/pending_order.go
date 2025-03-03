package app

import (
	"exchange-provider/internal/entity"
)

func (o *OrderUseCase) totalPendingOrders(ex entity.Exchange, fs ...*entity.Filter) (total int64, err error) {
	f1 := &entity.Filter{
		Param:    "exchange",
		Operator: entity.FilterOperatorEqual,
		Values:   []interface{}{ex.NID()},
	}

	f2 := &entity.Filter{
		Param:    "status",
		Operator: entity.FilterOperatorNotIn,
		Values:   []interface{}{"succeed", "failed"},
	}

	pa := &entity.PaginatedUserOrders{
		Page:    1,
		PerPage: 1,
		Total:   0,
		Filters: []*entity.Filter{f1, f2},
		Orders:  []*entity.UserOrder{},
	}

	pa.Filters = append(pa.Filters, fs...)

	if err = o.GetPaginated(pa); err != nil {
		return 0, err
	}

	return pa.Total, nil
}
