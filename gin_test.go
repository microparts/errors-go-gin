package ginerrors

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gopkg.in/go-playground/validator.v9"

	errs "github.com/microparts/errors-go"
)

func TestMakeResponse(t *testing.T) {
	type testCase struct {
		name      string
		err       interface{}
		httpCode  int
		errObject *errs.ErrorObject
	}
	errType := errs.ErrorTypeError

	cases := []testCase{
		{name: "common error", err: errors.New("common err"), httpCode: http.StatusBadRequest, errObject: &errs.ErrorObject{Message: "common err", Type: &errType}},

		{name: "validation error", err: makeValidationError(), httpCode: http.StatusUnprocessableEntity, errObject: &errs.ErrorObject{Message: "validation error", Validation: map[errs.FieldName][]errs.ValidationError{"String": {"Ошибка валидации для свойства `String` с правилом `%!s(MISSING)`"}}, Type: &errType}},

		{name: "mux err no method allowed", err: ErrNoMethod, httpCode: http.StatusMethodNotAllowed, errObject: &errs.ErrorObject{Message: ErrNoMethod.Error(), Type: &errType}},
		{name: "mux err route not found", err: ErrNotFound, httpCode: http.StatusNotFound, errObject: &errs.ErrorObject{Message: ErrNotFound.Error(), Type: &errType}},

		{name: "errors slice", err: []error{errors.New("common err 1"), errors.New("common err 2")}, httpCode: http.StatusInternalServerError, errObject: &errs.ErrorObject{Message: "common err 1; common err 2", Type: &errType}},
		{name: "map of errors", err: map[string]error{"common_err": errors.New("common err")}, httpCode: http.StatusBadRequest, errObject: &errs.ErrorObject{Message: map[string]string{"common_err": "common err"}, Type: &errType}},

		{name: "record not found", err: ErrRecordNotFound, httpCode: http.StatusNotFound, errObject: &errs.ErrorObject{Message: ErrRecordNotFound.Error(), Type: &errType}},
		{name: "sql error no rows", err: sql.ErrNoRows, httpCode: http.StatusNotFound, errObject: &errs.ErrorObject{Message: ErrRecordNotFound.Error(), Type: &errType}},
		{name: "sql error conn done", err: sql.ErrConnDone, httpCode: http.StatusInternalServerError, errObject: &errs.ErrorObject{Message: sql.ErrConnDone.Error(), Type: &errType}},
		{name: "sql error tx done", err: sql.ErrTxDone, httpCode: http.StatusInternalServerError, errObject: &errs.ErrorObject{Message: sql.ErrTxDone.Error(), Type: &errType}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			errCode, errObject := MakeResponse(testCase.err, "ru")
			assert.Equal(t, testCase.errObject, errObject, testCase.name)
			assert.Equal(t, testCase.httpCode, errCode, testCase.name)
		})
	}
}

func setupRouter() *gin.Engine {
	r := gin.New()

	r.NoRoute(func(c *gin.Context) { Response(c, ErrNotFound) })
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
	return r
}

func TestResponse(t *testing.T) {
	router := setupRouter()

	t.Run("not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/pong", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, "{\"error\":{\"message\":\"route not found\",\"type\":\"error\"}}\n", w.Body.String())
	})
}

func makeValidationError() error {
	// MyStruct ..
	type MyStruct struct {
		String string `validate:"is-awesome"`
	}

	// use a single instance of Validate, it caches struct info
	validate := validator.New()
	_ = validate.RegisterValidation("is-awesome", ValidateMyVal)

	s := MyStruct{String: "awesome"}

	err := validate.Struct(s)
	if err != nil {
		fmt.Printf("Err(s):\n%+v\n", err)
	}

	s.String = "not awesome"
	return validate.Struct(s)
}

// ValidateMyVal implements validator.Func
func ValidateMyVal(fl validator.FieldLevel) bool {
	return fl.Field().String() == "awesome"
}
