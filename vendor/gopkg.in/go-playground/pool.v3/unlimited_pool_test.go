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

func TestUnlimitedPool(t *testing.T) {

	var res []WorkUnit

	pool := New()
	defer pool.Close()

	newFunc := func(d time.Duration) WorkFunc {
		return func(WorkUnit) (interface{}, error) {
			time.Sleep(d)
			return nil, nil
		}
	}

	for i := 0; i < 4; i++ {
		wu := pool.Queue(newFunc(time.Second * 1))
		res = append(res, wu)
	}

	var count int

	for _, wu := range res {
		wu.Wait()
		Equal(t, wu.Error(), nil)
		Equal(t, wu.Value(), nil)
		count++
	}

	Equal(t, count, 4)

	pool.Close() // testing no error occurs as Close will be called twice once defer pool.Close() fires
}

func TestUnlimitedCancel(t *testing.T) {

	m := new(sync.RWMutex)
	var closed bool
	c := make(chan WorkUnit, 100)

	pool := unlimitedGpool
	defer pool.Close()

	newFunc := func(d time.Duration) WorkFunc {
		return func(WorkUnit) (interface{}, error) {
			time.Sleep(d)
			return 1, nil
		}
	}

	go func(ch chan WorkUnit) {
		for i := 0; i < 40; i++ {

			go func(ch chan WorkUnit) {
				m.RLock()
				if !closed {
					ch <- pool.Queue(newFunc(time.Second * 1))
				}
				m.RUnlock()
			}(ch)
		}
	}(c)

	time.Sleep(time.Second * 1)
	pool.Cancel()
	m.Lock()
	closed = true
	close(c)
	m.Unlock()

	var count int

	for wu := range c {
		wu.Wait()

		if wu.Error() != nil {
			_, ok := wu.Error().(*ErrCancelled)
			if !ok {
				_, ok = wu.Error().(*ErrPoolClosed)
				if ok {
					Equal(t, wu.Error().Error(), "ERROR: Work Unit added/run after the pool had been closed or cancelled")
				}
			} else {
				Equal(t, wu.Error().Error(), "ERROR: Work Unit Cancelled")
			}

			Equal(t, ok, true)
			continue
		}

		count += wu.Value().(int)
	}

	NotEqual(t, count, 40)

	// reset and test again
	pool.Reset()

	wrk := pool.Queue(newFunc(time.Millisecond * 300))
	wrk.Wait()

	_, ok := wrk.Value().(int)
	Equal(t, ok, true)

	wrk = pool.Queue(newFunc(time.Millisecond * 300))
	time.Sleep(time.Second * 1)
	wrk.Cancel()
	wrk.Wait() // proving we don't get stuck here after cancel
	Equal(t, wrk.Error(), nil)

	pool.Reset() // testing that we can do this and nothing bad will happen as it checks if pool closed

	pool.Close()

	wu := pool.Queue(newFunc(time.Second * 1))
	wu.Wait()
	NotEqual(t, wu.Error(), nil)
	Equal(t, wu.Error().Error(), "ERROR: Work Unit added/run after the pool had been closed or cancelled")
}

func TestCancelFromWithin(t *testing.T) {
	pool := New()
	defer pool.Close()

	newFunc := func(d time.Duration) WorkFunc {
		return func(wu WorkUnit) (interface{}, error) {
			time.Sleep(d)
			if wu.IsCancelled() {
				return nil, nil
			}

			return 1, nil
		}
	}

	q := pool.Queue(newFunc(time.Second * 5))

	time.Sleep(time.Second * 2)
	pool.Cancel()

	Equal(t, q.Value() == nil, true)
	NotEqual(t, q.Error(), nil)
	Equal(t, q.Error().Error(), "ERROR: Work Unit Cancelled")
}

func TestUnlimitedPanicRecovery(t *testing.T) {

	pool := New()
	defer pool.Close()

	newFunc := func(d time.Duration, i int) WorkFunc {
		return func(WorkUnit) (interface{}, error) {
			if i == 1 {
				panic("OMG OMG OMG! something bad happened!")
			}
			time.Sleep(d)
			return 1, nil
		}
	}

	var wrk WorkUnit
	for i := 0; i < 4; i++ {
		time.Sleep(time.Second * 1)
		if i == 1 {
			wrk = pool.Queue(newFunc(time.Second*1, i))
			continue
		}
		pool.Queue(newFunc(time.Second*1, i))
	}
	wrk.Wait()

	NotEqual(t, wrk.Error(), nil)
	Equal(t, wrk.Error().Error()[0:90], "ERROR: Work Unit failed due to a recoverable error: 'OMG OMG OMG! something bad happened!'")
}
