// +build dbtest

package main

import (
	"fmt"
	"os"
	"testing"
	"time"
)

// TestQuery tests DBQuery.Query
func TestQuery(t *testing.T) {

	host, ok := os.LookupEnv("PGHOST")
	if !ok {
		t.Fatal("PGHOST not set")
	}
	port, ok := os.LookupEnv("PGPORT")
	if !ok {
		t.Fatal("PGPORT not set")
	}
	user, ok := os.LookupEnv("PGUSER")
	if !ok {
		t.Fatal("PGUSER not set")
	}
	pass, ok := os.LookupEnv("PGPASS")
	if !ok {
		t.Fatal("PGPASS not set")
	}
	db, ok := os.LookupEnv("PGDATABASE")
	if !ok {
		t.Fatal("PGDATABASE not set")
	}

	dbq := DBQuery{
		DBName:     db, // a label
		DBURL:      fmt.Sprintf("postgres://%s:%s@%s:%v/%s", user, pass, host, port, db),
		Iterations: 2,
		Queries: []string{
			"select 1",
			"select * from pg_sleep(1)",
			"select * from x",
		},
	}

	err := dbq.checkConnection()
	if err != nil {
		t.Errorf("connection failed: %s", err)
	}

	errChan := make(chan error)
	done := make(chan struct{})

	go dbq.Query(done, errChan)

	after := time.After(time.Second * 3)

	errorCount := 0
	for {
		select {
		case e := <-errChan:
			fmt.Println("--")
			errorCount++
			t.Logf("error %s\n", e)
		case <-after:
			fmt.Println("++")
			t.Logf("planned timeout")
			done <- struct{}{}
			break
		}
	}

	if errorCount != dbq.Iterations {
		t.Errorf("error count should be %d, is %d", dbq.Iterations, errorCount)
	}
}
