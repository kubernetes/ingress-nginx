package pool

import "sync/atomic"

// WorkUnit contains a single uint of works values
type WorkUnit interface {

	// Wait blocks until WorkUnit has been processed or cancelled
	Wait()

	// Value returns the work units return value
	Value() interface{}

	// Error returns the Work Unit's error
	Error() error

	// Cancel cancels this specific unit of work, if not already committed
	// to processing.
	Cancel()

	// IsCancelled returns if the Work Unit has been cancelled.
	// NOTE: After Checking IsCancelled(), if it returns false the
	// Work Unit can no longer be cancelled and will use your returned values.
	IsCancelled() bool
}

var _ WorkUnit = new(workUnit)

// workUnit contains a single unit of works values
type workUnit struct {
	value      interface{}
	err        error
	done       chan struct{}
	fn         WorkFunc
	cancelled  atomic.Value
	cancelling atomic.Value
	writing    atomic.Value
}

// Cancel cancels this specific unit of work, if not already committed to processing.
func (wu *workUnit) Cancel() {
	wu.cancelWithError(&ErrCancelled{s: errCancelled})
}

func (wu *workUnit) cancelWithError(err error) {

	wu.cancelling.Store(struct{}{})

	if wu.writing.Load() == nil && wu.cancelled.Load() == nil {
		wu.cancelled.Store(struct{}{})
		wu.err = err
		close(wu.done)
	}
}

// Wait blocks until WorkUnit has been processed or cancelled
func (wu *workUnit) Wait() {
	<-wu.done
}

// Value returns the work units return value
func (wu *workUnit) Value() interface{} {
	return wu.value
}

// Error returns the Work Unit's error
func (wu *workUnit) Error() error {
	return wu.err
}

// IsCancelled returns if the Work Unit has been cancelled.
// NOTE: After Checking IsCancelled(), if it returns false the
// Work Unit can no longer be cancelled and will use your returned values.
func (wu *workUnit) IsCancelled() bool {
	wu.writing.Store(struct{}{}) // ensure that after this check we are committed as cannot be cancelled if not aalready
	return wu.cancelled.Load() != nil
}
