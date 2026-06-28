package multipartutil

import "mime/multipart"

var DefaultFormFileKeys = []string{"files", "images", "file"}

func CollectFileHeaders(form *multipart.Form, keys []string) []*multipart.FileHeader {
	if form == nil || form.File == nil {
		return nil
	}
	var out []*multipart.FileHeader
	for _, key := range keys {
		if list := form.File[key]; len(list) > 0 {
			out = append(out, list...)
		}
	}
	return out
}

func FormFileKeys(form *multipart.Form) []string {
	if form == nil || form.File == nil {
		return nil
	}
	keys := make([]string, 0, len(form.File))
	for k := range form.File {
		keys = append(keys, k)
	}
	return keys
}
