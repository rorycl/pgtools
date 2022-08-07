// +build dbtest

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"
)

var (
	host, port, user, pass, db string
	ok                         bool
)

func setup() error {
	host, ok = os.LookupEnv("PGHOST")
	if !ok {
		return errors.New("PGHOST not set")
	}
	port, ok = os.LookupEnv("PGPORT")
	if !ok {
		return errors.New("PGPORT not set")
	}
	user, ok = os.LookupEnv("PGUSER")
	if !ok {
		return errors.New("PGUSER not set")
	}
	pass, ok = os.LookupEnv("PGPASS")
	if !ok {
		return errors.New("PGPASS not set")
	}
	db, ok = os.LookupEnv("PGDATABASE")
	if !ok {
		return errors.New("PGDATABASE not set")
	}
	return nil
}

// TestQuery tests DBQuery.Query
func TestDBQuery(t *testing.T) {

	if err := setup(); err != nil {
		t.Fatal(err)
	}

	dbq := DBQuery{
		DBName:     db, // a label
		DBURL:      fmt.Sprintf("postgres://%s:%s@%s:%v/%s", user, pass, host, port, db),
		Iterations: 2,
		Queries: []string{
			"select 1",
			"select * from pg_sleep(0.1)",
			"select * from x",
		},
	}

	err := dbq.checkConnection()
	if err != nil {
		t.Errorf("connection failed: %s", err)
	}

	errChan := make(chan error)
	resultChan := make(chan string)
	ctx, cancel := context.WithDeadline(
		context.Background(),
		time.Now().Add(1*time.Second),
	)

	done := make(chan struct{})
	go func() {
		dbq.Query(ctx, "test", errChan, resultChan)
		done <- struct{}{}
	}()

	errorCount := 0
LOOP:
	for {
		select {
		case <-done:
			break LOOP
		case e := <-errChan:
			errorCount++
			t.Logf("error %s\n", e)
		case r := <-resultChan:
			t.Logf("result %s\n", r)
		case <-ctx.Done():
			cancel()
			t.Errorf("deadline timed out")
			break LOOP
		}
	}

	if errorCount != dbq.Iterations {
		t.Errorf("error count should be %d, is %d", dbq.Iterations, errorCount)
	}
}

// TestDBQueryCancel tests DBQuery.Query
func TestDBQueryCancel(t *testing.T) {

	if err := setup(); err != nil {
		t.Fatal(err)
	}

	dbq := DBQuery{
		DBName:     db, // a label
		DBURL:      fmt.Sprintf("postgres://%s:%s@%s:%v/%s", user, pass, host, port, db),
		Iterations: 2,
		Queries: []string{
			"select 1",
			"select * from pg_sleep(0.1)",
		},
	}

	err := dbq.checkConnection()
	if err != nil {
		t.Errorf("connection failed: %s", err)
	}

	errChan := make(chan error)
	resultChan := make(chan string)
	ctx, cancel := context.WithDeadline(
		context.Background(),
		time.Now().Add(100*time.Millisecond),
	)

	done := make(chan struct{})
	go func() {
		dbq.Query(ctx, "test", errChan, resultChan)
		done <- struct{}{}
	}()

LOOP:
	for {
		select {
		case <-done:
			t.Errorf("ctx.Done expected, not done")
			break LOOP
		case <-errChan:
			t.Log("ctx.Done expected, not error")
		case r := <-resultChan:
			t.Logf("result %s\n", r)
		case <-ctx.Done():
			cancel()
			break LOOP
		}
	}
}
