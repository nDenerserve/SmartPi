package models

type TValue struct {
	Type  string  `json:"type" xml:"type"`
	Unity string  `json:"unity" xml:"unity"`
	Info  string  `json:"info" xml:"info"`
	Data  float32 `json:"data" xml:"data"`
}

type TPhase struct {
	Phase  int       `json:"phase" xml:"phase"`
	Name   string    `json:"name" xml:"name"`
	Values []*TValue `json:"values" xml:"values"`
}

type TDataset struct {
	Time   string    `json:"time" xml:"time"`
	Phases []*TPhase `json:"phases" xml:"phases"`
}

type TMeasurement struct {
	Serial          string      `json:"serial" xml:"serial"`
	Name            string      `json:"name" xml:"name"`
	Lat             float64     `json:"lat" xml:"lat"`
	Lng             float64     `json:"lng" xml:"lng"`
	Time            string      `json:"time" xml:"time"`
	Softwareversion string      `json:"softwareversion" xml:"softwareversion"`
	Ipaddress       string      `json:"ipaddress" xml:"ipaddress"`
	Datasets        []*TDataset `json:"datasets" xml:"datasets"`
}
