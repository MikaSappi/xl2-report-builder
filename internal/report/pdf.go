package report

import (
	"bytes"
	"fmt"
	"math"
	"time"
	"xl2-report-builder/internal/model"

	"github.com/go-pdf/fpdf"
)

func GeneratePDF(data *model.ReportData, outputPath string) error {
	pdf := fpdf.New("L", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 20)

	// Register logo if provided
	if data.LogoPath != "" {
		pdf.RegisterImage(data.LogoPath, "")
	}

	// Footer on every page
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Helvetica", "", 8)
		pdf.SetTextColor(140, 140, 140)
		footerText := fmt.Sprintf("Page %d", pdf.PageNo())
		if data.CompanyName != "" {
			footerText = fmt.Sprintf("%s  |  Page %d", data.CompanyName, pdf.PageNo())
		}
		pdf.CellFormat(0, 8, footerText, "", 0, "C", false, 0, "")
		pdf.SetTextColor(0, 0, 0)
	})

	renderSummaryPage(pdf, data)
	renderChartPage(pdf, data)
	renderDataTable(pdf, data)

	return pdf.OutputFileAndClose(outputPath)
}

func renderSummaryPage(pdf *fpdf.Fpdf, data *model.ReportData) {
	pdf.AddPage()
	h := &data.Measurement.Header

	// Logo + company name header
	headerY := pdf.GetY()
	if data.LogoPath != "" {
		pdf.ImageOptions(data.LogoPath, 10, headerY, 30, 0, false, fpdf.ImageOptions{ReadDpi: true}, 0, "")
	}
	if data.CompanyName != "" {
		nameX := 10.0
		if data.LogoPath != "" {
			nameX = 44
		}
		pdf.SetFont("Helvetica", "B", 14)
		pdf.SetXY(nameX, headerY+2)
		pdf.CellFormat(0, 10, data.CompanyName, "", 1, "L", false, 0, "")
	}
	if data.LogoPath != "" || data.CompanyName != "" {
		pdf.Ln(6)
	}

	// Title
	pdf.SetFont("Helvetica", "B", 20)
	pdf.CellFormat(0, 14, "Sound Level Measurement Report", "", 1, "C", false, 0, "")
	pdf.Ln(8)

	// Device / measurement info
	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(0, 7, "Measurement Information", "", 1, "L", false, 0, "")
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(pdf.GetX(), pdf.GetY(), pdf.GetX()+267, pdf.GetY())
	pdf.Ln(2)

	infoRows := []struct{ label, value string }{
		{"Device", h.DeviceInfo},
		{"Microphone", h.MicSensitivity},
		{"Calibration", h.CalibrationInfo},
		{"Time Zone", h.TimeZone},
		{"Profile", h.Profile},
		{"Resolution", h.Resolution},
		{"Range", h.Range},
		{"Log Interval", h.LogInterval},
		{"Start", h.StartTime.Format("2006-01-02 15:04:05")},
		{"End", h.EndTime.Format("2006-01-02 15:04:05")},
		{"Duration", formatDuration(h.EndTime.Sub(h.StartTime))},
		{"Samples", fmt.Sprintf("%d", len(data.Measurement.Samples))},
	}

	pdf.SetFont("Helvetica", "", 9)
	for i, row := range infoRows {
		if i%2 == 0 {
			pdf.SetFillColor(245, 245, 245)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		pdf.CellFormat(55, 6, row.label, "", 0, "L", true, 0, "")
		pdf.CellFormat(212, 6, row.value, "", 1, "L", true, 0, "")
	}

	pdf.Ln(8)

	// Summary metrics
	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(0, 7, "Summary Results", "", 1, "L", false, 0, "")
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(pdf.GetX(), pdf.GetY(), pdf.GetX()+267, pdf.GetY())
	pdf.Ln(2)

	pdf.SetFont("Helvetica", "", 10)

	if data.IncludeLAeq {
		pdf.SetFillColor(245, 245, 245)
		pdf.CellFormat(55, 7, "Overall LAeq", "", 0, "L", true, 0, "")
		pdf.SetFont("Helvetica", "B", 10)
		pdf.CellFormat(212, 7, fmt.Sprintf("%.1f dB(A)", data.OverallLAeq), "", 1, "L", true, 0, "")
		pdf.SetFont("Helvetica", "", 10)
	}

	if data.IncludeLCeqMax {
		pdf.SetFillColor(255, 255, 255)
		pdf.CellFormat(55, 7, "Overall LCeq", "", 0, "L", true, 0, "")
		pdf.SetFont("Helvetica", "B", 10)
		pdf.CellFormat(212, 7, fmt.Sprintf("%.1f dB(C)", data.OverallLCeq), "", 1, "L", true, 0, "")
		pdf.SetFont("Helvetica", "", 10)

		pdf.SetFillColor(245, 245, 245)
		pdf.CellFormat(55, 7, "Max LCeq (1s)", "", 0, "L", true, 0, "")
		pdf.SetFont("Helvetica", "B", 10)
		pdf.CellFormat(212, 7, fmt.Sprintf("%.1f dB(C) at %s", data.MaxLCeq, data.MaxLCeqTime.Format("15:04:05")), "", 1, "L", true, 0, "")
		pdf.SetFont("Helvetica", "", 10)
	}

	pdf.Ln(6)
	pdf.SetFont("Helvetica", "I", 8)
	pdf.SetTextColor(120, 120, 120)
	pdf.MultiCell(0, 4, "Note: LCeq,max is the maximum 1-second C-weighted equivalent continuous level per interval, computed from Z-weighted 1/3 octave band data. This is not a true peak (LCpk) measurement.", "", "L", false)
	pdf.SetTextColor(0, 0, 0)
}

func renderChartPage(pdf *fpdf.Fpdf, data *model.ReportData) {
	if len(data.IntervalResults) == 0 {
		return
	}

	chartPNG, err := RenderChart(data.IntervalResults, data.IncludeLAeq, data.IncludeLCeqMax)
	if err != nil {
		return
	}

	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 12)
	chartTitle := fmt.Sprintf("Time History - %s Intervals", formatDuration(data.IntervalDuration))
	pdf.CellFormat(0, 8, chartTitle, "", 1, "C", false, 0, "")
	pdf.Ln(4)

	reader := bytes.NewReader(chartPNG)
	opts := fpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}
	pdf.RegisterImageOptionsReader("chart", opts, reader)
	pdf.ImageOptions("chart", 10, pdf.GetY(), 277, 0, false, opts, 0, "")
}

