package parser

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"xl2-report-builder/internal/model"
)

type parseState int

const (
	stateHeader parseState = iota
	stateColumnHeaders
	stateData
	stateSummary
	stateChecksum
	stateDone
)

func ParseLogFile(path string) (*model.Measurement, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	m := &model.Measurement{
		Samples: make([]model.Sample, 0, 5000),
	}

	state := stateHeader
	colHeaderLines := 0

	for scanner.Scan() {
		line := scanner.Text()

		switch state {
		case stateHeader:
			state = parseHeaderLine(line, &m.Header, state)

		case stateColumnHeaders:
			colHeaderLines++
			if colHeaderLines >= 3 {
				state = stateData
			}

		case stateData:
			if strings.HasPrefix(line, "# RTA LOG Results over") {
				state = stateSummary
				continue
			}
			if strings.TrimSpace(line) == "" {
				continue
			}
			sample, err := parseDataRow(line, m.Header.StartTime.Location())
			if err != nil {
				return nil, fmt.Errorf("parse data row: %w", err)
			}
			m.Samples = append(m.Samples, sample)

		case stateSummary:
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			if strings.HasPrefix(trimmed, "#CheckSum") {
				state = stateChecksum
				continue
			}
			bands, err := parseSummaryRow(line)
			if err != nil {
				return nil, fmt.Errorf("parse summary row: %w", err)
			}
			m.SummaryLeq = bands

		case stateChecksum:
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				m.Checksum = trimmed
				state = stateDone
			}

		case stateDone:
			// ignore trailing lines
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan file: %w", err)
	}

	return m, nil
}

func parseHeaderLine(line string, h *model.Header, currentState parseState) parseState {
	trimmed := strings.TrimSpace(line)

	if strings.HasPrefix(trimmed, "# RTA LOG Results") && !strings.Contains(trimmed, "over") {
		return stateColumnHeaders
	}

	if strings.HasPrefix(trimmed, "XL2") && strings.Contains(trimmed, "Logging:") {
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) >= 3 {
			h.FileName = strings.TrimSpace(parts[len(parts)-1])
		}
		return stateHeader
	}

	if strings.Contains(line, "Device Info:") {
		h.DeviceInfo = extractValue(line)
	} else if strings.Contains(line, "Mic Sensitivity:") {
		h.MicSensitivity = extractValue(line)
	} else if strings.Contains(line, "calibrated") {
		h.CalibrationInfo = strings.TrimSpace(trimmed)
	} else if strings.Contains(line, "Time Zone:") {
		h.TimeZone = extractValue(line)
	} else if strings.Contains(line, "Profile:") {
		h.Profile = extractValue(line)
	} else if strings.Contains(line, "Timer mode:") {
		h.TimerMode = extractValue(line)
	} else if strings.Contains(line, "Log-Interval:") {
		h.LogInterval = extractValue(line)
	} else if strings.Contains(line, "Resolution:") {
		h.Resolution = extractValue(line)
	} else if strings.Contains(line, "Range:") {
		h.Range = extractValue(line)
	} else if strings.Contains(line, "Start:") && !strings.Contains(line, "start") {
		h.StartTime = parseDateTime(extractValue(line), h.TimeZone)
	} else if strings.Contains(line, "End:") {
		h.EndTime = parseDateTime(extractValue(line), h.TimeZone)
	}

	return stateHeader
}

