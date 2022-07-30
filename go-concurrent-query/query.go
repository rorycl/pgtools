package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v4"
)

// DBQuery details that are needed to make queries against a db
type DBQuery struct {
	DBName     string
	DBURL      string
	Iterations int
	Queries    []string
}

// setDBURL constructs a database connection url
func (d *DBQuery) setDBURL(user, pass, host, port, database string) {
	var tpl = "postgres://%s:%s@%s:%v/%s"
	d.DBURL = fmt.Sprintf(tpl, user, pass, host, port, database)
}

// checkConnection checks if the required database can be access
func (d *DBQuery) checkConnection() error {
	if d.DBURL == "" {
		return errors.New("the database url is empty")
	}
	conn, err := pgx.Connect(context.Background(), d.DBURL)
	if err != nil {
		return err
	}
	_, err = conn.Exec(context.Background(), "select 1")
	if err != nil {
		return err
	}
	return nil
}

// Query queries a database, reporting errors on errorchan
func (d *DBQuery) Query(errorchan chan<- error) {
	if d.DBURL == "" {
		errorchan <- fmt.Errorf("db url for %s is empty", d.DBName)
		return
	}
	conn, err := pgx.Connect(context.Background(), d.DBURL)
	defer conn.Close(context.Background())
	if err != nil {
		errorchan <- fmt.Errorf(
			"error connecting to %s : %w",
			d.DBName, err,
		)
		return
	}
	for i := 1; i <= d.Iterations; i++ {
		log.Printf("db %s starting iteration %d\n", d.DBName, i)
		for _, q := range d.Queries {
			t1 := time.Now()
			_, err = conn.Exec(context.Background(), q)
			if err != nil {
				errorchan <- fmt.Errorf(
					"error connecting to %s : %w",
					d.DBName, err,
				)
				continue
			}
			t2 := time.Now()
			log.Printf("db %s runtime %3d for query %s\n",
				d.DBName, t2.Sub(t1), q,
			)
		}
	}
	return
}
