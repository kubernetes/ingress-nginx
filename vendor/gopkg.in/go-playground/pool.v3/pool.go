package pool

// Pool contains all information for a pool instance.
type Pool interface {

	// Queue queues the work to be run, and starts processing immediately
	Queue(fn WorkFunc) WorkUnit

	// Reset reinitializes a pool that has been closed/cancelled back to a working
	// state. if the pool has not been closed/cancelled, nothing happens as the pool
	// is still in a valid running state
	Reset()

	// Cancel cancels any pending work still not committed to processing.
	// Call Reset() to reinitialize the pool for use.
	Cancel()

	// Close cleans up pool data and cancels any pending work still not committed
	// to processing. Call Reset() to reinitialize the pool for use.
	Close()

	// Batch creates a new Batch object for queueing Work Units separate from any
	// others that may be running on the pool. Grouping these Work Units together
	// allows for individual Cancellation of the Batch Work Units without affecting
	// anything else running on the pool as well as outputting the results on a
	// channel as they complete. NOTE: Batch is not reusable, once QueueComplete()
	// has been called it's lifetime has been sealed to completing the Queued items.
	Batch() Batch
}

// WorkFunc is the function type needed by the pool for execution
type WorkFunc func(wu WorkUnit) (interface{}, error)
