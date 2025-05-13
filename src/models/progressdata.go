package models

import (
	"time"
)

type Progressdata struct {
	Device Device           `json:"device"`
	Data   []Timeseriesdata `json:"data"`
}

type Timeseriesdata struct {
	Field     string      `json:"field"`
	Datapoint []Datapoint `json:"datapoint"`
}

type Datapoint struct {
	Time  time.Time `json:"time"`
	Value float64   `json:"value"`
}

type Progressdatalist struct {
	Progressdatalist []Progressdata `json:"progressdatalist"`
}

func (progressdata *Progressdata) AddTimeseries(item Timeseriesdata) []Timeseriesdata {
	progressdata.Data = append(progressdata.Data, item)
	return progressdata.Data
}

func (timeseriesdata *Timeseriesdata) AddDatapoint(item Datapoint) []Datapoint {
	timeseriesdata.Datapoint = append(timeseriesdata.Datapoint, item)
	return timeseriesdata.Datapoint
}

func (progressdatalist *Progressdatalist) AddProgressdata(item Progressdata) []Progressdata {
	progressdatalist.Progressdatalist = append(progressdatalist.Progressdatalist, item)
	return progressdatalist.Progressdatalist
}

func (progressdatalist *Progressdatalist) AddDevicelist(item Devices) []Progressdata {
	for _, entry := range item.Devices {
		progdata := Progressdata{Device: entry}
		progressdatalist.Progressdatalist = append(progressdatalist.Progressdatalist, progdata)
	}
	return progressdatalist.Progressdatalist
}
