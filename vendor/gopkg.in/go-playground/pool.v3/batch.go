package pool

import "sync"

// Batch contains all information for a batch run of WorkUnits
type Batch interface {

	// Queue queues the work to be run in the pool and starts processing immediately
	// and also retains a reference for Cancellation and outputting to results.
	// WARNING be sure to call QueueComplete() once all work has been Queued.
	Queue(fn WorkFunc)

	// QueueComplete lets the batch know that there will be no more Work Units Queued
	// so that it may close the results channels once all work is completed.
	// WARNING: if this function is not called the results channel will never exhaust,
	// but block forever listening for more results.
	QueueComplete()

	// Cancel cancels the Work Units belonging to this Batch
	Cancel()

	// Results returns a Work Unit result channel that will output all
	// completed units of work.
	Results() <-chan WorkUnit

	// WaitAll is an alternative to Results() where you
	// may want/need to wait until all work has been
	// processed, but don't need to check results.
	// eg. individual units of work may handle their own
	// errors, logging...
	WaitAll()
}

// batch contains all information for a batch run of WorkUnits
type batch struct {
	pool    Pool
	m       sync.Mutex
	units   []WorkUnit
	results chan WorkUnit
	done    chan struct{}
	closed  bool
	wg      *sync.WaitGroup
}

func newBatch(p Pool) Batch {
	return &batch{
		pool:    p,
		units:   make([]WorkUnit, 0, 4), // capacity it to 4 so it doesn't grow and allocate too many times.
		results: make(chan WorkUnit),
		done:    make(chan struct{}),
		wg:      new(sync.WaitGroup),
	}
}

// Queue queues the work to be run in the pool and starts processing immediately
// and also retains a reference for Cancellation and outputting to results.
// WARNING be sure to call QueueComplete() once all work has been Queued.
func (b *batch) Queue(fn WorkFunc) {

	b.m.Lock()

	if b.closed {
		b.m.Unlock()
		return
	}

	wu := b.pool.Queue(fn)

	b.units = append(b.units, wu) // keeping a reference for cancellation purposes
	b.wg.Add(1)
	b.m.Unlock()

	go func(b *batch, wu WorkUnit) {
		wu.Wait()
		b.results <- wu
		b.wg.Done()
	}(b, wu)
}

// QueueComplete lets the batch know that there will be no more Work Units Queued
// so that it may close the results channels once all work is completed.
// WARNING: if this function is not called the results channel will never exhaust,
// but block forever listening for more results.
func (b *batch) QueueComplete() {
	b.m.Lock()
	b.closed = true
	close(b.done)
	b.m.Unlock()
}

// Cancel cancels the Work Units belonging to this Batch
func (b *batch) Cancel() {

	b.QueueComplete() // no more to be added

	b.m.Lock()

	// go in reverse order to try and cancel as many as possbile
	// one at end are less likely to have run than those at the beginning
	for i := len(b.units) - 1; i >= 0; i-- {
		b.units[i].Cancel()
	}

	b.m.Unlock()
}

// Results returns a Work Unit result channel that will output all
// completed units of work.
func (b *batch) Results() <-chan WorkUnit {

	go func(b *batch) {
		<-b.done
		b.m.Lock()
		b.wg.Wait()
		b.m.Unlock()
		close(b.results)
	}(b)

	return b.results
}

// WaitAll is an alternative to Results() where you
// may want/need to wait until all work has been
// processed, but don't need to check results.
// eg. individual units of work may handle their own
// errors and logging...
func (b *batch) WaitAll() {

	for range b.Results() {
	}
}
