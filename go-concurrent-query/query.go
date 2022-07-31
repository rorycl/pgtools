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

// Query queries a database, reporting errors on errorChan
func (d DBQuery) Query(label string, errorChan chan<- error) {
	if d.DBURL == "" {
		errorChan <- fmt.Errorf("db url for %s is empty", d.DBName)
		return
	}
	conn, err := pgx.Connect(context.Background(), d.DBURL)
	defer conn.Close(context.Background())
	if err != nil {
		log.Fatalf("error connecting to %s : %s", d.DBName, err)
	}
	for i := 1; i <= d.Iterations; i++ {
		for _, q := range d.Queries {
			t1 := time.Now()
			_, err = conn.Exec(context.Background(), q)
			if err != nil {
				errorChan <- fmt.Errorf(
					"error on %s executing %s: %s", d.DBName, q, err,
				)
				continue
			}
			t2 := time.Now()
			log.Printf("[%-20s:%02d] %0.3fs %s\n",
				label+":"+d.DBName, i, float64(t2.Sub(t1))/float64(time.Second), q,
			)
		}
	}
	return
}
