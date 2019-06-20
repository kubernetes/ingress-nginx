package pool

import (
	"testing"
	"time"
)

func BenchmarkUnlimitedSmallRun(b *testing.B) {

	res := make([]WorkUnit, 10)

	b.ReportAllocs()

	pool := New()
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

func BenchmarkUnlimitedSmallCancel(b *testing.B) {

	res := make([]WorkUnit, 0, 20)

	b.ReportAllocs()

	pool := New()
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

func BenchmarkUnlimitedLargeCancel(b *testing.B) {

	res := make([]WorkUnit, 0, 1000)

	b.ReportAllocs()

	pool := New()
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

func BenchmarkUnlimitedLargeRun(b *testing.B) {

	res := make([]WorkUnit, 100)

	b.ReportAllocs()

	pool := New()
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

func BenchmarkUnlimitedBatchSmallRun(b *testing.B) {

	fn := func(wu WorkUnit) (interface{}, error) {
		time.Sleep(time.Millisecond * 500)
		if wu.IsCancelled() {
			return nil, nil
		}
		time.Sleep(time.Millisecond * 500)
		return 1, nil
	}

	pool := New()
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
