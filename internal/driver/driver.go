package driver

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	SQL *sql.DB
}

var dbConn = &DB{}

const maxOpenConn = 10
const maxIdleConn = 5
const maxDbLifeTime = 5 * time.Minute

func ConnectSQL(dsn string) (*DB, error) {
	d, err := NewDataBase(dsn)
	if err != nil {
		panic(err)
	}

	d.SetMaxOpenConns(maxOpenConn)
	d.SetMaxIdleConns(maxIdleConn)
	d.SetConnMaxLifetime(maxDbLifeTime)

	dbConn.SQL = d
	err = TestDB(d)
	if err != nil {
		return nil, err
	}

	err = CreateTabels(d)
	if err != nil {
		return nil, err
	}

	return dbConn, nil
}

func CreateTabels(d *sql.DB) error {
	_, err := d.Exec(`
	CREATE TABLE IF NOT EXISTS public.countmetric (val integer,created timestamp default now() PRIMARY KEY);
	                                  
	CREATE TABLE IF NOT EXISTS public.gaugemetric (name varchar(100), val numeric,created timestamp default now(), CONSTRAINT name_time PRIMARY KEY(name,created));
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
