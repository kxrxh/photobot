package fileutil

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	jpegMagic = []byte{0xFF, 0xD8, 0xFF}
	pngMagic  = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
)

func TestMaxPrefixBytes(t *testing.T) {
	assert.Equal(t, 8192, MaxPrefixBytes)
}

func TestValidateFileContentFromPrefix_EmptyPrefix(t *testing.T) {
	contentType, err := ValidateFileContentFromPrefix(nil, "test.jpg")
	assert.Error(t, err)
	assert.Empty(t, contentType)
	assert.Contains(t, err.Error(), "empty prefix")

	contentType, err = ValidateFileContentFromPrefix([]byte{}, "test.jpg")
	assert.Error(t, err)
	assert.Empty(t, contentType)
	assert.Contains(t, err.Error(), "empty prefix")
}

func TestValidateFileContentFromPrefix_JPEG(t *testing.T) {
	prefix := append(slices.Clone(jpegMagic), make([]byte, 100)...)
	contentType, err := ValidateFileContentFromPrefix(prefix, "photo.jpg")
	require.NoError(t, err)
	assert.Equal(t, "image/jpeg", contentType)
}

func TestValidateFileContentFromPrefix_JPEG_ExtensionCaseInsensitive(t *testing.T) {
	prefix := append(slices.Clone(jpegMagic), make([]byte, 100)...)
	contentType, err := ValidateFileContentFromPrefix(prefix, "photo.JPG")
	require.NoError(t, err)
	assert.Equal(t, "image/jpeg", contentType)
}

func TestValidateFileContentFromPrefix_JPEG_DotJpegExtension(t *testing.T) {
	prefix := append(slices.Clone(jpegMagic), make([]byte, 100)...)
	contentType, err := ValidateFileContentFromPrefix(prefix, "IMG_4847.jpeg")
	require.NoError(t, err)
	assert.Equal(t, "image/jpeg", contentType)
}

func TestValidateFileContentFromPrefix_PNG(t *testing.T) {
	prefix := append(slices.Clone(pngMagic), make([]byte, 100)...)
	contentType, err := ValidateFileContentFromPrefix(prefix, "image.png")
	require.NoError(t, err)
	assert.Equal(t, "image/png", contentType)
}

func TestValidateFileContentFromPrefix_ExtensionMismatch(t *testing.T) {
	prefix := append(slices.Clone(jpegMagic), make([]byte, 100)...)
	contentType, err := ValidateFileContentFromPrefix(prefix, "photo.png")
	assert.Error(t, err)
	assert.Empty(t, contentType)
	assert.Contains(t, err.Error(), "extension doesn't match")
}

func TestValidateFileContentFromPrefix_UnsupportedMIME(t *testing.T) {
	pdfMagic := []byte{0x25, 0x50, 0x44, 0x46}
	prefix := append(slices.Clone(pdfMagic), make([]byte, 100)...)
	contentType, err := ValidateFileContentFromPrefix(prefix, "doc.pdf")
	assert.Error(t, err)
	assert.Empty(t, contentType)
	assert.Contains(t, err.Error(), "unsupported MIME type")
}

func TestValidateFileContentFromPrefix_FilenameInError(t *testing.T) {
	prefix := append(slices.Clone(jpegMagic), make([]byte, 100)...)
	_, err := ValidateFileContentFromPrefix(prefix, "photo.png")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "photo.png")
}
