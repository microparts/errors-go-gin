package ginerrors

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gopkg.in/go-playground/validator.v9"
)

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
	ErrNotFound       = errors.New("route not found")
	ErrNoMethod       = errors.New("method not allowed")
	ErrServerError    = errors.New("internal server error")
	ErrRecordNotFound = errors.New("record not found")
)

//Response makes common error response
func Response(c *gin.Context, err interface{}) {
	errCode, data := MakeResponse(err)
	resp := ResponseObject{Error: *data}
	c.AbortWithStatusJSON(errCode, resp)
}

//MakeResponse makes ErrorObject based on error type
func MakeResponse(err interface{}) (int, *ErrorObject) {
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
		errObj.Validation = MakeErrorsSlice(et)

	case error:
		errCode = getErrCode(et)

		errObj.Message = et.Error()

	case map[string]error:
		msgs := make(map[string]string)
		for k, e := range et {
			msgs[k] = e.Error()
		}

		errObj.Message = msgs
	}

	return errCode, errObj
}

func getErrCode(et error) (errCode int) {
	switch et.Error() {
	case ErrNotFound.Error():
		errCode = http.StatusNotFound
	case ErrNoMethod.Error():
		errCode = http.StatusMethodNotAllowed
	case ErrServerError.Error(), sql.ErrConnDone.Error(), sql.ErrTxDone.Error():
		errCode = http.StatusInternalServerError
	case ErrRecordNotFound.Error(), sql.ErrNoRows.Error():
		errCode = http.StatusNotFound
	default:
		errCode = http.StatusBadRequest
	}
	return
}
