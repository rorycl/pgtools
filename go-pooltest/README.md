# pgtools/go-pooltest

A Go tool for testing pgbouncer connection pools or postgresql
connections.

Rory Campbell-Lange  
February 2021

## Introduction

This small programme was written to test pgbouncer connections to get a
clearer idea of how reserve pools work and how these behave in relation
to other pgbouncer pool size limits. It can also be used for direct
postgresql connection testing.

## Usage

	Usage:
	  pooltest 

	A simple programme to test pgbouncer or postgresql connections.

	Specify -d several times to connect to more than one database.

	Application Options:
	  -u, --user=      database user
	  -p, --password=  database pass
	  -d, --databases= database/s for pool tests
	  -P, --port=      server port (default: 6432)
	  -H, --host=      server host (default: 127.0.0.1)
	  -w, --wait=      per-connection pg_sleep seconds (default: 10)
	  -s, --sleep=     milliseconds between launching next query (default: 800)
	  -c, --conns=     number of database connections per database (default: 10)

	Help Options:
	  -h, --help       Show this help message

## Example

A simple example testing only one pool with 12 more or less concurrent
connections running a postgres query lasting 5 seconds each, with a 2
microsecond sleep between launching connections.

	./pooltest -u user -p password -d pooltest1 -w 5 -s 2 -c 12
	1 : pooltest1
	2 : pooltest1
	3 : pooltest1
	4 : pooltest1
	5 : pooltest1
	6 : pooltest1
	7 : pooltest1
	8 : pooltest1
	9 : pooltest1
	10 : pooltest1
	11 : pooltest1
	12 : pooltest1
	runtime for 6: 5.014423393s
	runtime for 4: 5.015612711s
	runtime for 5: 7.083233373s
	runtime for 3: 7.089432405s
	runtime for 2: 10.01496856s
	runtime for 1: 10.017172524s
	runtime for 7: 12.08788673s
	runtime for 8: 12.090459103s
	runtime for 9: 15.012166809s
	runtime for 11: 15.004735847s
	runtime for 10: 17.072524473s
	runtime for 12: 17.081095591s

Connections 5 and 3 in this example hit a pgbouncer reserve pool with a
pool size of 2 with a 2 second delay.


