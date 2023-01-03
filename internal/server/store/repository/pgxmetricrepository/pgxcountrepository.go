package pgxmetricrepository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/fdanis/ygtrack/internal/server/store/dataclass"
	"github.com/fdanis/ygtrack/internal/server/store/repository"
)

type pgxCountRepository struct {
	db *sql.DB
}

func NewCountRepository(d *sql.DB) repository.MetricRepository[int64] {
	return pgxCountRepository{db: d}
}

func (r pgxCountRepository) GetAll(ctx context.Context) ([]dataclass.Metric[int64], error) {
	res := make([]dataclass.Metric[int64], 0, 1)

	m, err := r.GetByName(ctx, "")
	if err != nil {
		return nil, err
	}

	res = append(res, *m)
	return res, nil
}

func (r pgxCountRepository) GetByName(ctx context.Context, name string) (*dataclass.Metric[int64], error) {
	row := r.db.QueryRowContext(ctx, "SELECT val FROM public.countmetric order by created desc limit 1")
	var val sql.NullInt64
	err := row.Scan(&val)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	if !val.Valid {
		return nil, nil
	}
	return &dataclass.Metric[int64]{Name: "count", Value: val.Int64}, nil
}

func (r pgxCountRepository) Add(ctx context.Context, data dataclass.Metric[int64]) error {
	s, err := r.db.ExecContext(ctx, "insert into public.countmetric (val) values ($1)", data.Value)
	if err != nil {
		return err
	}
	i, err := s.RowsAffected()
	if err != nil {
		return err
	}
	if i == 0 {
		return errors.New("metric was not save")
	}
	return nil
}
