package pool

import (
	"sync"
	"testing"
	"time"

	. "gopkg.in/go-playground/assert.v1"
)

// NOTES:
// - Run "go test" to run tests
// - Run "gocov test | gocov report" to report on test converage by file
// - Run "gocov test | gocov annotate -" to report on all code and functions, those ,marked with "MISS" were never called
//
// or
//
// -- may be a good idea to change to output path to somewherelike /tmp
// go test -coverprofile cover.out && go tool cover -html=cover.out -o cover.html
//

func TestLimitedBatch(t *testing.T) {

	newFunc := func(i int) func(WorkUnit) (interface{}, error) {
		return func(WorkUnit) (interface{}, error) {
			time.Sleep(time.Second * 1)
			return i, nil
		}
	}

	pool := NewLimited(4)
	defer pool.Close()

	batch := pool.Batch()

	for i := 0; i < 4; i++ {
		batch.Queue(newFunc(i))
	}

	batch.QueueComplete()

	var count int

	for range batch.Results() {
		count++
	}

	Equal(t, count, 4)
}

func TestLimitedBatchGlobalPool(t *testing.T) {

	newFunc := func(i int) func(WorkUnit) (interface{}, error) {
		return func(WorkUnit) (interface{}, error) {
			time.Sleep(time.Second * 1)
			return i, nil
		}
	}

	batch := limitedGpool.Batch()

	for i := 0; i < 4; i++ {
		batch.Queue(newFunc(i))
	}

	batch.QueueComplete()

	var count int

	for range batch.Results() {
		count++
	}

	Equal(t, count, 4)
}

func TestLimitedBatchCancelItemsThrownAway(t *testing.T) {

	newFunc := func(i int) func(WorkUnit) (interface{}, error) {
		return func(WorkUnit) (interface{}, error) {
			time.Sleep(time.Second * 1)
			return i, nil
		}
	}

	pool := NewLimited(4)
	defer pool.Close()

	batch := pool.Batch()

	go func() {
		for i := 0; i < 40; i++ {
			batch.Queue(newFunc(i))
		}
	}()

	batch.Cancel()

	var count int

	for range batch.Results() {
		count++
	}

	NotEqual(t, count, 40)
}

func TestLimitedBatchCancelItemsCancelledAfterward(t *testing.T) {

	newFunc := func(i int) func(WorkUnit) (interface{}, error) {
		return func(WorkUnit) (interface{}, error) {
			time.Sleep(time.Second * 1)
			return i, nil
		}
	}

	pool := NewLimited(4)
	defer pool.Close()

	batch := pool.Batch()

	go func() {
		for i := 0; i < 40; i++ {
			batch.Queue(newFunc(i))
		}
	}()

	time.Sleep(time.Second * 2)
	batch.Cancel()

	var count int

	for range batch.Results() {
		count++
	}

	Equal(t, count, 40)
}

func TestLimitedBatchWaitAll(t *testing.T) {

	var count int
	var m sync.Mutex

	newFunc := func(i int) func(WorkUnit) (interface{}, error) {
		return func(WorkUnit) (interface{}, error) {
			time.Sleep(time.Second * 1)
			m.Lock()
			count++
			m.Unlock()
			return i, nil
		}
	}

	pool := NewLimited(4)
	defer pool.Close()

	batch := pool.Batch()

	go func() {

		for i := 0; i < 10; i++ {
			batch.Queue(newFunc(i))
		}

		batch.QueueComplete()
	}()

	batch.WaitAll()

	Equal(t, count, 10)
}
