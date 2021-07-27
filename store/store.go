package store

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/dt3zr/dlog/counters"
)

const maxRecordCount = 60

type ParameterAverages map[string]float64

type DeviceAverages struct {
	timestamp time.Time
	paramAvgs ParameterAverages
}

type IndexedDeviceAverages struct {
	index    int
	averages [maxRecordCount]DeviceAverages
}

type Store map[string]IndexedDeviceAverages

func (s Store) GetTimeSeriesAsJSON(macId string) ([]byte, error) {
	// if devAvgs, ok := s[macId]; ok {
	if indexed, ok := s[macId]; ok {
		log.Printf("Get %s <- %v", macId, indexed.averages)

		type marshallableDevAvg struct {
			Timestamp time.Time          `json:"timestamp"`
			ParamAvgs map[string]float64 `json:"parameterAverages"`
		}

		type marshallableDevAvgs struct {
			MacId      string               `json:"macId"`
			TimeSeries []marshallableDevAvg `json:"timeseries"`
		}

		mdas := marshallableDevAvgs{MacId: macId}

		for _, devAvg := range indexed.averages {
			mda := marshallableDevAvg{
				devAvg.timestamp,
				devAvg.paramAvgs,
			}
			mdas.TimeSeries = append(mdas.TimeSeries, mda)
		}
		output, err := json.Marshal(mdas)
		return output, err
	}
	return nil, fmt.Errorf("MAC Id %s not found", macId)
}

func (s Store) String() string {
	output := make([]string, 0)
	// for macId, devAvgs := range s {
	for macId, indexed := range s {
		output = append(output, fmt.Sprintf("macId = %s", macId))
		for i, devAvg := range indexed.averages {
			rowOutput := make([]string, 0)
			for paramName, paramValue := range devAvg.paramAvgs {
				rowOutput = append(rowOutput, fmt.Sprintf("%s = %f", paramName, paramValue))
			}
			output = append(output, fmt.Sprintf("[id = %d, %s, %s]", i, devAvg.timestamp, strings.Join(rowOutput, ",")))
		}
	}
	return strings.Join(output, "\n")
}

func New(deviceAverageStream <-chan counters.DeviceAverages, done <-chan struct{}) Store {
	averageStore := make(Store)
	go func() {
		for {
			select {
			case da := <-deviceAverageStream:
				record := DeviceAverages{da.Timestamp, ParameterAverages(da.ParameterAverages)}
				indexed := averageStore[da.MacId]
				indexed.averages[indexed.index] = record
				indexed.updateIndex()
				averageStore[da.MacId] = indexed
			case <-done:
				return
			}
		}
	}()
	return averageStore
}

func (indexed *IndexedDeviceAverages) updateIndex() {
	indexed.index = (indexed.index + 1) % maxRecordCount
}
