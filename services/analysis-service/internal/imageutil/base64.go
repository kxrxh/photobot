package imageutil

import (
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var ErrIndexOutOfRange = errors.New("image index out of range")

var dataURLRegex = regexp.MustCompile(`^data:([^;]+);base64,(.+)$`)

func ParseImageArray(imageString string) []string {
	if imageString == "" {
		return []string{}
	}

	if strings.HasPrefix(imageString, "data:") {
		if _, data, found := strings.Cut(imageString, ","); found {
			return []string{data}
		}
	}

	if strings.HasPrefix(imageString, "['") && strings.HasSuffix(imageString, "']") {
		arrayContent := imageString[2 : len(imageString)-2]
		const sep = "', '"

		count := strings.Count(arrayContent, sep) + 1
		result := make([]string, 0, count)

		remainder := arrayContent
		for {
			if before, after, found := strings.Cut(remainder, sep); found {
				if trimmed := strings.TrimSpace(before); trimmed != "" {
					result = append(result, trimmed)
				}
				remainder = after
			} else {
				if trimmed := strings.TrimSpace(remainder); trimmed != "" {
					result = append(result, trimmed)
				}
				break
			}
		}
		return result
	}

	return []string{imageString}
}

func DecodeBase64Image(base64String *string) (*ImageData, error) {
	if base64String == nil {
		return nil, errors.New("base64 string is nil")
	}
	if *base64String == "" {
		return nil, errors.New("base64 string is empty")
	}

	var mimeType string
	var base64Data string

	if strings.HasPrefix(*base64String, "data:") {
		matches := dataURLRegex.FindStringSubmatch(*base64String)
		if len(matches) == 3 {
			mimeType = matches[1]
			base64Data = matches[2]
		} else {
			return nil, errors.New("invalid data URL format")
		}
	} else {
		base64Data = strings.ReplaceAll(*base64String, " ", "")
		mimeType = getMimeTypeFromBase64(base64Data)
	}

	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	return &ImageData{
		Data:     data,
		MimeType: mimeType,
	}, nil
}

func getMimeTypeFromBase64(base64String string) string {
	clean := strings.ReplaceAll(base64String, " ", "")

	if strings.HasPrefix(clean, "iVBOR") {
		return "image/png"
	}
	if strings.HasPrefix(clean, "/9j/") {
		return "image/jpeg"
	}
	if strings.HasPrefix(clean, "R0lGOD") || strings.HasPrefix(clean, "R0lGO") {
		return "image/gif"
	}
	if strings.HasPrefix(clean, "UklGR") {
		return "image/webp"
	}
	if isHEICFromBase64(clean) {
		return "image/heic"
	}

	return "image/jpeg"
}

func isHEICFromBytes(data []byte) bool {
	if len(data) < 12 {
		return false
	}
	if string(data[4:8]) != "ftyp" {
		return false
	}
	brand := string(data[8:12])
	switch brand {
	case "heic", "heix", "hevc", "hevx", "mif1", "msf1":
		return true
	}
	return false
}

func GetMimeTypeFromBytes(data []byte) string {
	if len(data) < 12 {
		return "image/jpeg"
	}
	if len(data) >= 8 && string(data[0:8]) == "\x89PNG\r\n\x1a\n" {
		return "image/png"
	}
	if len(data) >= 3 && string(data[0:3]) == "\xff\xd8\xff" {
		return "image/jpeg"
	}
	if len(data) >= 6 && (string(data[0:6]) == "GIF87a" || string(data[0:6]) == "GIF89a") {
		return "image/gif"
	}
	if len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP" {
		return "image/webp"
	}
	if isHEICFromBytes(data) {
		return "image/heic"
	}
	return "image/jpeg"
}

func isHEICFromBase64(b64 string) bool {
	const minLen = 16
	if len(b64) < minLen {
		return false
	}
	data, err := base64.StdEncoding.DecodeString(b64[:minLen])
	if err != nil {
		return false
	}
	return isHEICFromBytes(data)
}
