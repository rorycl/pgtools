package main

import (
	"fmt"
	"log"
)

// Querier is an interface for DBQuery.Query, to allow for testing
type Querier interface {
	Query(label string, errorChan chan<- error)
}

// DBQueryGroup represents all the information needed for a query group
type DBQueryGroup struct {
	Name        string
	Concurrency int
	DBQueries   []Querier
	errorChan   chan error
	queryChan   chan Querier
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
	dbqg.queryChan = make(chan Querier)
	return &dbqg
}

// AddQuerier adds a query
func (dbqg *DBQueryGroup) AddQuerier(q Querier) {
	dbqg.DBQueries = append(dbqg.DBQueries, q)
}

// Process the queries in the group, printing goroutine errors on errorChan
func (dbqg *DBQueryGroup) Process(done chan<- struct{}) {

	if len(dbqg.DBQueries) < 1 {
		dbqg.errorChan <- fmt.Errorf("no queries to run in querygroup %s", dbqg.Name)
		return
	}

	go func() {
		for e := range dbqg.errorChan {
			log.Print(e)
		}
	}()

	for i := 1; i <= dbqg.Concurrency; i++ {
		// launch consumer goroutines for processing queries
		go func(errorChan chan<- error, queryChan <-chan Querier) {
			for d := range queryChan {
				d.Query(dbqg.Name, errorChan)
			}
		}(dbqg.errorChan, dbqg.queryChan)
	}

	// either iterate over the dbqueries or continously produce database
	// queries
	if dbqg.dontCycle {
		for _, q := range dbqg.DBQueries {
			dbqg.queryChan <- q
		}
		log.Println("Done")
		done <- struct{}{}
	} else {
		for counter := 0; ; counter++ {
			i := counter % dbqg.Concurrency
			dbqg.queryChan <- dbqg.DBQueries[i]
		}
	}
}
