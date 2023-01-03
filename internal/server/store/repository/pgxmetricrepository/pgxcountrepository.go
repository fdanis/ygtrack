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

	row, err := r.db.QueryContext(ctx, `	
	with last as 
		(select 
			id, 
			max(created) created 
		 FROM public.countmetric 
		 group by id)
	select 
		g.id,
		g.val 
	from public.countmetric g 
	where exists (select 1 from last where last.id = g.id and last.created = g.created limit 1);`)
	if err != nil {
		return nil, err
	}
	defer row.Close()
	for row.Next() {
		m := dataclass.Metric[int64]{}
		err = row.Scan(&m.Name, &m.Value)
		if err != nil {
			return nil, err
		}
		res = append(res, m)
	}
	err = row.Err()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r pgxCountRepository) GetByName(ctx context.Context, name string) (*dataclass.Metric[int64], error) {
	row := r.db.QueryRowContext(ctx, "SELECT id, val FROM public.countmetric where id=$1 order by created desc limit 1", name)
	var val int64
	err := row.Scan(&val)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &dataclass.Metric[int64]{Name: "count", Value: val}, nil
}

func (r pgxCountRepository) Add(ctx context.Context, data dataclass.Metric[int64]) error {
	s, err := r.db.ExecContext(ctx, "insert into public.countmetric (val,id) values ($1,$2)", data.Value, data.Name)
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
