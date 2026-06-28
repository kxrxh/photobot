package validation

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetJSONFieldName(t *testing.T) {
	type TestStruct struct {
		UserID   int    `json:"user_id"`
		Email    string `json:"email"`
		NoTag    string
		OmitJSON string `json:"-"`
		WithOpt  string `json:"with_opt,omitempty"`
	}

	t.Run("field with json tag", func(t *testing.T) {
		var s TestStruct
		got := GetJSONFieldName("UserID", reflect.TypeOf(s))
		assert.Equal(t, "user_id", got)
	})

	t.Run("field email", func(t *testing.T) {
		var s TestStruct
		got := GetJSONFieldName("Email", reflect.TypeOf(s))
		assert.Equal(t, "email", got)
	})

	t.Run("field without json tag", func(t *testing.T) {
		var s TestStruct
		got := GetJSONFieldName("NoTag", reflect.TypeOf(s))
		assert.Equal(t, "NoTag", got)
	})

	t.Run("field with json omitempty", func(t *testing.T) {
		var s TestStruct
		got := GetJSONFieldName("WithOpt", reflect.TypeOf(s))
		assert.Equal(t, "with_opt", got)
	})

	t.Run("field with json -", func(t *testing.T) {
		var s TestStruct
		got := GetJSONFieldName("OmitJSON", reflect.TypeOf(s))
		assert.Equal(t, "OmitJSON", got)
	})

	t.Run("field not found", func(t *testing.T) {
		var s TestStruct
		got := GetJSONFieldName("NonExistent", reflect.TypeOf(s))
		assert.Equal(t, "NonExistent", got)
	})

	t.Run("pointer to struct", func(t *testing.T) {
		var s *TestStruct
		got := GetJSONFieldName("UserID", reflect.TypeOf(s))
		assert.Equal(t, "user_id", got)
	})
}
