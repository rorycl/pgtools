# pgtools/go-concurrent-query

A Go programme for running queries concurrently on a set of Postgresql
databases.

Rory Campbell-Lange  
July 2022

## Introduction

This programme was written to perform load testing on a set of databses
in a cluster.

## Usage

Programme options

    Usage:
      concurrent-query

    A programme to run concurrent queries against Postgresql databaes.

    Application Options:
      -u, --user=      database user
      -p, --password=  database pass
      -c, --config=    configuration yaml file (default: config.yaml)
      -P, --port=      server port (default: 5432)
      -H, --host=      server host (default: 127.0.0.1)
      -d, --duration=  maximum duration of tests (default: indefinite)
      -e, --errexit=      exit on first error (default: false)

    Help Options:
      -h, --help       Show this help message

Yaml configuration

    label:
        # list of databases for test
        databases: [list, of, databases]
        # number of databases on which to run concurrent queries
        # needs to be <= len(databases)
        concurrency: 3
        # how many times to run the queries on each database until
        # moving onto the next database
        iterations: 3
        queries:
            - >
                select * from function1()
            - >
                select 1
            - >
                select pg_sleep(5)

It is possible to configure more than one set of tests to run
concurrently, each with their own label.

