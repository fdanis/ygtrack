package pgxmetricrepository

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/fdanis/ygtrack/internal/server/store/dataclass"
	"github.com/fdanis/ygtrack/internal/server/store/repository"
)

type pgxCountRepository struct {
	db         *sql.DB
	insertStmt *sql.Stmt
}

func NewCountRepository(d *sql.DB) repository.MetricRepository[int64] {
	insertStmt, err := d.PrepareContext(context.TODO(), "insert into public.countmetric (id,val) values ($1,$2)")
	if err != nil {
		log.Fatal("can not create statement for pgxCountRepository")
	}

	return pgxCountRepository{db: d, insertStmt: insertStmt}
}

func (r pgxCountRepository) GetAll() ([]dataclass.Metric[int64], error) {
	res := make([]dataclass.Metric[int64], 0, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	row, err := r.db.QueryContext(ctx, `	
	with last as 
		(select 
			id, 
			max(k) k 
		 FROM public.countmetric 
		 group by id)
	select 
		g.id,
		g.val 
	from public.countmetric g 
	where exists (select 1 from last where last.id = g.id and last.k = g.k limit 1);`)
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

func (r pgxCountRepository) GetByName(name string) (*dataclass.Metric[int64], error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	row := r.db.QueryRowContext(ctx, "SELECT id, val FROM public.countmetric where id=$1 order by created desc limit 1", name)
	var val int64
	var n string
	err := row.Scan(&n, &val)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &dataclass.Metric[int64]{Name: n, Value: val}, nil
}

func (r pgxCountRepository) Add(data dataclass.Metric[int64]) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s, err := r.insertStmt.ExecContext(ctx, data.Name, data.Value)
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

func (r pgxCountRepository) AddBatch(tx *sql.Tx, data []dataclass.Metric[int64]) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	txStmt := tx.StmtContext(ctx, r.insertStmt)
	for _, v := range data {
		if _, err := txStmt.ExecContext(ctx, v.Name, v.Value); err != nil {
			return err
		}
	}
	return nil
}
