// A Go programme for running queries concurrently on a set of Postgresql
// databases.
//
// Rory Campbell-Lange
// July 2022

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
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
		dbqs := []DBQuery{}

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
			dbqs = append(dbqs, dbq)
		}

		// make the query group and add it to the slice
		dbqg := NewDBQueryGroup(
			dbGroupName,
			dbGroup.Concurrency,
			options.DontCycle,
		)
		// cannot send slice of interface; add one by one
		for _, d := range dbqs {
			dbqg.AddQuerier(d)
		}
		queryGroups = append(queryGroups, dbqg)
	}

	done := make(chan struct{})
	// process
	for _, d := range queryGroups {
		go d.Process()
	}

	<-done

}
