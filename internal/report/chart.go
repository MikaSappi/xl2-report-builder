package report

import (
	"bytes"
	"time"
	"xl2-report-builder/internal/model"

	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

func RenderChart(results []model.IntervalResult, includeLAeq, includeLCeqMax bool) ([]byte, error) {
	var series []chart.Series

	if includeLAeq {
		var xvals []time.Time
		var yvals []float64
		for _, r := range results {
			xvals = append(xvals, r.StartTime)
			yvals = append(yvals, r.LAeq)
		}
		series = append(series, chart.TimeSeries{
			Name:    "LAeq dB(A)",
			XValues: xvals,
			YValues: yvals,
			Style: chart.Style{
				StrokeColor: drawing.ColorFromHex("2563eb"),
				StrokeWidth: 2,
			},
		})
	}

	if includeLCeqMax {
		var xvals []time.Time
		var yvals []float64
		for _, r := range results {
			xvals = append(xvals, r.StartTime)
			yvals = append(yvals, r.LCeqMax)
		}
		series = append(series, chart.TimeSeries{
			Name:    "LCeq,max dB(C)",
			XValues: xvals,
			YValues: yvals,
			Style: chart.Style{
				StrokeColor: drawing.ColorFromHex("dc2626"),
				StrokeWidth: 2,
				StrokeDashArray: []float64{5, 3},
			},
		})
	}

	graph := chart.Chart{
		Width:  1400,
		Height: 500,
		Background: chart.Style{
			Padding: chart.Box{Top: 20, Left: 10, Right: 20, Bottom: 20},
		},
		XAxis: chart.XAxis{
			Name: "Time",
			Style: chart.Style{
				FontSize: 10,
			},
			ValueFormatter: chart.TimeValueFormatterWithFormat("15:04"),
		},
		YAxis: chart.YAxis{
			Name: "Level [dB]",
			Style: chart.Style{
				FontSize: 10,
			},
		},
		Series: series,
	}

	graph.Elements = []chart.Renderable{
		chart.LegendLeft(&graph),
	}

	buf := &bytes.Buffer{}
	if err := graph.Render(chart.PNG, buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
