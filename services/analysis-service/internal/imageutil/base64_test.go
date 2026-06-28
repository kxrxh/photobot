package imageutil

import (
	"encoding/base64"
	"reflect"
	"testing"
)

func TestParseImageArray(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  []string
	}{
		{"Empty", "", []string{}},
		{"SingleBase64", "SGVsbG8=", []string{"SGVsbG8="}},
		{"DataURL", "data:image/png;base64,SGVsbG8=", []string{"SGVsbG8="}},
		{"ArraySingle", "['SGVsbG8=']", []string{"SGVsbG8="}},
		{"ArrayMultiple", "['SGVsbG8=', 'V29ybGQ=']", []string{"SGVsbG8=", "V29ybGQ="}},
		{"ArrayWithSpaces", "[' SGVsbG8= ', ' V29ybGQ= ']", []string{"SGVsbG8=", "V29ybGQ="}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseImageArray(tc.input)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ParseImageArray(%s) = %v; want %v", tc.name, got, tc.want)
			}
		})
	}
}

var heicMagic = []byte{0x00, 0x00, 0x00, 0x20, 'f', 't', 'y', 'p', 'h', 'e', 'i', 'c'}

func TestDecodeBase64Image_HEIC(t *testing.T) {
	b64Full := base64.StdEncoding.EncodeToString(append(heicMagic, make([]byte, 100)...))
	img, err := DecodeBase64Image(&b64Full)
	if err != nil {
		t.Fatalf("DecodeBase64Image(HEIC): %v", err)
	}
	if img.MimeType != "image/heic" {
		t.Errorf("DecodeBase64Image(HEIC) MimeType = %q; want image/heic", img.MimeType)
	}
}

func BenchmarkParseImageArray_Single(b *testing.B) {
	input := "SGVsbG8="
	b.ResetTimer()
	for range b.N {
		ParseImageArray(input)
	}
}

func BenchmarkParseImageArray_DataURL(b *testing.B) {
	input := "data:image/png;base64,SGVsbG8="
	b.ResetTimer()
	for range b.N {
		ParseImageArray(input)
	}
}

func BenchmarkParseImageArray_ArrayMultiple(b *testing.B) {
	input := "['SGVsbG8=', 'V29ybGQ=', 'U29tZXRoaW5n', 'RWxzZQ==']"
	b.ResetTimer()
	for range b.N {
		ParseImageArray(input)
	}
}
