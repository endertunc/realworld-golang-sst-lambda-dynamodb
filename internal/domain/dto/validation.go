package dto

import (
	"errors"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	v10 "github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/non-standard/validators"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"log"
	"strings"
	"sync"
)

// Validatable is an interface that matches any type with a Validate method
// that returns a map[string]string and bool
type Validatable interface {
	Validate() (map[string]string, bool)
}

var translator = sync.OnceValue(func() ut.Translator {
	en := en.New()
	uni := ut.New(en, en)

	translator, ok := uni.GetTranslator("en")
	if !ok {
		log.Fatalf("error getting translator")
	}
	return translator
})

var validator = sync.OnceValue(func() *v10.Validate {
	validator := v10.New(v10.WithRequiredStructEnabled())
	t := translator()

	// register custom notblank validator
	notBlankTag := "notblank"
	err := validator.RegisterValidation(notBlankTag, validators.NotBlank)
	if err != nil {
		log.Fatalf("error registering notblank validator: %v", err)
	}

	// register default translations
	err = en_translations.RegisterDefaultTranslations(validator, t)
	if err != nil {
		log.Fatalf("error registering translations: %v", err)
	}

	// register custom translation for notblank
	err = validator.RegisterTranslation(notBlankTag, t, func(ut ut.Translator) error {
		return ut.Add(notBlankTag, "{0} cannot be blank", true)
	}, func(ut ut.Translator, fe v10.FieldError) string {
		t, err := ut.T(notBlankTag, fe.Field())
		if err != nil {
			log.Printf("warning: error translating FieldError: %#v", fe)
			return fe.(error).Error()
		}
		return t
	})
	if err != nil {
		log.Fatalf("error registering translation for notblank: %v", err)
	}

	return validator
})

func validateStruct[T Validatable](value T) (map[string]string, bool) {
	err := validator().Struct(value)
	if err != nil {
		trans := make(map[string]string)
		var errs v10.ValidationErrors
		if errors.As(err, &errs) {
			for _, fieldError := range errs {
				fieldStructPath := dropRootStructName(fieldError.Namespace())
				trans[fieldStructPath] = fieldError.Translate(translator())
			}
			return trans, true
		}

	}
	return nil, false
}

func dropRootStructName(s string) string {
	indexOfPeriod := strings.IndexByte(s, '.')
	if indexOfPeriod == -1 {
		// no periods found, return the original string
		return s
	}
	// return the substring starting from the index after the first period
	return s[indexOfPeriod+1:]
}
