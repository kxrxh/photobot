package render

import (
	"bytes"
	"errors"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"csort.ru/reports-service/internal/view"
)

var (
	reportMu   sync.RWMutex
	reportTmpl *template.Template
)

func LoadReportTemplate(templatePath string) error {
	prep, err := PrepareTemplateForPDF(templatePath)
	if err != nil {
		return err
	}
	tmpl, err := template.New("main.html").Option("missingkey=zero").Parse(prep)
	if err != nil {
		return err
	}
	bodyPath := filepath.Join(filepath.Dir(templatePath), "body.html")
	tmpl, err = tmpl.ParseFiles(bodyPath)
	if err != nil {
		return err
	}
	reportMu.Lock()
	reportTmpl = tmpl
	reportMu.Unlock()
	return nil
}

type ReportHTMLData struct {
	BodyHTML template.HTML
}

func RenderReportHTML(data any) (string, error) {
	reportMu.RLock()
	tmpl := reportTmpl
	reportMu.RUnlock()
	if tmpl == nil {
		return "", errors.New("report template not loaded; call LoadReportTemplate first")
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

const bodyTemplateName = "body"

func RenderBody(data view.Body) (template.HTML, error) {
	reportMu.RLock()
	tmpl := reportTmpl
	reportMu.RUnlock()
	if tmpl == nil {
		return "", errors.New("report template not loaded; call LoadReportTemplate first")
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, bodyTemplateName, data); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}

func LoadTemplateFallback(templatePath string) string {
	raw, err := os.ReadFile(templatePath)
	if err != nil {
		return "<html><body><h1>Report</h1></body></html>"
	}
	return strings.TrimSpace(string(raw))
}
