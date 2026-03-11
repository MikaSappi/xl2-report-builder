package model

import "time"

const BandCount = 36

var CenterFrequencies = [BandCount]float64{
	6.3, 8.0, 10.0, 12.5, 16.0, 20.0, 25.0, 31.5,
	40.0, 50.0, 63.0, 80.0, 100.0, 125.0, 160.0, 200.0,
	250.0, 315.0, 400.0, 500.0, 630.0, 800.0, 1000.0, 1250.0,
	1600.0, 2000.0, 2500.0, 3150.0, 4000.0, 5000.0, 6300.0, 8000.0,
	10000.0, 12500.0, 16000.0, 20000.0,
}

type Header struct {
	FileName        string
	DeviceInfo      string
	MicSensitivity  string
	CalibrationInfo string
	TimeZone        string
	Profile         string
	TimerMode       string
	LogInterval     string
	Resolution      string
	Range           string
	StartTime       time.Time
	EndTime         time.Time
}

type Sample struct {
	Timestamp time.Time
	Timer     time.Duration
	Bands     [BandCount]float64
}

type Measurement struct {
	Header     Header
	Samples    []Sample
	SummaryLeq [BandCount]float64
	Checksum   string
}

type IntervalResult struct {
	StartTime time.Time
	EndTime   time.Time
	LAeq      float64
	LCeqMax   float64
	LCeqMaxAt time.Time
}

type ReportData struct {
	Measurement      *Measurement
	IntervalResults  []IntervalResult
	IntervalDuration time.Duration
	OverallLAeq      float64
	OverallLCeq      float64
	MaxLCeq          float64
	MaxLCeqTime      time.Time
	IncludeLAeq      bool
	IncludeLCeqMax   bool
	CompanyName      string
	LogoPath         string
}
