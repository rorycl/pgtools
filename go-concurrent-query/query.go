package main

import (
	"context"
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

// checkConnection checks if the required database can be access
func (d *DBQuery) checkConnection() error {
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

// dbquery queries a database, reporting erros on errorchan, cancelling
// the function if anything is received on the done channel
func (d *DBQuery) query(done <-chan struct{}, errorchan chan<- error) {

	go func() {
		for {
			select {
			case <-done:
				return
			}
		}
	}()

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
