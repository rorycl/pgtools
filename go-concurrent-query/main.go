// A Go programme for running queries concurrently on a set of Postgresql
// databases.
//
// Rory Campbell-Lange
// August 2022
// MIT Licence

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {

	// retrieve options
	options, err := ParseOpts()
	if err != nil {
		os.Exit(1)
	}

	// retrieve yaml configuration
	filer, err := ioutil.ReadFile(options.Config)
	if err != nil {
		fmt.Printf("could not load file: %s", err)
		os.Exit(1)
	}
	config, err := LoadYaml(filer)
	if err != nil {
		fmt.Printf("yaml file error: %s", err)
		os.Exit(1)
	}

	// setup dbquerygroups
	queryGroups := []*DBQueryGroup{}

	for dbGroupName, dbGroup := range config {

		// make a query group
		dbqg := NewDBQueryGroup(
			dbGroupName,
			dbGroup.Concurrency,
			options.DontCycle,
		)

		// setup each database
		for _, db := range dbGroup.Databases {
			dbq := DBQuery{
				DBName:     db,
				Iterations: dbGroup.Iterations,
				Queries:    dbGroup.Queries,
			}
			// make connection url
			dbq.setDBURL(
				options.User, options.Pass, options.Host, strconv.Itoa(options.Port), db,
			)
			// cannot send slice of interface; add one by one
			dbqg.AddQuerier(dbq)
		}

		queryGroups = append(queryGroups, dbqg)
	}

	// create context to allow closing of all goroutines
	ctx, cancel := context.WithCancel(context.Background())
	if options.Duration > 0 {
		ctx, cancel = context.WithDeadline(
			ctx,
			time.Now().Add(time.Duration(options.Duration)*time.Second),
		)
	}
	defer cancel()

	// process each queryGroup, using a context to allow cancellation of
	// associated goroutines and database queries
	doneCount := 0
	t1 := time.Now()
	for _, qg := range queryGroups {
		go func(qgHere *DBQueryGroup) {
			go qgHere.Process(ctx)

			for {
				select {
				case e := <-qgHere.errorChan:
					log.Println(e)
					if options.ErrExit {
						log.Println("exiting on first error")
						cancel()
					}
				case r := <-qgHere.resultChan:
					log.Println(r)
				case <-qgHere.done:
					log.Printf("query group %s done", qgHere.Name)
					doneCount++
					if doneCount == len(queryGroups) {
						cancel()
					}
					// ctx.Done() is caught by the general querygroup
					// select below
				}
			}

		}(qg)
	}

LOOP:
	for {
		select {
		// catch context cancellations due to cancel or timeout events
		// across querygroups
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				log.Println(err)
			}
			cancel()
			break LOOP
		}
	}

	// finish up
	t2 := time.Now()
	log.Printf("Completed in %s\n", t2.Sub(t1))
}
