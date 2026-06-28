package multipartutil

import (
	"bytes"
	"mime/multipart"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultFormFileKeys(t *testing.T) {
	assert.Contains(t, DefaultFormFileKeys, "files")
	assert.Contains(t, DefaultFormFileKeys, "images")
	assert.Contains(t, DefaultFormFileKeys, "file")
	assert.Len(t, DefaultFormFileKeys, 3)
}

func TestCollectFileHeaders_NilForm(t *testing.T) {
	got := CollectFileHeaders(nil, DefaultFormFileKeys)
	assert.Nil(t, got)
}

func TestCollectFileHeaders_NilFileMap(t *testing.T) {
	form := &multipart.Form{File: nil}
	got := CollectFileHeaders(form, DefaultFormFileKeys)
	assert.Nil(t, got)
}

func TestCollectFileHeaders_EmptyForm(t *testing.T) {
	form := &multipart.Form{File: make(map[string][]*multipart.FileHeader)}
	got := CollectFileHeaders(form, DefaultFormFileKeys)
	assert.Empty(t, got)
}

func TestCollectFileHeaders_NilKeys(t *testing.T) {
	form := createTestForm(t, map[string][]string{"files": {"a.jpg"}})
	got := CollectFileHeaders(form, nil)
	assert.Empty(t, got)
}

func TestCollectFileHeaders_SingleKey(t *testing.T) {
	form := createTestForm(t, map[string][]string{"files": {"a.jpg", "b.png"}})
	got := CollectFileHeaders(form, []string{"files"})
	require.Len(t, got, 2)
	assert.Equal(t, "a.jpg", got[0].Filename)
	assert.Equal(t, "b.png", got[1].Filename)
}

func TestCollectFileHeaders_MultipleKeys(t *testing.T) {
	form := createTestForm(t, map[string][]string{
		"files":  {"f1.jpg"},
		"images": {"i1.png"},
		"file":   {"single.jpg"},
	})
	got := CollectFileHeaders(form, []string{"files", "images", "file"})
	require.Len(t, got, 3)
	assert.Equal(t, "f1.jpg", got[0].Filename)
	assert.Equal(t, "i1.png", got[1].Filename)
	assert.Equal(t, "single.jpg", got[2].Filename)
}

func TestCollectFileHeaders_KeyOrder(t *testing.T) {
	form := createTestForm(t, map[string][]string{
		"images": {"img.png"},
		"files":  {"f.jpg"},
	})
	got := CollectFileHeaders(form, []string{"files", "images"})
	require.Len(t, got, 2)
	assert.Equal(t, "f.jpg", got[0].Filename)
	assert.Equal(t, "img.png", got[1].Filename)
}

func TestCollectFileHeaders_SkipsMissingKeys(t *testing.T) {
	form := createTestForm(t, map[string][]string{"files": {"a.jpg"}})
	got := CollectFileHeaders(form, []string{"missing", "files", "also_missing"})
	require.Len(t, got, 1)
	assert.Equal(t, "a.jpg", got[0].Filename)
}

func TestFormFileKeys_NilForm(t *testing.T) {
	got := FormFileKeys(nil)
	assert.Nil(t, got)
}

func TestFormFileKeys_NilFileMap(t *testing.T) {
	form := &multipart.Form{File: nil}
	got := FormFileKeys(form)
	assert.Nil(t, got)
}

func TestFormFileKeys_EmptyForm(t *testing.T) {
	form := &multipart.Form{File: make(map[string][]*multipart.FileHeader)}
	got := FormFileKeys(form)
	assert.Empty(t, got)
}

func TestFormFileKeys_SingleKey(t *testing.T) {
	form := createTestForm(t, map[string][]string{"files": {"a.jpg"}})
	got := FormFileKeys(form)
	assert.ElementsMatch(t, []string{"files"}, got)
}

func TestFormFileKeys_MultipleKeys(t *testing.T) {
	form := createTestForm(t, map[string][]string{
		"files":  {"a.jpg"},
		"images": {"b.png"},
	})
	got := FormFileKeys(form)
	assert.ElementsMatch(t, []string{"files", "images"}, got)
}

func createTestForm(t *testing.T, files map[string][]string) *multipart.Form {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	for key, filenames := range files {
		for _, fn := range filenames {
			h := make(textproto.MIMEHeader)
			h.Set("Content-Disposition", `form-data; name="`+key+`"; filename="`+fn+`"`)
			h.Set("Content-Type", "application/octet-stream")
			part, err := w.CreatePart(h)
			require.NoError(t, err)
			_, err = part.Write([]byte("fake content"))
			require.NoError(t, err)
		}
	}
	boundary := w.Boundary()
	require.NoError(t, w.Close())

	r := multipart.NewReader(&buf, boundary)
	form, err := r.ReadForm(1 << 20)
	require.NoError(t, err)
	return form
}
