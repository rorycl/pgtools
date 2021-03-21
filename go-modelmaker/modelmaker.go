/*
 MModelmaker

  Describe postgres function signatures, possibly for use in creating
  model interface files for other languages, eg. python

  Rory Campbell-Lange : 21 March 2021
*/

package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	flags "github.com/jessevdk/go-flags"
)

const query = `
SELECT
    n.nspname as "schema"
	,p.proname as "name"
	,pg_catalog.pg_get_function_arguments(p.oid) as "arguments"
	,pg_catalog.pg_get_function_result(p.oid) as "returns"
FROM
	pg_catalog.pg_proc p
	LEFT JOIN pg_catalog.pg_namespace n ON n.oid = p.pronamespace
	LEFT JOIN pg_catalog.pg_language l ON l.oid = p.prolang
WHERE
	p.proname OPERATOR(pg_catalog.~) 'XXX'
	AND p.prokind IN ('p', 'f')
	AND n.nspname NOT IN ('pg_catalog')
	AND pg_catalog.pg_function_is_visible(p.oid)
ORDER BY 1, 2, 4;
`

// Arg are component arguments to a db function
type Arg struct {
	Name    string
	Typer   string
	Default string
}

// NameNoIn removes 'in_' from any field name
func (a *Arg) NameNoIn() string {
	return strings.Replace(a.Name, "in_", "", 1)
}

// TypeDefaulted returns True if the type is defaulted
func (a *Arg) TypeDefaulted() bool {
	return a.Default != ""
}

// Result is a db function descriptor
type Result struct {
	Schema    string `db:"schema"`
	Function  string `db:"name"`
	Arguments string `db:"arguments"`
	Returns   string `db:"returns"`
	Args      []Arg
	tpl       string
}

// Arger splits result arguments into Arg types
func (r *Result) Arger() (err error) {
	a := strings.Split(r.Arguments, ",")
	for _, p := range a {
		s := strings.SplitN(strings.Trim(p, " "), " ", 2)
		t := strings.Split(s[1], " DEFAULT ")
		var d string
		if len(t) > 1 {
			d = t[1]
		}
		a := Arg{
			Name:    s[0],
			Typer:   t[0],
			Default: d,
		}
		r.Args = append(r.Args, a)
	}
	return err
}

// ResultsSingular tries to determine if a result set only returns one row
func (r *Result) ResultsSingular() bool {
	switch r.Returns {
	case "boolean":
		return true
	case "string":
		return true
	}
	if strings.Contains(r.Function, "manage") {
		return true
	}
	return false
}

// String returns a string representation of a result
func (r *Result) String() (returner string) {

	var b bytes.Buffer
	t := template.Must(template.ParseFiles(r.tpl))
	t.Execute(&b, r)
	return fmt.Sprint(b.String())
}

func dbquery(url string, searchpath string, searchstring string) []*Result {

	ctx := context.Background()
	db, err := pgx.Connect(ctx, url)
	if err != nil {
		log.Fatalf("Connection error : %s : %v\n", url, err)
	}
	_, err = db.Exec(ctx, "set search_path="+searchpath)
	if err != nil {
		log.Fatalf("Exec error: %v\n", err)
	}

	var results []*Result
	hquery := strings.Replace(query, "XXX", searchstring, 1)
	pgxscan.Select(ctx, db, &results, hquery)

	return results
}

// Options show flag options
type Options struct {
	User       string `short:"u" long:"user"       description:"database user" required:"true"`
	Pass       string `short:"p" long:"password"   description:"database pass" required:"true"`
	Database   string `short:"d" long:"database"   description:"database" required:"true"`
	Port       int    `short:"P" long:"port"       description:"server port" default:"5432"`
	Host       string `short:"H" long:"host"       description:"server host" default:"127.0.0.1"`
	Searchpath string `short:"s" long:"searchpath" description:"searchpath" required:"true"`
	Template   string `short:"t" long:"template"   description:"template file" default:"output.tpl"`
	Filter     string `short:"f" long:"filter"     description:"filter for function names (regexes allowed)"`
}

// dburl constructs a database connection url
func (o *Options) dburl(database string) string {
	var tpl = "postgres://%s:%s@%s:%v/%s"
	return fmt.Sprintf(tpl, o.User, o.Pass, o.Host, o.Port, database)
}

var usage = `: introspect a postgres database's plpgsql functions`

func main() {

	var options Options
	var parser = flags.NewParser(&options, flags.Default)
	parser.Usage = usage

	// parse flags
	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}

	db := options.dburl(options.Database)
	results := dbquery(db, options.Searchpath, options.Filter)

	for _, r := range results {
		r.Arger()
		r.tpl = options.Template
		fmt.Println(r)
	}
}
