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
	qg := NewDBQueryGroup("test1", 1, false)
	done := make(chan struct{})
	go qg.Process(done)
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
	qg := NewDBQueryGroup("test2", 1, false)
	qg.AddQuerier(QueryMock{})
	done := make(chan struct{})
	go qg.Process(done)
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

func TestQueryGroupNoCycle(t *testing.T) {
	qg := NewDBQueryGroup("test2", 1, true)
	qg.AddQuerier(QueryMock{})
	qg.AddQuerier(QueryMock{})
	done := make(chan struct{})
	go qg.Process(done)
	counter := 0
	ta := time.After(time.Millisecond * 10)
LOOP:
	for {
		select {
		case <-qg.errorChan:
			counter++
			t.Errorf("hit errorchan, should hit done")
			break LOOP
		case <-done:
			break LOOP
		case <-ta:
			t.Errorf("hit time.after, should hit done")
			break LOOP
		}
	}
	if counter != 0 {
		t.Errorf("counter %d should be 0", counter)
	}
}
