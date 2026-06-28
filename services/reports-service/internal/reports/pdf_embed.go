package reports

import _ "embed"

// PDFWarmupHTML is minimal HTML used to warm up headless Chrome at API startup.
//
//go:embed pdfassets/warmup.html
var PDFWarmupHTML string

// printReadyJS is evaluated in the report page before PrintToPDF.
//
//go:embed pdfassets/print_ready.js
var printReadyJS string
