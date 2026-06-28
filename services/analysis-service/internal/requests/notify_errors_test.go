package requests

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestNotifyErrors_UnmarshalJSON_stringSlice(t *testing.T) {
	const raw = `{"errors":["analisis error","line two"]}`
	var body struct {
		E NotifyErrors `json:"errors"`
	}
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		t.Fatal(err)
	}
	want := NotifyErrors{
		{"": "analisis error"},
		{"": "line two"},
	}
	if !reflect.DeepEqual(body.E, want) {
		t.Fatalf("got %#v want %#v", body.E, want)
	}
}

func TestNotifyErrors_UnmarshalJSON_maps(t *testing.T) {
	const raw = `{"errors":[{"a.jpg":"bad"}]}`
	var body struct {
		E NotifyErrors `json:"errors"`
	}
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		t.Fatal(err)
	}
	want := NotifyErrors{{"a.jpg": "bad"}}
	if !reflect.DeepEqual(body.E, want) {
		t.Fatalf("got %#v want %#v", body.E, want)
	}
}
