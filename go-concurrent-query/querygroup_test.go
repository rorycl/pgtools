package main

import (
	"testing"
	"time"
)

// QueryMock mocks DBQuery.Query
type QueryMock struct{}

func (q QueryMock) Query(label string, errorChan chan<- error) {
	return
}

// should receive an error for a TestQueryGroup with no
// DBQueries/Querier
func TestQueryGroup1(t *testing.T) {
	qg := NewDBQueryGroup("test1", 1)
	go qg.Process()
	counter := 0
	ta := time.After(time.Millisecond * 1)
LOOP:
	for {
		select {
		case e := <-qg.errorChan:
			t.Log(e)
			counter++
			break LOOP
		case <-ta:
			break LOOP
		}
	}
	if counter != 1 {
		t.Errorf("counter %d should be 1", counter)
	}
}

func TestQueryGroup2(t *testing.T) {
	qg := NewDBQueryGroup("test2", 1)
	qg.AddQuerier(QueryMock{})
	go qg.Process()
	counter := 0
	ta := time.After(time.Millisecond * 1)
LOOP:
	for {
		select {
		case <-qg.errorChan:
			counter++
			break LOOP
		case <-ta:
			break LOOP
		}
	}
	if counter != 0 {
		t.Errorf("counter %d should be 0", counter)
	}
}
