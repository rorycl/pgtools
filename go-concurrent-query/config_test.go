package main

import (
	"os"
	"testing"
)

var testSettings = "config.yaml"

func TestTomlSettings(t *testing.T) {

	filer, err := os.ReadFile(testSettings)
	if err != nil {
		t.Errorf("could not read test file: %s", err)
	}

	y, err := LoadYaml(filer)
	if err != nil {
		t.Errorf("Could not parse yaml %v", err)
	}

	if _, ok := y["type1"]; !ok {
		t.Errorf("no type1 dbquerygroup found")
	}
	type1 := y["type1"]

	if len(type1.Databases) != 3 {
		t.Errorf("expected 3 type1 databases")
	}

	if len(type1.Queries) != 3 {
		t.Errorf("expected 3 type1 queries")
	}

	if type1.Concurrency != 3 {
		t.Errorf("expected concurrency to equal 3 for type1")
	}

	if type1.Iterations != 3 {
		t.Errorf("expected iterations to equal 3 for type1")
	}

	t.Logf("%+v\n", y)
}

// TestNoQueries tests that there is at least one query for the group
func TestNoQueries(t *testing.T) {

	inlineYaml := ` 
---
typeerror:
  databases:
    - db_type1_1
    - db_type1_2
    - db_type1_3
  concurrency: 4
  iterations: 3
`

	_, err := LoadYaml([]byte(inlineYaml))
	if err == nil {
		t.Error("yaml should error with no queries")
	}
}
