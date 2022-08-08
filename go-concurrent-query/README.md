# pgtools/go-concurrent-query

A Go programme for running queries concurrently on a set of Postgresql
databases.

Rory Campbell-Lange  
August 2022

## Introduction

This programme was written to perform load testing on a set of databses
in a cluster.

## Usage

Programme usage

    Usage:
      concurrent-query 

    Run queries concurrently on a set of Postgresql databases.

    Application Options:
      -u, --user=      database user
      -p, --password=  database pass
      -c, --config=    database query group yaml file
      -P, --port=      server port (default: 5432)
      -H, --host=      server host (default: 127.0.0.1)
      -d, --duration=  limit test duration in seconds (default: 0)
          --dontcycle  don't cycle databases, process each only once
      -e, --errexit    exit on first query err

    Help Options:
      -h, --help       Show this help message

Yaml configuration

```yaml
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
```

It is possible to configure more than one set of tests to run
concurrently, each with their own label.
