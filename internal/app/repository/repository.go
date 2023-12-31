package repository

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

var _ Repository = (*repoImpl)(nil)

type Repository interface {
	WithSelect(ctx context.Context, clientId int64) error
	WithCte(ctx context.Context, clientId int64) error
	WithLock(ctx context.Context, clientId int64) error
	List(ctx context.Context, clientId int64) (*sql.Rows, error)
}

// TODO using gorm style
type repoImpl struct {
	db *gorm.DB
}

func (r *repoImpl) List(ctx context.Context, clientId int64) (*sql.Rows, error) {
	q, err := r.db.DB()
	if err != nil {
		return nil, err
	}

	return q.QueryContext(ctx, "select "+
		"id, "+
		"client_id, "+
		"number, "+
		"order_number "+
		"from orders "+
		"where client_id = $1 "+
		"order by id "+
		"limit 10", clientId)
}

func (r *repoImpl) WithSelect(ctx context.Context, clientId int64) error {
	q, err := r.db.DB()
	if err != nil {
		return err
	}
	var nextNumber int64

	err = q.QueryRowContext(ctx,
		`select number from orders where client_id = $1 order by id desc limit 1`,
		clientId,
	).Scan(&nextNumber)

	switch {
	case err == sql.ErrNoRows:
		//skip
	case err != nil:
		return err
	}

	nextNumber = nextNumber + 1
	_, err = q.Exec(
		`insert into orders(client_id, number, order_number)values($1, $2, concat($1, '-', $2))`,
		clientId, nextNumber)
	if err != nil {
		return err
	}

	return nil
}

func (r *repoImpl) WithLock(ctx context.Context, clientId int64) (err error) {
	q, err := r.db.DB()
	if err != nil {
		return err
	}
	var nextNumber int64
	tx, err := q.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				return
			}
		}
	}()

	_, err = tx.ExecContext(ctx,
		`SELECT pg_advisory_xact_lock($1)`,
		clientId,
	)
	if err != nil {
		return err
	}

	err = tx.QueryRowContext(ctx,
		`select number from orders where client_id = $1 order by id desc limit 1`,
		clientId,
	).Scan(&nextNumber)

	switch {
	case err == sql.ErrNoRows:
		//skip
	case err != nil:
		return err
	}

	nextNumber = nextNumber + 1
	_, err = tx.Exec(
		`insert into orders(client_id, number, order_number)values($1, $2, concat($1, '-', $2))`,
		clientId, nextNumber)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (r *repoImpl) WithCte(ctx context.Context, clientId int64) error {
	q, err := r.db.DB()
	if err != nil {
		return err
	}

	_, err = q.ExecContext(ctx, `with x as (update orders_id
		set orders_count = orders_count + 1 where client_id = $1
		returning *
		)
		insert into orders(number, client_id, order_number)
		select x.orders_count, $1, concat($1, '-', x.orders_count) from x;`, clientId)
	if err != nil {
		return err
	}

	return nil
}

func New(db *gorm.DB) Repository {
	return &repoImpl{
		db: db,
	}
}