func extractValue(line string) string {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func parseDateTime(s string, tz string) time.Time {
	s = strings.TrimSpace(s)
	loc := parseTimeZone(tz)
	t, err := time.ParseInLocation("2006-01-02, 15:04:05", s, loc)
	if err != nil {
		return time.Time{}
	}
	return t
}

func parseTimeZone(tz string) *time.Location {
	if tz == "" {
		return time.UTC
	}
	// Try to extract offset like "UTC+02:00"
	tz = strings.TrimSpace(tz)
	if idx := strings.Index(tz, "("); idx >= 0 {
		tz = strings.TrimSpace(tz[:idx])
	}

	if strings.HasPrefix(tz, "UTC") {
		offsetStr := strings.TrimPrefix(tz, "UTC")
		if offsetStr == "" {
			return time.UTC
		}
		sign := 1
		if offsetStr[0] == '-' {
			sign = -1
			offsetStr = offsetStr[1:]
		} else if offsetStr[0] == '+' {
			offsetStr = offsetStr[1:]
		}
		parts := strings.Split(offsetStr, ":")
		hours, _ := strconv.Atoi(parts[0])
		minutes := 0
		if len(parts) > 1 {
			minutes, _ = strconv.Atoi(parts[1])
		}
		offset := sign * (hours*3600 + minutes*60)
		return time.FixedZone(fmt.Sprintf("UTC%+d", sign*hours), offset)
	}

	return time.UTC
}

func parseDataRow(line string, loc *time.Location) (model.Sample, error) {
	fields := strings.Split(line, "\t")

	// Clean up fields
	for i := range fields {
		fields[i] = strings.TrimSpace(fields[i])
	}

	// Remove empty leading field if present
	if len(fields) > 0 && fields[0] == "" {
		fields = fields[1:]
	}

	// Need at least: Date, Time, Timer, empty band label, 36 band values = 40 fields
	if len(fields) < 40 {
		return model.Sample{}, fmt.Errorf("expected at least 40 fields, got %d", len(fields))
	}

	dateStr := fields[0]
	timeStr := fields[1]
	timerStr := fields[2]

	ts, err := time.ParseInLocation("2006-01-02 15:04:05", dateStr+" "+timeStr, loc)
	if err != nil {
		return model.Sample{}, fmt.Errorf("parse timestamp %q %q: %w", dateStr, timeStr, err)
	}

	timer, err := parseTimerDuration(timerStr)
	if err != nil {
		return model.Sample{}, fmt.Errorf("parse timer %q: %w", timerStr, err)
	}

	var bands [model.BandCount]float64
	// Fields 3 is the empty band label column, bands start at field 4
	bandStart := 4
	if fields[3] != "" {
		// If field 3 is not empty, bands might start at 3
		bandStart = 3
	}

	for i := 0; i < model.BandCount; i++ {
		idx := bandStart + i
		if idx >= len(fields) {
			return model.Sample{}, fmt.Errorf("missing band %d (field %d)", i, idx)
		}
		bands[i], err = parseEuropeanFloat(fields[idx])
		if err != nil {
			return model.Sample{}, fmt.Errorf("parse band %d value %q: %w", i, fields[idx], err)
		}
	}

	return model.Sample{
		Timestamp: ts,
		Timer:     timer,
		Bands:     bands,
	}, nil
}

func parseSummaryRow(line string) ([model.BandCount]float64, error) {
	fields := strings.Split(line, "\t")
	for i := range fields {
		fields[i] = strings.TrimSpace(fields[i])
	}
	if len(fields) > 0 && fields[0] == "" {
		fields = fields[1:]
	}

	var bands [model.BandCount]float64
	// Summary row has same structure: Date, Time, Timer, empty, then 36 bands
	bandStart := 4
	if len(fields) > 3 && fields[3] != "" {
		bandStart = 3
	}

	for i := 0; i < model.BandCount; i++ {
		idx := bandStart + i
		if idx >= len(fields) {
			return bands, fmt.Errorf("summary: missing band %d", i)
		}
		var err error
		bands[i], err = parseEuropeanFloat(fields[idx])
		if err != nil {
			return bands, fmt.Errorf("summary: parse band %d value %q: %w", i, fields[idx], err)
		}
	}

	return bands, nil
}

func parseEuropeanFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	s = strings.Replace(s, ",", ".", 1)
	return strconv.ParseFloat(s, 64)
}

func parseTimerDuration(s string) (time.Duration, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("expected HH:MM:SS, got %q", s)
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}
	sec, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, err
	}
	return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(sec)*time.Second, nil
}
