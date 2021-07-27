package counters

import (
	"time"

	"github.com/dt3zr/dlog/collector"
)

type totalAndCount struct {
	total float64
	count int64
}

type parameterRunningTotal map[string]totalAndCount
type deviceRunningTotal map[string]parameterRunningTotal

type ParameterAverages map[string]float64
type DeviceAverages struct { // map[string]ParameterAverages
	Timestamp         time.Time
	MacId             string
	ParameterAverages ParameterAverages
}

func New(
	collectorResultStream <-chan collector.ParameterLog,
	timeStream <-chan struct{},
	done <-chan struct{}) <-chan DeviceAverages {

	countersResultStream := make(chan DeviceAverages)
	go func() {
		defer close(countersResultStream)
		drt := make(deviceRunningTotal)
		warmedUp := false
		for {
			select {
			case <-timeStream:
				if warmedUp {
					computeAndSendAverages(drt, countersResultStream)
				} else {
					warmedUp = true
				}
				drt = make(deviceRunningTotal)
			case paramLog := <-collectorResultStream:
				updateTotalAndCount(paramLog, drt)
			case <-done:
				return
			}
		}
	}()
	return countersResultStream
}

func computeAndSendAverages(drt deviceRunningTotal, countersResultStream chan<- DeviceAverages) {
	for macId, prt := range drt {
		pa := make(ParameterAverages)
		for pn, tac := range prt {
			pa[pn] = tac.total / float64(tac.count)
		}
		result := DeviceAverages{time.Now(), macId, pa}
		countersResultStream <- result
	}
}

func updateTotalAndCount(paramLog collector.ParameterLog, drt deviceRunningTotal) {
	prt, ok := drt[paramLog.MacId]
	if !ok {
		drt[paramLog.MacId] = make(parameterRunningTotal, 8)
		prt = drt[paramLog.MacId]
	}
	tac := prt[paramLog.Name]
	prt[paramLog.Name] = totalAndCount{tac.total + paramLog.Value, tac.count + 1}
}
