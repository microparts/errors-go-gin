package ginerrors

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gopkg.in/go-playground/validator.v9"
)

type langName string
type validationRule string
type errorPattern string
type validationError map[validationRule]errorPattern

func (ve errorPattern) string() string {
	return string(ve)
}

var CommonValidationErrors = map[langName]validationError{
	"ru": {
		"ek":       "Ошибка валидации для свойства `%s` с правилом `%s`",
		"required": "Свойство `%s` обязательно для заполнения",
		"gt":       "Свойство `%s` должно содержать более `%s` элементов",
	},
}

type ResponseObject struct {
	Error ErrorObject `json:"error,omitempty"`
}

type ErrorObject struct {
	Message    interface{}         `json:"message"`
	Code       int                 `json:"code,omitempty"`
	Validation map[string][]string `json:"validation,omitempty"`
	Debug      string              `json:"debug,omitempty"`
}

var (
	defaultLang = "ru"

	ErrNotFound       = errors.New("route not found")
	ErrNoMethod       = errors.New("method not allowed")
	ErrServerError    = errors.New("internal server error")
	ErrRecordNotFound = errors.New("record not found")
)

//Response makes common error response
func Response(c *gin.Context, err interface{}) {
	errCode, data := MakeResponse(err, getLang(c))
	resp := ResponseObject{Error: *data}
	c.AbortWithStatusJSON(errCode, resp)
}

func getLang(c *gin.Context) langName {
	lang := c.GetHeader("lang")
	if lang == "" {
		lang = c.DefaultQuery("lang", defaultLang)
	}

	return langName(lang)
}

//MakeResponse makes ErrorObject based on error type
func MakeResponse(err interface{}, lang langName) (int, *ErrorObject) {
	errObj := &ErrorObject{}
	errCode := http.StatusBadRequest

	switch et := err.(type) {
	case []error:
		errCode = http.StatusInternalServerError
		msgs := make([]string, 0)
		for _, e := range err.([]error) {
			msgs = append(msgs, e.Error())
		}
		errObj.Message = strings.Join(msgs, "; ")

	case validator.ValidationErrors:
		errCode = http.StatusUnprocessableEntity

		errObj.Message = "validation error"
		errObj.Validation = MakeErrorsSlice(et, lang)

	case error:
		errCode, errObj.Message = getErrCode(et)

	case map[string]error:
		msgs := make(map[string]string)
		for k, e := range et {
			msgs[k] = e.Error()
		}

		errObj.Message = msgs
	}

	return errCode, errObj
}

func getErrCode(et error) (errCode int, msg string) {
	msg = et.Error()
	switch msg {
	case ErrNotFound.Error():
		errCode = http.StatusNotFound
	case ErrNoMethod.Error():
		errCode = http.StatusMethodNotAllowed
	case ErrServerError.Error(), sql.ErrConnDone.Error(), sql.ErrTxDone.Error():
		errCode = http.StatusInternalServerError
	case ErrRecordNotFound.Error():
		errCode = http.StatusNotFound
	case sql.ErrNoRows.Error():
		errCode = http.StatusNotFound
		msg = ErrRecordNotFound.Error()
	default:
		errCode = http.StatusBadRequest
	}

	return
}

// validationErrors Формирование массива ошибок
func MakeErrorsSlice(err error, lang langName) map[string][]string {
	ve := map[string][]string{}
	for _, e := range err.(validator.ValidationErrors) {
		field := getFieldName(e.Namespace(), e.Field())
		if _, ok := ve[field]; !ok {
			ve[field] = []string{}
		}
		ve[field] = append(
			ve[field],
			getErrMessage(validationRule(e.ActualTag()), field, e.Param(), lang),
		)
	}
	return ve
}
func getFieldName(namespace string, field string) string {
	namespace = strings.Replace(namespace, "]", "", -1)
	namespace = strings.Replace(namespace, "[", ".", -1)
	namespaceSlice := strings.Split(namespace, ".")
	fieldName := field

	if len(namespaceSlice) > 2 {
		fieldName = strings.Join([]string{strings.Join(namespaceSlice[1:len(namespaceSlice)-1], "."), field}, ".")
	}

	return fieldName
}

func getErrMessage(errorType validationRule, field string, param string, lang langName) string {
	errKey := errorType
	_, ok := CommonValidationErrors[lang][errorType]
	if !ok {
		errKey = "ek"
	}

	if param != "" && errKey == "ek" {
		return fmt.Sprintf(CommonValidationErrors[lang][errKey].string(), field, errorType)
	}

	return fmt.Sprintf(CommonValidationErrors[lang][errKey].string(), field)
}
