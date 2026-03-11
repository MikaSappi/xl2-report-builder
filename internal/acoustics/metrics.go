package acoustics

import (
	"math"
	"time"
	"xl2-report-builder/internal/model"
)

// BroadbandLevel computes the total weighted level from per-band Z-weighted levels.
// For each band: add the weighting correction, convert to energy, sum, convert back to dB.
func BroadbandLevel(bands [model.BandCount]float64, weights [model.BandCount]float64) float64 {
	sumEnergy := 0.0
	for i := 0; i < model.BandCount; i++ {
		weighted := bands[i] + weights[i]
		sumEnergy += math.Pow(10, weighted/10.0)
	}
	return 10.0 * math.Log10(sumEnergy)
}

// LALevel computes the A-weighted broadband level from Z-weighted band data.
func LALevel(bands [model.BandCount]float64) float64 {
	return BroadbandLevel(bands, AWeighting)
}

// LCLevel computes the C-weighted broadband level from Z-weighted band data.
func LCLevel(bands [model.BandCount]float64) float64 {
	return BroadbandLevel(bands, CWeighting)
}

// LeqFromLevels computes the equivalent continuous level from a slice of dB levels.
// Leq = 10 * log10( (1/N) * sum(10^(Li/10)) )
func LeqFromLevels(levels []float64) float64 {
	if len(levels) == 0 {
		return 0
	}
	sumEnergy := 0.0
	for _, l := range levels {
		sumEnergy += math.Pow(10, l/10.0)
	}
	return 10.0 * math.Log10(sumEnergy/float64(len(levels)))
}

// ComputeIntervalResults groups samples by the given interval duration and computes
// LAeq and max LCeq for each interval.
func ComputeIntervalResults(samples []model.Sample, interval time.Duration) []model.IntervalResult {
	if len(samples) == 0 {
		return nil
	}

	start := samples[0].Timestamp
	var results []model.IntervalResult

	i := 0
	for i < len(samples) {
		intervalEnd := start.Add(interval)

		var laLevels []float64
		maxLC := math.Inf(-1)
		var maxLCTime time.Time

		for i < len(samples) && samples[i].Timestamp.Before(intervalEnd) {
			la := LALevel(samples[i].Bands)
			lc := LCLevel(samples[i].Bands)
			laLevels = append(laLevels, la)
			if lc > maxLC {
				maxLC = lc
				maxLCTime = samples[i].Timestamp
			}
			i++
		}

		if len(laLevels) > 0 {
			results = append(results, model.IntervalResult{
				StartTime: start,
				EndTime:   intervalEnd,
				LAeq:      LeqFromLevels(laLevels),
				LCeqMax:   maxLC,
				LCeqMaxAt: maxLCTime,
			})
		}

		start = intervalEnd
	}

	return results
}

// OverallLAeq computes the A-weighted Leq across all samples.
func OverallLAeq(samples []model.Sample) float64 {
	levels := make([]float64, len(samples))
	for i, s := range samples {
		levels[i] = LALevel(s.Bands)
	}
	return LeqFromLevels(levels)
}

// OverallLCeq computes the C-weighted Leq across all samples.
func OverallLCeq(samples []model.Sample) float64 {
	levels := make([]float64, len(samples))
	for i, s := range samples {
		levels[i] = LCLevel(s.Bands)
	}
	return LeqFromLevels(levels)
}

// MaxLCeq finds the maximum 1-second C-weighted level and its timestamp.
func MaxLCeq(samples []model.Sample) (float64, time.Time) {
	maxLC := math.Inf(-1)
	var maxTime time.Time
	for _, s := range samples {
		lc := LCLevel(s.Bands)
		if lc > maxLC {
			maxLC = lc
			maxTime = s.Timestamp
		}
	}
	return maxLC, maxTime
}
