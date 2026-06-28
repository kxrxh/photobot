package validator

import (
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

// Translate returns a user-friendly message for a validation field error.
func Translate(err validator.FieldError) string {
	return err.Translate(Trans)
}
