package main

import (
	"os"
	"strings"
	"testing"
)

func failString(b bool) string {
	if b {
		return "should fail"
	}
	return "should succeed"
}

func TestParseOpts(t *testing.T) {

	for i, test := range []struct {
		msg    string
		args   string
		errors bool
	}{
		{
			msg:    "standard simple options",
			args:   `prog -u user -p pass -c config.yaml -e`,
			errors: false,
		},
		{
			msg:    "no user",
			args:   `prog -p pass -c config.yaml -e`,
			errors: true,
		},
		{
			msg:    "no password",
			args:   `prog -u user -c config.yaml -e`,
			errors: true,
		},
		{
			msg:    "invalid host",
			args:   `prog -u user -p pass -H nonsense -c config.yaml -e`,
			errors: true,
		},
		{
			msg:    "valid host",
			args:   `prog -u user -p pass -H 8.8.8.8 -c config.yaml -e`,
			errors: false,
		},
		{
			msg:    "invalid duration",
			args:   `prog -u user -p pass -H 8.8.8.8 -c config.yaml -d -1`,
			errors: true,
		},
		{
			msg:    "valid duration",
			args:   `prog -u user -p pass -H 8.8.8.8 -c config.yaml -d 3`,
			errors: false,
		},
	} {

		os.Args = strings.Fields(test.args)
		options, err := ParseOpts()
		if test.errors && err == nil {
			t.Errorf("test %d should fail", i)
		}
		if !test.errors && err != nil {
			t.Errorf("test %d should succeed (err %s)", i, err)
		}
		t.Logf("test %d %s (%s):\n", i, test.msg, failString(test.errors))
		t.Logf("  result: %+v\n", options)
	}
}
