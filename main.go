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

func recv(dataCh <-chan time.Time) *welford.Stats {
	stats := welford.New()
	for sentAt := range dataCh {
		stats.Add(float64(time.Since(sentAt).Microseconds()))
	}
	return stats
}

func work() {
	for {
	}
}

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	fmt.Println("running with ", runtime.NumCPU(), " cores")
	dataCh := make(chan time.Time, 1)
	for i := 0; i < runtime.NumCPU()-2; i++ {
		go work()
	}
	go send(ctx, dataCh)
	stats := recv(dataCh)
	fmt.Printf("mean=%fus,stddev=%fus,min=%fus,max=%fus\n", stats.Mean(), stats.Stddev(), stats.Min(), stats.Max())
}
