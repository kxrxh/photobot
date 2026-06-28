package render

import "path/filepath"

const (
	relHTML         = "html"
	relCSS          = "assets/css"
	relImages       = "assets/images"
	relQR           = "assets/qr"
	mainName        = "main.html"
	bodyName        = "body.html"
	cssName         = "report.css"
	qrFileName      = "qrcode_caliber.svg"
	qrRasterName    = "qrcode_caliber_raster.png"
	qrRasterWebPath = relQR + "/" + qrRasterName

	LogoRelPath = "assets/images/logo.png"
)

func MainHTMLPath(root string) string {
	return filepath.Join(root, relHTML, mainName)
}

func TemplateRootFromMainPath(mainHTMLPath string) string {
	return filepath.Clean(filepath.Join(filepath.Dir(mainHTMLPath), ".."))
}

func reportCSSPath(mainHTMLPath string) string {
	return filepath.Join(TemplateRootFromMainPath(mainHTMLPath), relCSS, cssName)
}

func qrSVGPath(templateRoot string) string {
	return filepath.Join(templateRoot, relQR, qrFileName)
}