func renderDataTable(pdf *fpdf.Fpdf, data *model.ReportData) {
	if len(data.IntervalResults) == 0 {
		return
	}

	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, 8, "Interval Data", "", 1, "C", false, 0, "")
	pdf.Ln(2)

	// Table header
	colWidths := []float64{35, 35, 35}
	headers := []string{"Start", "End", "Duration"}

	if data.IncludeLAeq {
		colWidths = append(colWidths, 40)
		headers = append(headers, "LAeq dB(A)")
	}
	if data.IncludeLCeqMax {
		colWidths = append(colWidths, 45, 35)
		headers = append(headers, "LCeq,max dB(C)", "Max at")
	}

	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetFillColor(50, 50, 50)
	pdf.SetTextColor(255, 255, 255)
	for i, h := range headers {
		pdf.CellFormat(colWidths[i], 7, h, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)
	pdf.SetTextColor(0, 0, 0)

	// Table rows
	pdf.SetFont("Helvetica", "", 9)

	// Find max LAeq for highlighting
	maxLAeq := math.Inf(-1)
	maxLCeq := math.Inf(-1)
	for _, r := range data.IntervalResults {
		if r.LAeq > maxLAeq {
			maxLAeq = r.LAeq
		}
		if r.LCeqMax > maxLCeq {
			maxLCeq = r.LCeqMax
		}
	}

	for i, r := range data.IntervalResults {
		if i%2 == 0 {
			pdf.SetFillColor(250, 250, 250)
		} else {
			pdf.SetFillColor(240, 240, 240)
		}

		dur := formatDuration(r.EndTime.Sub(r.StartTime))
		colIdx := 0

		pdf.CellFormat(colWidths[colIdx], 6, r.StartTime.Format("15:04:05"), "1", 0, "C", true, 0, "")
		colIdx++
		pdf.CellFormat(colWidths[colIdx], 6, r.EndTime.Format("15:04:05"), "1", 0, "C", true, 0, "")
		colIdx++
		pdf.CellFormat(colWidths[colIdx], 6, dur, "1", 0, "C", true, 0, "")
		colIdx++

		if data.IncludeLAeq {
			if r.LAeq == maxLAeq {
				pdf.SetFont("Helvetica", "B", 9)
				pdf.SetTextColor(220, 50, 50)
			}
			pdf.CellFormat(colWidths[colIdx], 6, fmt.Sprintf("%.1f", r.LAeq), "1", 0, "C", true, 0, "")
			pdf.SetFont("Helvetica", "", 9)
			pdf.SetTextColor(0, 0, 0)
			colIdx++
		}

		if data.IncludeLCeqMax {
			if r.LCeqMax == maxLCeq {
				pdf.SetFont("Helvetica", "B", 9)
				pdf.SetTextColor(220, 50, 50)
			}
			pdf.CellFormat(colWidths[colIdx], 6, fmt.Sprintf("%.1f", r.LCeqMax), "1", 0, "C", true, 0, "")
			pdf.SetFont("Helvetica", "", 9)
			pdf.SetTextColor(0, 0, 0)
			colIdx++
			pdf.CellFormat(colWidths[colIdx], 6, r.LCeqMaxAt.Format("15:04:05"), "1", 0, "C", true, 0, "")
			colIdx++
		}

		pdf.Ln(-1)
	}
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %02dm %02ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %02ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
