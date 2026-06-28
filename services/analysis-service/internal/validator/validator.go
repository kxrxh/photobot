package validator

import (
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
)

var (
	Default *validator.Validate
	Trans   ut.Translator
)

func init() {
	Default = validator.New()
	locale := en.New()
	uni := ut.New(locale, locale)
	Trans, _ = uni.GetTranslator("en")
	_ = enTranslations.RegisterDefaultTranslations(Default, Trans)
}

func Translate(err validator.FieldError) string {
	return err.Translate(Trans)
}

func GetJSONFieldName(fieldName string, structType reflect.Type) string {
	if structType.Kind() == reflect.Pointer {
		structType = structType.Elem()
	}

	field, found := structType.FieldByName(fieldName)
	if !found {
		return fieldName
	}

	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		return fieldName
	}

	jsonName := strings.Split(jsonTag, ",")[0]
	if jsonName == "-" {
		return fieldName
	}

	return jsonName
}
