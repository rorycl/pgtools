package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

// QueryMock mocks DBQuery.Query
type QueryMock struct{}

func (q QueryMock) Query(ctx context.Context, label string, errorChan chan<- error, resultChan chan<- string) {
	return
}

// QueryMockReturn mocks DBQuery.Query
type QueryMockReturn struct{}

func (q QueryMockReturn) Query(ctx context.Context, label string, errorChan chan<- error, resultChan chan<- string) {
	resultChan <- "result"
	return
}

// QueryMockSlow mocks DBQuery.Query
type QueryMockSlow struct{}

func (q QueryMockSlow) Query(ctx context.Context, label string, errorChan chan<- error, resultChan chan<- string) {
	time.Sleep(10 * time.Millisecond)
	return
}

// QueryMockError mocks DBQuery.Query
type QueryMockError struct{}

func (q QueryMockError) Query(ctx context.Context, label string, errorChan chan<- error, resultChan chan<- string) {
	errorChan <- errors.New("mock error")
	return
}

// should receive an error for a TestQueryGroup with no
// DBQueries/Querier
func TestQueryGroupNoQueries(t *testing.T) {
	ctx := context.Background()
	counter := 0

	qg := NewDBQueryGroup("test1", 1, false)
	go qg.Process(ctx)

	timeout := make(chan struct{})
	go func() {
		time.Sleep(20 * time.Millisecond)
		timeout <- struct{}{}
	}()

LOOP:
	for {
		select {
		case e := <-qg.errorChan:
			t.Log(e)
			counter++
			break LOOP
		case <-qg.resultChan:
			break LOOP
		case <-qg.done:
			break LOOP
		case <-timeout:
			t.Errorf("hit done, should hit errorchan")
			break LOOP
		}
	}
	if counter != 1 {
		t.Errorf("counter %d should be 1", counter)
	}
}

// receive two errors
func TestQueryGroupErrorQueries(t *testing.T) {
	ctx := context.Background()
	counter := 0

	qg := NewDBQueryGroup("test1", 1, false) // this cycles
	qg.AddQuerier(QueryMockError{})
	qg.AddQuerier(QueryMockError{})
	go qg.Process(ctx)

	timeout := make(chan struct{})
	go func() {
		time.Sleep(20 * time.Millisecond)
		timeout <- struct{}{}
	}()

LOOP:
	for {
		select {
		case <-qg.errorChan:
			counter++ // does not fall through to done because of cycling
			if counter > 1 {
				break LOOP
			}
		case <-qg.resultChan:
			t.Errorf("hit resultchan, should hit errorchan")
			break LOOP
		case <-qg.done:
			break LOOP
		case <-timeout:
			t.Errorf("hit timeout, should hit errorchan")
			break LOOP
		}
	}
	if counter != 2 {
		t.Errorf("counter %d should be 1", counter)
	}
}

// Test with timeout
func TestQueryGroupTimout(t *testing.T) {
	ctx, cancel := context.WithDeadline(
		context.Background(),
		time.Now().Add(15*time.Millisecond),
	)
	defer cancel()

	qg := NewDBQueryGroup("test1", 1, false)
	qg.AddQuerier(QueryMockSlow{})
	qg.AddQuerier(QueryMockSlow{})
	go qg.Process(ctx)

	timeout := make(chan struct{})
	go func() {
		time.Sleep(20 * time.Millisecond)
		timeout <- struct{}{}
	}()

LOOP:
	for {
		select {
		case <-qg.errorChan:
			t.Errorf("hit errorchan, should hit errorchan")
			break LOOP
		case <-qg.resultChan:
			t.Errorf("hit resultchan, should hit errorchan")
			break LOOP
		case <-qg.done:
			t.Errorf("hit done, should hit errorchan")
			break LOOP
		case <-ctx.Done():
			cancel()
			break LOOP
		case <-timeout:
			t.Errorf("hit timeout, should hit errorchan")
			break LOOP
		}
	}
}

// Test cycling : no cycling
func TestQueryGroupNoCycle(t *testing.T) {
	ctx := context.Background()
	counter := 0

	qg := NewDBQueryGroup("test2", 1, true)
	qg.AddQuerier(QueryMockReturn{})
	qg.AddQuerier(QueryMockReturn{})
	go qg.Process(ctx)

	timeout := time.After(200 * time.Millisecond)

LOOP:
	for {
		select {
		case <-qg.errorChan:
			t.Errorf("hit errorchan, should hit resultchan")
			break LOOP
		case <-qg.resultChan:
			counter++
			if counter > 1 {
				break LOOP
			}
		case <-qg.done:
			break LOOP
		case <-timeout:
			t.Errorf("hit timeout, should hit resultchan")
			break LOOP
		}
	}
	if counter != 2 {
		t.Errorf("counter %d should be 2", counter)
	}
}

// Test cycling : 4 cycles
func TestQueryGroupCycle(t *testing.T) {
	ctx, cancel := context.WithDeadline(
		context.Background(),
		time.Now().Add(2*time.Millisecond),
	)
	defer cancel()
	counter := 0

	qg := NewDBQueryGroup("test4", 4, false)
	qg.AddQuerier(QueryMockReturn{})
	qg.AddQuerier(QueryMockReturn{})
	go qg.Process(ctx)

LOOP:
	for {
		select {
		case <-qg.errorChan:
			t.Errorf("hit errorchan, should hit ctx.Done")
			break LOOP
		case <-qg.resultChan:
			counter++ // falls through to qg.done
		case <-qg.done:
			t.Errorf("hit done, should hit ctx.Done")
			break LOOP
		case <-ctx.Done():
			cancel()
			break LOOP
		}
	}
	if counter < 20 {
		t.Errorf("counter %d should be >20", counter)
	}
}

// Test with cancel
func TestQueryGroupWithCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	qg := NewDBQueryGroup("test2", 1, false)
	qg.AddQuerier(QueryMockSlow{})
	qg.AddQuerier(QueryMockSlow{})
	go qg.Process(ctx)

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	timeout := make(chan struct{})
	go func() {
		time.Sleep(20 * time.Millisecond)
		timeout <- struct{}{}
	}()

LOOP:
	for {
		select {
		case <-qg.errorChan:
			t.Errorf("hit ctx.done, should hit ctx.done")
			break LOOP
		case <-qg.resultChan:
			t.Errorf("hit resultchan, should hit ctx.done")
			break LOOP
		case <-qg.done:
			t.Errorf("hit done, should hit ctx.done")
			break LOOP
		case <-timeout:
			t.Errorf("hit timeout, should hit ctx.done")
			break LOOP
		case <-ctx.Done():
			break LOOP
		}
	}
}
