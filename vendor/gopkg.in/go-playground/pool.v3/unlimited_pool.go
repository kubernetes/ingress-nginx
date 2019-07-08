package pool

import (
	"fmt"
	"math"
	"runtime"
	"sync"
)

var _ Pool = new(unlimitedPool)

// unlimitedPool contains all information for an unlimited pool instance.
type unlimitedPool struct {
	units  []*workUnit
	cancel chan struct{}
	closed bool
	m      sync.Mutex
}

// New returns a new unlimited pool instance
func New() Pool {

	p := &unlimitedPool{
		units: make([]*workUnit, 0, 4), // init capacity to 4, assuming if using pool, then probably a few have at least that many and will reduce array resizes
	}
	p.initialize()

	return p
}

func (p *unlimitedPool) initialize() {

	p.cancel = make(chan struct{})
	p.closed = false
}

// Queue queues the work to be run, and starts processing immediately
func (p *unlimitedPool) Queue(fn WorkFunc) WorkUnit {

	w := &workUnit{
		done: make(chan struct{}),
		fn:   fn,
	}

	p.m.Lock()

	if p.closed {
		w.err = &ErrPoolClosed{s: errClosed}
		// if w.cancelled.Load() == nil {
		close(w.done)
		// }
		p.m.Unlock()
		return w
	}

	p.units = append(p.units, w)
	go func(w *workUnit) {

		defer func(w *workUnit) {
			if err := recover(); err != nil {

				trace := make([]byte, 1<<16)
				n := runtime.Stack(trace, true)

				s := fmt.Sprintf(errRecovery, err, string(trace[:int(math.Min(float64(n), float64(7000)))]))

				w.cancelled.Store(struct{}{})
				w.err = &ErrRecovery{s: s}
				close(w.done)
			}
		}(w)

		// support for individual WorkUnit cancellation
		// and batch job cancellation
		if w.cancelled.Load() == nil {
			val, err := w.fn(w)

			w.writing.Store(struct{}{})

			// need to check again in case the WorkFunc cancelled this unit of work
			// otherwise we'll have a race condition
			if w.cancelled.Load() == nil && w.cancelling.Load() == nil {

				w.value, w.err = val, err

				// who knows where the Done channel is being listened to on the other end
				// don't want this to block just because caller is waiting on another unit
				// of work to be done first so we use close
				close(w.done)
			}
		}
	}(w)

	p.m.Unlock()

	return w
}

// Reset reinitializes a pool that has been closed/cancelled back to a working state.
// if the pool has not been closed/cancelled, nothing happens as the pool is still in
// a valid running state
func (p *unlimitedPool) Reset() {

	p.m.Lock()

	if !p.closed {
		p.m.Unlock()
		return
	}

	// cancelled the pool, not closed it, pool will be usable after calling initialize().
	p.initialize()
	p.m.Unlock()
}

func (p *unlimitedPool) closeWithError(err error) {

	p.m.Lock()

	if !p.closed {
		close(p.cancel)
		p.closed = true

		// clear out array values for garbage collection, but reuse array just in case going to reuse
		// go in reverse order to try and cancel as many as possbile
		// one at end are less likely to have run than those at the beginning
		for i := len(p.units) - 1; i >= 0; i-- {
			p.units[i].cancelWithError(err)
			p.units[i] = nil
		}

		p.units = p.units[0:0]
	}

	p.m.Unlock()
}

// Cancel cleans up the pool workers and channels and cancels and pending
// work still yet to be processed.
// call Reset() to reinitialize the pool for use.
func (p *unlimitedPool) Cancel() {

	err := &ErrCancelled{s: errCancelled}
	p.closeWithError(err)
}

// Close cleans up the pool workers and channels and cancels any pending
// work still yet to be processed.
// call Reset() to reinitialize the pool for use.
func (p *unlimitedPool) Close() {

	err := &ErrPoolClosed{s: errClosed}
	p.closeWithError(err)
}

// Batch creates a new Batch object for queueing Work Units separate from any others
// that may be running on the pool. Grouping these Work Units together allows for individual
// Cancellation of the Batch Work Units without affecting anything else running on the pool
// as well as outputting the results on a channel as they complete.
// NOTE: Batch is not reusable, once QueueComplete() has been called it's lifetime has been sealed
// to completing the Queued items.
func (p *unlimitedPool) Batch() Batch {
	return newBatch(p)
}
