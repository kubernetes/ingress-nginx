/*
Package pool implements a limited consumer goroutine or unlimited goroutine pool for easier goroutine handling and cancellation.


Features:

    - Dead simple to use and makes no assumptions about how you will use it.
    - Automatic recovery from consumer goroutines which returns an error to
      the results

Pool v2 advantages over Pool v1:

    - Up to 300% faster due to lower contention,
      BenchmarkSmallRun used to take 3 seconds, now 1 second
    - Cancels are much faster
    - Easier to use, no longer need to know the # of Work Units to be processed.
    - Pool can now be used as a long running/globally defined pool if desired,
      v1 Pool was only good for one run
    - Supports single units of work as well as batching
    - Pool can easily be reset after a Close() or Cancel() for reuse.
    - Multiple Batches can be run and even cancelled on the same Pool.
    - Supports individual Work Unit cancellation.

Pool v3 advantages over Pool v2:

    - Objects are not interfaces allowing for less breaking changes going forward.
    - Now there are 2 Pool types, both completely interchangeable, a limited worker pool
      and unlimited pool.
    - Simpler usage of Work Units, instead of `<-work.Done` now can do `work.Wait()`

Important Information READ THIS!

important usage information

    - It is recommended that you cancel a pool or batch from the calling
      function and not inside of the Unit of Work, it will work fine, however
      because of the goroutine scheduler and context switching it may not
      cancel as soon as if called from outside.

    - When Batching DO NOT FORGET TO CALL batch.QueueComplete(),
      if you do the Batch WILL deadlock

    - It is your responsibility to call WorkUnit.IsCancelled() to check if it's cancelled
      after a blocking operation like waiting for a connection from a pool. (optional)


Usage and documentation

both Limited Pool and Unlimited Pool have the same signatures and are completely interchangeable.

Per Unit Work

    package main

    import (
        "fmt"
        "time"

        "gopkg.in/go-playground/pool.v3"
    )

    func main() {

        p := pool.NewLimited(10)
        defer p.Close()

        user := p.Queue(getUser(13))
        other := p.Queue(getOtherInfo(13))

        user.Wait()
        if err := user.Error(); err != nil {
            // handle error
        }

        // do stuff with user
        username := user.Value().(string)
        fmt.Println(username)

        other.Wait()
        if err := other.Error(); err != nil {
            // handle error
        }

        // do stuff with other
        otherInfo := other.Value().(string)
        fmt.Println(otherInfo)
    }

    func getUser(id int) pool.WorkFunc {

        return func(wu pool.WorkUnit) (interface{}, error) {

            // simulate waiting for something, like TCP connection to be established
            // or connection from pool grabbed
            time.Sleep(time.Second * 1)

            if wu.IsCancelled() {
                // return values not used
                return nil, nil
            }

            // ready for processing...

            return "Joeybloggs", nil
        }
    }

    func getOtherInfo(id int) pool.WorkFunc {

        return func(wu pool.WorkUnit) (interface{}, error) {

            // simulate waiting for something, like TCP connection to be established
            // or connection from pool grabbed
            time.Sleep(time.Second * 1)

            if wu.IsCancelled() {
                // return values not used
                return nil, nil
            }

            // ready for processing...

            return "Other Info", nil
        }
    }


Batch Work

    package main

    import (
        "fmt"
        "time"

        "gopkg.in/go-playground/pool.v3"
    )

    func main() {

        p := pool.NewLimited(10)
        defer p.Close()

        batch := p.Batch()

        // for max speed Queue in another goroutine
        // but it is not required, just can't start reading results
        // until all items are Queued.

        go func() {
            for i := 0; i < 10; i++ {
                batch.Queue(sendEmail("email content"))
            }

            // DO NOT FORGET THIS OR GOROUTINES WILL DEADLOCK
            // if calling Cancel() it calles QueueComplete() internally
            batch.QueueComplete()
        }()

        for email := range batch.Results() {

            if err := email.Error(); err != nil {
                // handle error
                // maybe call batch.Cancel()
            }

            // use return value
            fmt.Println(email.Value().(bool))
        }
    }

    func sendEmail(email string) pool.WorkFunc {
        return func(wu pool.WorkUnit) (interface{}, error) {

            // simulate waiting for something, like TCP connection to be established
            // or connection from pool grabbed
            time.Sleep(time.Second * 1)

            if wu.IsCancelled() {
                // return values not used
                return nil, nil
            }

            // ready for processing...

            return true, nil // everything ok, send nil, error if not
        }
    }


Benchmarks

Run on MacBook Pro (Retina, 15-inch, Late 2013) 2.6 GHz Intel Core i7 16 GB 1600 MHz DDR3 using Go 1.6.2

run with 1, 2, 4,8 and 16 cpu to show it scales well...16 is double the # of logical cores on this machine.

NOTE: Cancellation times CAN vary depending how busy your system is and how the goroutine scheduler is but
worse case I've seen is 1 second to cancel instead of 0ns

    go test -cpu=1,2,4,8,16 -bench=. -benchmem=true
    PASS
    BenchmarkLimitedSmallRun                       1    1002492008 ns/op        3552 B/op         55 allocs/op
    BenchmarkLimitedSmallRun-2                     1    1002347196 ns/op        3568 B/op         55 allocs/op
    BenchmarkLimitedSmallRun-4                     1    1010533571 ns/op        4720 B/op         73 allocs/op
    BenchmarkLimitedSmallRun-8                     1    1008883324 ns/op        4080 B/op         63 allocs/op
    BenchmarkLimitedSmallRun-16                    1    1002317677 ns/op        3632 B/op         56 allocs/op
    BenchmarkLimitedSmallCancel             2000000000           0.00 ns/op        0 B/op          0 allocs/op
    BenchmarkLimitedSmallCancel-2           2000000000           0.00 ns/op        0 B/op          0 allocs/op
    BenchmarkLimitedSmallCancel-4           2000000000           0.00 ns/op        0 B/op          0 allocs/op
    BenchmarkLimitedSmallCancel-8           2000000000           0.00 ns/op        0 B/op          0 allocs/op
    BenchmarkLimitedSmallCancel-16          2000000000           0.00 ns/op        0 B/op          0 allocs/op
    BenchmarkLimitedLargeCancel             2000000000           0.00 ns/op        0 B/op          0 allocs/op
    BenchmarkLimitedLargeCancel-2           2000000000           0.00 ns/op        0 B/op          0 allocs/op
    BenchmarkLimitedLargeCancel-4           2000000000           0.00 ns/op        0 B/op          0 allocs/op
    BenchmarkLimitedLargeCancel-8            1000000          1006 ns/op           0 B/op          0 allocs/op
    BenchmarkLimitedLargeCancel-16          2000000000           0.00 ns/op        0 B/op          0 allocs/op
    BenchmarkLimitedOverconsumeLargeRun            1    4027153081 ns/op       36176 B/op        572 allocs/op
    BenchmarkLimitedOverconsumeLargeRun-2          1    4003489261 ns/op       32336 B/op        512 allocs/op
    BenchmarkLimitedOverconsumeLargeRun-4          1    4005579847 ns/op       34128 B/op        540 allocs/op
    BenchmarkLimitedOverconsumeLargeRun-8          1    4004639857 ns/op       34992 B/op        553 allocs/op
    BenchmarkLimitedOverconsumeLargeRun-16         1    4022695297 ns/op       36864 B/op        532 allocs/op
    BenchmarkLimitedBatchSmallRun                  1    1000785511 ns/op        6336 B/op         94 allocs/op
    BenchmarkLimitedBatchSmallRun-2                1    1001459945 ns/op        4480 B/op         65 allocs/op
    BenchmarkLimitedBatchSmallRun-4                1    1002475371 ns/op        6672 B/op         99 allocs/op
    BenchmarkLimitedBatchSmallRun-8                1    1002498902 ns/op        4624 B/op         67 allocs/op
    BenchmarkLimitedBatchSmallRun-16               1    1002202273 ns/op        5344 B/op         78 allocs/op
    BenchmarkUnlimitedSmallRun                     1    1002361538 ns/op        3696 B/op         59 allocs/op
    BenchmarkUnlimitedSmallRun-2                   1    1002230293 ns/op        3776 B/op         60 allocs/op
    BenchmarkUnlimitedSmallRun-4                   1    1002148953 ns/op        3776 B/op         60 allocs/op
    BenchmarkUnlimitedSmallRun-8                   1    1002120679 ns/op        3584 B/op         57 allocs/op
    BenchmarkUnlimitedSmallRun-16                  1    1001698519 ns/op        3968 B/op         63 allocs/op
    BenchmarkUnlimitedSmallCancel           2000000000           0.00 ns/op        0 B/op          0 allocs/op
    BenchmarkUnlimitedSmallCancel-2         2000000000           0.00 ns/op        0 B/op          0 allocs/op
    BenchmarkUnlimitedSmallCancel-4         2000000000           0.00 ns/op        0 B/op          0 allocs/op
    BenchmarkUnlimitedSmallCancel-8         2000000000           0.00 ns/op        0 B/op          0 allocs/op
    BenchmarkUnlimitedSmallCancel-16        2000000000           0.00 ns/op        0 B/op          0 allocs/op
    BenchmarkUnlimitedLargeCancel           2000000000           0.00 ns/op        0 B/op          0 allocs/op
    BenchmarkUnlimitedLargeCancel-2         2000000000           0.00 ns/op        0 B/op          0 allocs/op
    BenchmarkUnlimitedLargeCancel-4         2000000000           0.00 ns/op        0 B/op          0 allocs/op
    BenchmarkUnlimitedLargeCancel-8         2000000000           0.00 ns/op        0 B/op          0 allocs/op
    BenchmarkUnlimitedLargeCancel-16        2000000000           0.00 ns/op        0 B/op          0 allocs/op
    BenchmarkUnlimitedLargeRun                     1    1001631711 ns/op       40352 B/op        603 allocs/op
    BenchmarkUnlimitedLargeRun-2                   1    1002603908 ns/op       38304 B/op        586 allocs/op
    BenchmarkUnlimitedLargeRun-4                   1    1001452975 ns/op       38192 B/op        584 allocs/op
    BenchmarkUnlimitedLargeRun-8                   1    1005382882 ns/op       35200 B/op        537 allocs/op
    BenchmarkUnlimitedLargeRun-16                  1    1001818482 ns/op       37056 B/op        566 allocs/op
    BenchmarkUnlimitedBatchSmallRun                1    1002391247 ns/op        4240 B/op         63 allocs/op
    BenchmarkUnlimitedBatchSmallRun-2              1    1010313222 ns/op        4688 B/op         70 allocs/op
    BenchmarkUnlimitedBatchSmallRun-4              1    1008364651 ns/op        4304 B/op         64 allocs/op
    BenchmarkUnlimitedBatchSmallRun-8              1    1001858192 ns/op        4448 B/op         66 allocs/op
    BenchmarkUnlimitedBatchSmallRun-16             1    1001228000 ns/op        4320 B/op         64 allocs/op

To put some of these benchmarks in perspective:

    - BenchmarkLimitedSmallRun did 10 seconds worth of processing in 1.002492008s
    - BenchmarkLimitedSmallCancel ran 20 jobs, cancelled on job 6 and and ran in 0s
    - BenchmarkLimitedLargeCancel ran 1000 jobs, cancelled on job 6 and and ran in 0s
    - BenchmarkLimitedOverconsumeLargeRun ran 100 jobs using 25 workers in 4.027153081s

*/
package pool
