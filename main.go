package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

	_ "net/http/pprof"

	"github.com/eclesh/welford"
)

func send(ctx context.Context, dataCh chan<- time.Time) {
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
loop:
	for range ticker.C {
		select {
		case dataCh <- time.Now():
		case <-ctx.Done():
			break loop
		}
	}

	close(dataCh)
}

func recv(dataCh <-chan time.Time, collectorCh chan<- time.Duration) {
	for sentAt := range dataCh {
		collectorCh <- time.Since(sentAt)
	}
	close(collectorCh)
}

func pair(ctx context.Context, n int) *welford.Stats {
	collectorChs := make([]chan time.Duration, n)
	for i := 0; i < n; i++ {
		dataCh := make(chan time.Time, 1)
		collectorChs[i] = make(chan time.Duration, 1)
		go send(ctx, dataCh)
		go recv(dataCh, collectorChs[i])
	}

	stats := welford.New()
	running := true
	for running {
		running = false
		for i := 0; i < n; i++ {
			ch := collectorChs[i]
			if ch == nil {
				continue
			}

			dur, ok := <-ch
			if !ok {
				collectorChs[i] = nil
				continue
			}

			stats.Add(float64(dur.Microseconds()))
			running = true
		}
	}

	return stats
}

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	fmt.Println("running with ", runtime.NumCPU(), " cores")
	stats := pair(ctx, runtime.NumCPU()-1)
	fmt.Printf("mean=%fus,stddev=%fus,min=%fus,max=%fus\n", stats.Mean(), stats.Stddev(), stats.Min(), stats.Max())
}
