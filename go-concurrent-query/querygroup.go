package main

import (
	"context"
	"fmt"
)

// Querier is an interface for DBQuery.Query, to allow for testing
type Querier interface {
	Query(ctx context.Context, label string, errorChan chan<- error, resultChan chan<- string)
}

// DBQueryGroup represents all the information needed for a query group
type DBQueryGroup struct {
	Name        string
	Concurrency int
	DBQueries   []Querier
	errorChan   chan error    // queryChan errors
	resultChan  chan string   // queryChan results
	done        chan struct{} // signal the querygroup queries as complete
	dontCycle   bool
}

// NewDBQueryGroup returns a new DBQueryGroup
func NewDBQueryGroup(name string, concurrency int, dontCycle bool) *DBQueryGroup {
	dbqg := DBQueryGroup{
		Name:        name,
		Concurrency: concurrency,
		dontCycle:   dontCycle,
	}
	dbqg.errorChan = make(chan error)
	dbqg.resultChan = make(chan string)
	dbqg.done = make(chan struct{})
	return &dbqg
}

// AddQuerier adds a query
func (dbqg *DBQueryGroup) AddQuerier(q Querier) {
	dbqg.DBQueries = append(dbqg.DBQueries, q)
}

// Process the queries in the group, controlled by a context and
// printing goroutine errors on errorChan
func (dbqg *DBQueryGroup) Process(ctx context.Context) {

	if len(dbqg.DBQueries) < 1 {
		dbqg.errorChan <- fmt.Errorf("no queries to run in querygroup %s", dbqg.Name)
		dbqg.done <- struct{}{}
		return
	}

	// producer: if dontCycle is true simply iterate over the dbqueries and
	// push them onto the query channel (to be processed by the block above),
	// otherwise continously push database queries onto the query channel
	runQueries := func() <-chan Querier {
		rq := make(chan Querier)
		go func() {
			if dbqg.dontCycle {
				for _, q := range dbqg.DBQueries {
					rq <- q
				}
				close(rq)
				// dbqg.done <- struct{}{}
			} else {
				for counter := 0; ; counter++ {
					i := counter % len(dbqg.DBQueries)
					rq <- dbqg.DBQueries[i]
				}
			}
		}()
		return rq
	}

	// consumer: launch consumer goroutines for processing queries
	for i := 1; i <= dbqg.Concurrency; i++ {
		go func() {
			for d := range runQueries() {
				d.Query(ctx, dbqg.Name, dbqg.errorChan, dbqg.resultChan)
			}
			dbqg.done <- struct{}{}
			return
		}()
	}

	// close(dbqg.errorChan)
	// close(dbqg.resultChan)
	// close(dbqg.done)
	return
}
