package ginerrors

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gopkg.in/go-playground/validator.v9"

	errs "github.com/microparts/errors-go"
)

//Response makes common error response
func Response(c *gin.Context, err interface{}) {
	errCode, data := MakeResponse(err, getLang(c))
	resp := errs.Response{Error: *data}
	c.AbortWithStatusJSON(errCode, resp)
}

//MakeResponse makes ErrorObject based on error type
func MakeResponse(err interface{}, lang langName) (int, *errs.ErrorObject) {
	errObj := &errs.ErrorObject{}
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
		errObj.Validation = makeErrorsSlice(et, lang)

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
