package ginerrors

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"gopkg.in/go-playground/validator.v9"

	errs "github.com/microparts/errors-go"
)

type langName string
type validationRule string
type errorPattern string
type validationErrors map[validationRule]errorPattern

func (ve errorPattern) string() string {
	return string(ve)
}

var CommonValidationErrors = map[langName]validationErrors{
	"ru": {
		"ek":       "Ошибка валидации для свойства `%s` с правилом `%s`",
		"required": "Свойство `%s` обязательно для заполнения",
		"gt":       "Свойство `%s` должно содержать более `%s` элементов",
	},
}

var (
	defaultLang = "ru"

	ErrNotFound       = errors.New("route not found")
	ErrNoMethod       = errors.New("method not allowed")
	ErrServerError    = errors.New("internal server error")
	ErrRecordNotFound = errors.New("record not found")
)

func getLang(c *gin.Context) langName {
	lang := c.GetHeader("lang")
	if lang == "" {
		lang = c.DefaultQuery("lang", defaultLang)
	}

	return langName(lang)
}

// validationErrors Формирование массива ошибок
func makeErrorsSlice(err error, lang langName) map[errs.FieldName][]errs.ValidationError {
	ve := make(map[errs.FieldName][]errs.ValidationError)
	for _, e := range err.(validator.ValidationErrors) {
		field := getFieldName(e.Namespace(), e.Field())
		if _, ok := ve[field]; !ok {
			ve[field] = make([]errs.ValidationError, 0)
		}
		ve[field] = append(
			ve[field],
			getErrMessage(validationRule(e.ActualTag()), field, e.Param(), lang),
		)
	}
	return ve
}
func getFieldName(namespace string, field string) errs.FieldName {
	namespace = strings.Replace(namespace, "]", "", -1)
	namespace = strings.Replace(namespace, "[", ".", -1)
	namespaceSlice := strings.Split(namespace, ".")
	fieldName := field

	if len(namespaceSlice) > 2 {
		fieldName = strings.Join([]string{strings.Join(namespaceSlice[1:len(namespaceSlice)-1], "."), field}, ".")
	}

	return errs.FieldName(fieldName)
}

func getErrMessage(errorType validationRule, field errs.FieldName, param string, lang langName) errs.ValidationError {
	errKey := errorType
	_, ok := CommonValidationErrors[lang][errorType]
	if !ok {
		errKey = "ek"
	}

	if param != "" && errKey == "ek" {
		return errs.ValidationError(fmt.Sprintf(CommonValidationErrors[lang][errKey].string(), field, errorType))
	}

	return errs.ValidationError(fmt.Sprintf(CommonValidationErrors[lang][errKey].string(), field))
}
