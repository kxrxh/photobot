package render

import (
	"os"
	"regexp"
	"strings"
)

const (
	qrCaliberSVGPlaceholder = `<span id="report-qr-caliber-placeholder"></span>`
)

var (
	svgXMLDeclRe = regexp.MustCompile(`<\?xml[^?]*\?>`)
	svgDoctypeRe = regexp.MustCompile(`(?i)<!DOCTYPE[^>]*>`)
)

func stripSVGPreamble(raw string) string {
	s := svgXMLDeclRe.ReplaceAllString(raw, "")
	s = svgDoctypeRe.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

func rewriteQRSVGRasterRef(s string) string {
	s = strings.ReplaceAll(s, `xlink:href="`+qrRasterName+`"`, `xlink:href="`+qrRasterWebPath+`"`)
	s = strings.ReplaceAll(s, `xlink:href='`+qrRasterName+`'`, `xlink:href='`+qrRasterWebPath+`'`)
	return s
}

func InjectQRSVG(html string, templateDir string) string {
	if !strings.Contains(html, qrCaliberSVGPlaceholder) {
		return html
	}
	qrPath := qrSVGPath(templateDir)
	raw, err := os.ReadFile(qrPath)
	repl := ""
	if err == nil {
		repl = rewriteQRSVGRasterRef(stripSVGPreamble(string(raw)))
	}
	return strings.ReplaceAll(html, qrCaliberSVGPlaceholder, repl)
}

func PrepareTemplateForPDF(templatePath string) (string, error) {
	raw, err := os.ReadFile(templatePath)
	if err != nil {
		return "", err
	}
	cssPath := reportCSSPath(templatePath)
	css, err := os.ReadFile(cssPath)
	if err != nil {
		return "", err
	}
	s := string(raw)
	linkRe := regexp.MustCompile(
		`(?i)<link\s+rel=['"]stylesheet['"]\s+href=['"]\.\./assets/css/report\.css['"]\s*/?>`,
	)
	s = linkRe.ReplaceAllString(s, "<style>"+string(css)+"</style>")
	return s, nil
}
