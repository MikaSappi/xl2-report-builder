# Sound level report generator for NTi XL2

This program is designed to convert the raw .txt data from an XL2 SL meter to a user-friendly pdf file.

## How to use
#### Executable names
* macOS: `xl2-report-maker-macos`
* Linux: `xl2-report-maker-linux`
* Windows: `xl2-report-maker-windows.exe`

### Fast report
`/path/to/xl2-report-builder /path/to/report.txt`

You will get a pdf with `LAeq5 min` and `LCeq 1s` reports.

### Advanced report

`xl2-report-builder [flags] <input-file>`

#### Flags:
```
  --laeq         Include LAeq time history in report (default: both enabled if none specified)
  --lceqmax      Include LCeq max time history in report (default: both enabled if none specified)
  --interval     Aggregation interval, e.g. 5m, 10m, 1m (default: 5m)
  --output       Output PDF file path (default: input_report.pdf)
  --company      Company name to display in report header and footer
  --logo         Path to company logo (PNG or JPEG) for report header
```

## Contact

For any inquiries, please [contact us](https://fremen.fi/contact).
