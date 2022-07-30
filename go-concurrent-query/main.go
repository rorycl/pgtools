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

	// setup dbqueries
	dbQueries := []DBQuery{}
	for _, dbGroup := range config {
		for _, db := range dbGroup.Databases {
			dbq := DBQuery{
				DBName:     db,
				DBURL:      "something",
				Iterations: dbGroup.Iterations,
				Queries:    dbGroup.Queries,
			}
			dbQueries = append(dbQueries, dbq)
		}
	}

	// start processing
}
