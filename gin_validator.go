package ginErrors

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/gin-gonic/gin/binding"
	"gopkg.in/go-playground/validator.v9"
)

type defaultValidator struct {
	once     sync.Once
	validate *validator.Validate
}

var (
	_ binding.StructValidator = &defaultValidator{}

	CommonValidationErrors = map[string]string{
		"ek":       "Ошибка валидации для свойства `%s` с правилом `%s`",
		"required": "Свойство `%s` обязательно для заполнения",
		"gt":       "Свойство `%s` должно содержать более `%s` элементов",
	}
)

func InitValidator() {
	binding.Validator = new(defaultValidator)
}

func (v *defaultValidator) ValidateStruct(obj interface{}) error {

	if kindOfData(obj) == reflect.Struct {

		v.lazyinit()

		if err := v.validate.Struct(obj); err != nil {
			return error(err)
		}
	}

	return nil
}

func (v *defaultValidator) Engine() interface{} {
	v.lazyinit()
	return v.validate
}

func (v *defaultValidator) lazyinit() {
	v.once.Do(func() {
		v.validate = validator.New()
		v.validate.SetTagName("binding")
		v.validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

			if name == "-" {
				return ""
			}

			return name
		})

		// add any custom validations etc. here
	})
}

func kindOfData(data interface{}) reflect.Kind {

	value := reflect.ValueOf(data)
	valueType := value.Kind()

	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}
	return valueType
}

// validationErrors Формирование массива ошибок
func MakeErrorsSlice(err error) map[string][]string {
	ve := map[string][]string{}
	for _, e := range err.(validator.ValidationErrors) {
		field := getFieldName(e.Namespace(), e.Field())
		if _, ok := ve[field]; !ok {
			ve[field] = []string{}
		}
		ve[field] = append(
			ve[field],
			getErrMessage(e.ActualTag(), field, e.Param()),
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

func getErrMessage(errorType string, field string, param string) string {
	errKey := errorType
	_, ok := CommonValidationErrors[errorType]
	if !ok {
		errKey = "ek"
	}

	if param != "" && errKey == "ek" {
		return fmt.Sprintf(CommonValidationErrors[errKey], field, errorType)
	}

	return fmt.Sprintf(CommonValidationErrors[errKey], field)
}
