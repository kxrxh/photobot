package fileutil

import (
	"fmt"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

var allowedMIMETypes = []string{
	"image/jpeg",
	"image/png",
	"image/gif",
	"image/webp",
	"image/bmp",
	"image/tiff",
	"image/heic",
	"image/heif",
	"video/mp4",
	"video/quicktime",
	"video/x-msvideo",
	"video/x-matroska",
	"video/webm",
	"video/mpeg",
	"video/3gpp",
	"video/3gpp2",
}

const MaxPrefixBytes = 8192

func ValidateFileContentFromPrefix(prefix []byte, filename string) (contentType string, err error) {
	if len(prefix) == 0 {
		return "", fmt.Errorf("file %s: empty prefix for MIME detection", filename)
	}
	mtype := mimetype.Detect(prefix)
	return checkMIMEAndExtension(mtype, filename)
}

func checkMIMEAndExtension(mtype *mimetype.MIME, filename string) (contentType string, err error) {
	isAllowed := false
	for _, allowed := range allowedMIMETypes {
		if mtype.Is(allowed) {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return "", fmt.Errorf("file %s has unsupported MIME type: %s", filename, mtype.String())
	}

	filenameLower := strings.ToLower(filename)
	if !extensionMatchesDetectedMIME(filenameLower, mtype) {
		return "", fmt.Errorf(
			"file extension doesn't match detected MIME type: %s != %s",
			filenameLower,
			mtype.Extension(),
		)
	}

	return mtype.String(), nil
}

func extensionMatchesDetectedMIME(filenameLower string, mtype *mimetype.MIME) bool {
	if ext := mtype.Extension(); ext != "" && strings.HasSuffix(filenameLower, ext) {
		return true
	}
	switch {
	case mtype.Is("image/jpeg"):
		return strings.HasSuffix(filenameLower, ".jpeg") ||
			strings.HasSuffix(filenameLower, ".jpg") ||
			strings.HasSuffix(filenameLower, ".jpe") ||
			strings.HasSuffix(filenameLower, ".jfif")
	case mtype.Is("image/tiff"):
		return strings.HasSuffix(filenameLower, ".tif") ||
			strings.HasSuffix(filenameLower, ".tiff")
	case mtype.Is("image/heic"), mtype.Is("image/heif"):
		return strings.HasSuffix(filenameLower, ".heic") ||
			strings.HasSuffix(filenameLower, ".heif")
	case mtype.Is("video/mp4"):
		return strings.HasSuffix(filenameLower, ".mp4") ||
			strings.HasSuffix(filenameLower, ".m4v")
	default:
		return false
	}
}
