package pool

import (
	"testing"
	"time"
)

func BenchmarkLimitedSmallRun(b *testing.B) {

	res := make([]WorkUnit, 10)

	b.ReportAllocs()

	pool := NewLimited(10)
	defer pool.Close()

	fn := func(wu WorkUnit) (interface{}, error) {
		time.Sleep(time.Millisecond * 500)
		if wu.IsCancelled() {
			return nil, nil
		}
		time.Sleep(time.Millisecond * 500)
		return 1, nil
	}

	for i := 0; i < 10; i++ {
		res[i] = pool.Queue(fn)
	}

	var count int

	for _, cw := range res {

		cw.Wait()

		if cw.Error() == nil {
			count += cw.Value().(int)
		}
	}

	if count != 10 {
		b.Fatal("Count Incorrect")
	}
}

func BenchmarkLimitedSmallCancel(b *testing.B) {

	res := make([]WorkUnit, 0, 20)

	b.ReportAllocs()

	pool := NewLimited(4)
	defer pool.Close()

	newFunc := func(i int) WorkFunc {
		return func(wu WorkUnit) (interface{}, error) {
			time.Sleep(time.Millisecond * 500)
			if wu.IsCancelled() {
				return nil, nil
			}
			time.Sleep(time.Millisecond * 500)
			return i, nil
		}
	}

	for i := 0; i < 20; i++ {
		if i == 6 {
			pool.Cancel()
		}
		res = append(res, pool.Queue(newFunc(i)))
	}

	for _, wrk := range res {
		if wrk == nil {
			continue
		}
		wrk.Wait()
	}
}

func BenchmarkLimitedLargeCancel(b *testing.B) {

	res := make([]WorkUnit, 0, 1000)

	b.ReportAllocs()

	pool := NewLimited(4)
	defer pool.Close()

	newFunc := func(i int) WorkFunc {
		return func(wu WorkUnit) (interface{}, error) {
			time.Sleep(time.Millisecond * 500)
			if wu.IsCancelled() {
				return nil, nil
			}
			time.Sleep(time.Millisecond * 500)
			return i, nil
		}
	}

	for i := 0; i < 1000; i++ {
		if i == 6 {
			pool.Cancel()
		}
		res = append(res, pool.Queue(newFunc(i)))
	}

	for _, wrk := range res {
		if wrk == nil {
			continue
		}
		wrk.Wait()
	}
}

func BenchmarkLimitedOverconsumeLargeRun(b *testing.B) {

	res := make([]WorkUnit, 100)

	b.ReportAllocs()

	pool := NewLimited(25)
	defer pool.Close()

	newFunc := func(i int) WorkFunc {
		return func(wu WorkUnit) (interface{}, error) {
			time.Sleep(time.Millisecond * 500)
			if wu.IsCancelled() {
				return nil, nil
			}
			time.Sleep(time.Millisecond * 500)
			return 1, nil
		}
	}

	for i := 0; i < 100; i++ {
		res[i] = pool.Queue(newFunc(i))
	}

	var count int

	for _, cw := range res {

		cw.Wait()

		count += cw.Value().(int)
	}

	if count != 100 {
		b.Fatalf("Count Incorrect, Expected '100' Got '%d'", count)
	}
}

func BenchmarkLimitedBatchSmallRun(b *testing.B) {

	fn := func(wu WorkUnit) (interface{}, error) {
		time.Sleep(time.Millisecond * 500)
		if wu.IsCancelled() {
			return nil, nil
		}
		time.Sleep(time.Millisecond * 500)
		return 1, nil
	}

	pool := NewLimited(10)
	defer pool.Close()

	batch := pool.Batch()

	for i := 0; i < 10; i++ {
		batch.Queue(fn)
	}

	batch.QueueComplete()

	var count int

	for cw := range batch.Results() {
		count += cw.Value().(int)
	}

	if count != 10 {
		b.Fatal("Count Incorrect")
	}
}
