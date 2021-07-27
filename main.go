package main

import (
	"github.com/dt3zr/dlog/collector"
	"github.com/dt3zr/dlog/counters"
	"github.com/dt3zr/dlog/server"
	"github.com/dt3zr/dlog/store"
	"github.com/dt3zr/dlog/timer"
)

func main() {
	done := make(chan struct{})
	defer close(done)
	collectorResultStream := collector.New(done)
	timeStream := timer.NewForEveryMinute(done)
	countersResultStream := counters.New(collectorResultStream, timeStream, done)
	s := store.New(countersResultStream, done)
	server.NewHttpServer(s)
}
