package pgxmetricrepository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/fdanis/ygtrack/internal/server/store/dataclass"
	"github.com/fdanis/ygtrack/internal/server/store/repository"
)

type pgxGougeRepository struct {
	db *sql.DB
}

func NewGougeRepository(d *sql.DB) repository.MetricRepository[float64] {
	return pgxGougeRepository{db: d}
}

func (r pgxGougeRepository) GetAll() ([]dataclass.Metric[float64], error) {
	res := make([]dataclass.Metric[float64], 0, 40)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	row, err := r.db.QueryContext(ctx, `	
	with last as 
		(select 
			id, 
			max(created) created 
		 FROM public.gaugemetric 
		 group by id)
	select 
		g.id,
		g.val 
	from public.gaugemetric g 
	where exists (select 1 from last where last.id = g.id and last.created = g.created limit 1);`)
	if err != nil {
		return nil, err
	}
	defer row.Close()
	for row.Next() {
		m := dataclass.Metric[float64]{}
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

func (r pgxGougeRepository) GetByName(name string) (*dataclass.Metric[float64], error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	row := r.db.QueryRowContext(ctx, "select id, val FROM public.gaugemetric where id = $1 order by created desc limit 1", name)

	m := dataclass.Metric[float64]{}
	err := row.Scan(&m.Name, &m.Value)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r pgxGougeRepository) Add(data dataclass.Metric[float64]) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s, err := r.db.ExecContext(ctx, "insert into public.gaugemetric (id,val) values ($1,$2)", data.Name, data.Value)
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
