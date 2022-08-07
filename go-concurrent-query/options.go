package main

import (
	"errors"
	"net"

	flags "github.com/jessevdk/go-flags"
)

// Options show flag options
type Options struct {
	User      string `short:"u" long:"user"     description:"database user" required:"true"`
	Pass      string `short:"p" long:"password" description:"database pass" required:"true"`
	Config    string `short:"c" long:"config"   description:"database query group yaml file" required:"true"`
	Port      int    `short:"P" long:"port"     description:"server port" default:"5432"`
	Host      string `short:"H" long:"host"     description:"server host" default:"127.0.0.1"`
	Duration  int    `short:"d" long:"duration" description:"limit test duration in seconds" default:"0"`
	DontCycle bool   `long:"dontcycle" description:"don't cycle databases, process each only once"`
	ErrExit   bool   `short:"e" long:"errexit"  description:"exit on first query err"`
}

var usage = `

Run queries concurrently on a set of Postgresql databases.`

// ParseOpts returns the filled options or error
func ParseOpts() (Options, error) {

	var options Options
	var parser = flags.NewParser(&options, flags.Default)
	parser.Usage = usage

	// parse flags
	if _, err := parser.Parse(); err != nil {
		return options, err
	}

	if options.User == "" || options.Pass == "" {
		return options, errors.New("Both database user and password must be supplied")
	}

	if net.ParseIP(options.Host) == nil {
		return options, errors.New("Invalid IP address for host")
	}

	if options.Duration < 0 {
		return options, errors.New("only 0 or positive duration seconds accepted")
	}

	return options, nil
}
