package storage

import (
	"errors"
	"strings"
)

const (
	reportFmtCSV = "csv"
	reportFmtPDF = "pdf"

	MetaReportOwnerUserID = "report-owner-user-id"
)

func NormalizeReportFormat(s string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case reportFmtCSV:
		return reportFmtCSV, nil
	case reportFmtPDF:
		return reportFmtPDF, nil
	default:
		return "", errors.New("report format must be csv or pdf")
	}
}

func ReportObjectKey(analysisID, format string) (string, error) {
	ft, err := NormalizeReportFormat(format)
	if err != nil {
		return "", err
	}
	return analysisID + "." + ft, nil
}

func reportContentType(format string) (string, error) {
	ft, err := NormalizeReportFormat(format)
	if err != nil {
		return "", err
	}
	switch ft {
	case reportFmtCSV:
		return "text/csv", nil
	case reportFmtPDF:
		return "application/pdf", nil
	default:
		return "", errors.New("unsupported format")
	}
}

func ReportAttachmentFilename(analysisID, format string) (string, error) {
	ft, err := NormalizeReportFormat(format)
	if err != nil {
		return "", err
	}
	switch ft {
	case reportFmtCSV:
		return analysisID + "_report.csv", nil
	case reportFmtPDF:
		return analysisID + "_report.pdf", nil
	default:
		return "", errors.New("unsupported format")
	}
}
