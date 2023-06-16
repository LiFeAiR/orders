package orders

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
	"pmozhchil/orders/api/pmozhchil/orders"
	"pmozhchil/orders/internal/app/repository"
)

type App interface {
	List(context.Context, *orders.ListRequest) (*orders.ListResponse, error)
	Create(context.Context, *orders.CreateRequest) (*emptypb.Empty, error)
}

type appImpl struct {
	db   *gorm.DB
	repo repository.Repository

	serviceVersion string
}

func NewApp(version string, db *gorm.DB) (App, error) {
	return &appImpl{
		db:   db,
		repo: repository.New(db),

		serviceVersion: version,
	}, nil
}

func (a *appImpl) List(ctx context.Context, request *orders.ListRequest) (*orders.ListResponse, error) {
	sql, err := a.db.DB()
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

func (a *appImpl) Create(ctx context.Context, request *orders.CreateRequest) (*emptypb.Empty, error) {
	//err := a.repo.WithCte(ctx, request.ClientId)
	err := a.repo.WithSelect(ctx, request.ClientId)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
