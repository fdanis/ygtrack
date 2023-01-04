package driver

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const maxOpenConn = 15
const maxIdleConn = 15
const maxDbLifeTime = 5 * time.Minute

func ConnectSQL(dsn string) (*sql.DB, error) {
	d, err := NewDataBase(dsn)
	if err != nil {
		panic(err)
	}

	d.SetMaxOpenConns(maxOpenConn)
	d.SetMaxIdleConns(maxIdleConn)
	d.SetConnMaxLifetime(maxDbLifeTime)

	err = TestDB(d)
	if err != nil {
		return nil, err
	}

	err = CreateTabels(d)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func CreateTabels(d *sql.DB) error {
	_, err := d.Exec(`
	CREATE TABLE IF NOT EXISTS public.countmetric (id varchar(100),val bigint,created timestamp default now(),CONSTRAINT counter_id_time PRIMARY KEY(id,created));
	                                  
	CREATE TABLE IF NOT EXISTS public.gaugemetric (id varchar(100), val numeric(100,32), created timestamp default now(), CONSTRAINT id_time PRIMARY KEY(id,created));
	`)
	if err != nil {
		return err
	}

	return nil
}

func TestDB(d *sql.DB) error {
	err := d.Ping()
	if err != nil {
		return err
	}
	return nil
}

func NewDataBase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
