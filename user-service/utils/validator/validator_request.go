package validator

import (
	"fmt"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

type Validator struct {
	Validator  *validator.Validate
	Translator ut.Translator
}

func NewValidator() *Validator {
	en := en.New()
	uni := ut.New(en, en)
	trans, found := uni.GetTranslator("en")
	if !found {
		panic("translator not found")
	}
	validate := validator.New()

	return &Validator{Validator: validate, Translator: trans}
}

func (v *Validator) Validate(i interface{}) error {
	err := v.Validator.Struct(i)
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return err
	}

	translatedErrors := make(map[string][]string)
	for _, e := range validationErrors {
		field := e.Field()
		msg := e.Translate(v.Translator)
		translatedErrors[field] = append(translatedErrors[field], msg)
	}

	return ValidationError{Errors: translatedErrors}
}

type ValidationError struct {
	Errors map[string][]string
}

func (v ValidationError) Error() string {
	return fmt.Sprintf("validation error: %+v", v.Errors)
}
