/*
Test pgbouncer pools
Rory Campbell-Lange : 21 March 2021
*/

package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v4"
	flags "github.com/jessevdk/go-flags"
)

func dbquery(id int, url string, sleep int, wg *sync.WaitGroup) {

	defer wg.Done()

	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		fmt.Printf("Unable to connect to database: %v\n", err)
		return
	}
	defer conn.Close(context.Background())

	t1 := time.Now()
	_, err = conn.Exec(context.Background(), "select pg_sleep($1)", sleep)
	if err != nil {
		fmt.Printf("Exec failed: %v\n", err)
		return
	}
	t2 := time.Now()

	fmt.Printf("runtime for %3d: %v\n", id, t2.Sub(t1))
}

// Options show flag options
type Options struct {
	User      string   `short:"u" long:"user"      description:"database user" required:"true"`
	Pass      string   `short:"p" long:"password"  description:"database pass" required:"true"`
	Databases []string `short:"d" long:"databases" description:"database/s for pool tests" required:"true"`
	Port      int      `short:"P" long:"port"      description:"server port" default:"6432"`
	Host      string   `short:"H" long:"host"      description:"server host" default:"127.0.0.1"`
	Wait      int      `short:"w" long:"wait"      description:"per-connection pg_sleep seconds" default:"10"`
	Sleep     int      `short:"s" long:"sleep"     description:"milliseconds between launching next query" default:"800"`
	Conns     int      `short:"c" long:"conns"     description:"number of database connections per database" default:"10"`
}

// dburl constructs a database connection url
func (o *Options) dburl(database string) string {
	var tpl = "postgres://%s:%s@%s:%v/%s"
	return fmt.Sprintf(tpl, o.User, o.Pass, o.Host, o.Port, database)
}

var usage = `

A simple programme to test pgbouncer or postgresql connections.

Specify -d several times to connect to more than one database.`

func main() {

	var options Options
	var parser = flags.NewParser(&options, flags.Default)
	parser.Usage = usage

	// parse flags
	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}

	if options.User == "" || options.Pass == "" {
		fmt.Println("Both database user and password must be supplied")
		os.Exit(1)
	}

	if len(options.Databases) < 1 {
		fmt.Println("More than one database needs to be supplied")
		os.Exit(1)
	}

	for _, d := range options.Databases {
		if d == "" {
			fmt.Println("Empty databases strings are not supported")
			os.Exit(1)
		}
	}

	if net.ParseIP(options.Host) == nil {
		fmt.Println("Invalid IP address for host")
		os.Exit(1)
	}

	var wg sync.WaitGroup
	start := time.Now()
	sleepDuration := time.Duration(options.Sleep) * time.Millisecond
	counter := 0
	for i := 1; i <= options.Conns; i++ {
		for _, d := range options.Databases {
			counter++
			time.Sleep(sleepDuration)
			db := options.dburl(d)
			fmt.Printf("%3d : %s\n", counter, d)
			wg.Add(1)
			go dbquery(counter, db, options.Wait, &wg)
		}
	}
	wg.Wait()
	fmt.Printf("\nTotal runtime %v\n", time.Now().Sub(start))
}
