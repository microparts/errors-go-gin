package ginerrors

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/go-playground/validator.v9"
)

func TestMakeResponse(t *testing.T) {
	type testCase struct {
		name      string
		err       interface{}
		isErr     bool
		httpCode  int
		errObject *ErrorObject
	}

	cases := []testCase{
		{name: "common error", err: errors.New("common err"), isErr: true, httpCode: http.StatusBadRequest, errObject: &ErrorObject{Message: "common err"}},

		{name: "validation error", err: makeValidationError(), isErr: true, httpCode: http.StatusUnprocessableEntity, errObject: &ErrorObject{Message: "validation error", Validation: map[string][]string{"String": {"Ошибка валидации для свойства `String` с правилом `%!s(MISSING)`"}}, Debug: ""}},

		{name: "mux err no method allowed", err: ErrNoMethod, isErr: true, httpCode: http.StatusMethodNotAllowed, errObject: &ErrorObject{Message: ErrNoMethod.Error()}},
		{name: "mux err route not found", err: ErrNotFound, isErr: true, httpCode: http.StatusNotFound, errObject: &ErrorObject{Message: ErrNotFound.Error()}},

		{name: "errors slice", err: []error{errors.New("common err 1"), errors.New("common err 2")}, isErr: true, httpCode: http.StatusInternalServerError, errObject: &ErrorObject{Message: "common err 1; common err 2"}},
		{name: "map of errors", err: map[string]error{"common_err": errors.New("common err")}, isErr: true, httpCode: http.StatusBadRequest, errObject: &ErrorObject{Message: map[string]string{"common_err": "common err"}}},

		{name: "record not found", err: ErrRecordNotFound, isErr: true, httpCode: http.StatusNotFound, errObject: &ErrorObject{Message: ErrRecordNotFound.Error()}},
		{name: "sql error no rows", err: sql.ErrNoRows, isErr: true, httpCode: http.StatusNotFound, errObject: &ErrorObject{Message: ErrRecordNotFound.Error()}},
		{name: "sql error conn done", err: sql.ErrConnDone, isErr: true, httpCode: http.StatusInternalServerError, errObject: &ErrorObject{Message: sql.ErrConnDone.Error()}},
		{name: "sql error tx done", err: sql.ErrTxDone, isErr: true, httpCode: http.StatusInternalServerError, errObject: &ErrorObject{Message: sql.ErrTxDone.Error()}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			errCode, errObject := MakeResponse(testCase.err)
			assert.Equal(t, testCase.errObject, errObject, testCase.name)
			assert.Equal(t, testCase.httpCode, errCode, testCase.name)
		})
	}
}

func setupRouter() *gin.Engine {
	InitValidator()
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
		assert.Equal(t, "{\"error\":{\"message\":\"route not found\"}}\n", w.Body.String())
	})
}

func makeValidationError() error {
	// MyStruct ..
	type MyStruct struct {
		String string `validate:"is-awesome"`
	}

	// use a single instance of Validate, it caches struct info
	var validate *validator.Validate

	validate = validator.New()
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
