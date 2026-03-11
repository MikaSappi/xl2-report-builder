package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"xl2-report-builder/internal/acoustics"
	"xl2-report-builder/internal/model"
	"xl2-report-builder/internal/parser"
	"xl2-report-builder/internal/report"
)

func main() {
	laeq := flag.Bool("laeq", false, "Include LAeq time history in report")
	lceqmax := flag.Bool("lceqmax", false, "Include LCeq max time history in report")
	interval := flag.Duration("interval", 5*time.Minute, "Aggregation interval (e.g. 5m, 10m, 1m)")
	output := flag.String("output", "", "Output PDF file path (default: derived from input)")
	company := flag.String("company", "", "Company name to display in report header and footer")
	logo := flag.String("logo", "", "Path to company logo (PNG or JPEG) for report header")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: xl2-report-builder [flags] <input-file>\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	inputPath := flag.Arg(0)

	// If no flags specified, enable both by default
	if !*laeq && !*lceqmax {
		*laeq = true
		*lceqmax = true
	}

	outputPath := *output
	if outputPath == "" {
		base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
		outputPath = base + "_report.pdf"
	}

	fmt.Printf("Parsing %s...\n", inputPath)
	measurement, err := parser.ParseLogFile(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Parsed %d samples (%s to %s)\n",
		len(measurement.Samples),
		measurement.Header.StartTime.Format("15:04:05"),
		measurement.Header.EndTime.Format("15:04:05"),
	)

	fmt.Printf("Computing metrics with %s intervals...\n", *interval)
	results := acoustics.ComputeIntervalResults(measurement.Samples, *interval)
	overallLAeq := acoustics.OverallLAeq(measurement.Samples)
	overallLCeq := acoustics.OverallLCeq(measurement.Samples)
	maxLC, maxLCTime := acoustics.MaxLCeq(measurement.Samples)

	reportData := &model.ReportData{
		Measurement:      measurement,
		IntervalResults:  results,
		IntervalDuration: *interval,
		OverallLAeq:      overallLAeq,
		OverallLCeq:      overallLCeq,
		MaxLCeq:          maxLC,
		MaxLCeqTime:      maxLCTime,
		IncludeLAeq:      *laeq,
		IncludeLCeqMax:   *lceqmax,
		CompanyName:      *company,
		LogoPath:         *logo,
	}

	fmt.Printf("Generating PDF report: %s\n", outputPath)
	if err := report.GeneratePDF(reportData, outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating PDF: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Done!")
	fmt.Printf("  Overall LAeq:   %.1f dB(A)\n", overallLAeq)
	fmt.Printf("  Overall LCeq:   %.1f dB(C)\n", overallLCeq)
	fmt.Printf("  Max LCeq (1s):  %.1f dB(C) at %s\n", maxLC, maxLCTime.Format("15:04:05"))
	fmt.Printf("  Intervals:      %d\n", len(results))
}
