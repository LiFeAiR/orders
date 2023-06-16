package orders

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
	"pmozhchil/orders/api/pmozhchil/orders"
	gw "pmozhchil/orders/api/pmozhchil/orders"
	"pmozhchil/orders/internal/app/repository"
)

var _ gw.OrdersServiceServer = (*ordersServiceImpl)(nil)

type ordersServiceImpl struct {
	db   *gorm.DB
	repo repository.Repository

	serviceVersion string
}

func NewOrdersService(version string, db *gorm.DB) (gw.OrdersServiceServer, error) {
	return &ordersServiceImpl{
		db:   db,
		repo: repository.New(db),

		serviceVersion: version,
	}, nil
}

func (o *ordersServiceImpl) List(ctx context.Context, request *orders.ListRequest) (*orders.ListResponse, error) {
	sql, err := o.db.DB()
	if err != nil {
		return nil, err
	}

	rows, err := sql.QueryContext(ctx, "select "+
		"id, "+
		"client_id, "+
		"number, "+
		"order_number "+
		"from orders "+
		"where client_id = $1 "+
		"order by id "+
		"limit 10", request.ClientId)
	if err != nil {
		return nil, err
	}

	items := make([]*orders.ListResponse_Order, 0)
	for rows.Next() {
		var item = &orders.ListResponse_Order{}
		if err := rows.Scan(&item.Id, &item.ClientId, &item.Number, &item.OrderNumber); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return &orders.ListResponse{Orders: items}, nil
}

func (o *ordersServiceImpl) Create(ctx context.Context, request *orders.CreateRequest) (*emptypb.Empty, error) {
	err := o.repo.WithCte(ctx, request.ClientId)
	//err := o.repo.WithSelect(ctx, request.ClientId)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
