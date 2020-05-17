package ginerrors

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	unknownErrMessage        = "rpc error: code = InvalidArgument desc = foo"
	unknownErrValue          = "foo"
	recordNotFountMessage    = "rpc error: code = NotFound desc = record not found"
	unavailableMethodMessage = "rpc error: code = Unavailable desc = method not allowed"
)

func TestWrapErrWithRPC(t *testing.T) {
	cases := []struct {
		caseName string
		err      error
		lang     string
		expected string
	}{
		{
			caseName: "unknown err",
			err:      errors.New(unknownErrValue),
			lang:     "en",
			expected: unknownErrMessage,
		},
		{
			caseName: "sql err",
			err:      sql.ErrNoRows,
			lang:     "en",
			expected: recordNotFountMessage,
		},
		{
			caseName: "unavailable method",
			err:      ErrNoMethod,
			lang:     "en",
			expected: unavailableMethodMessage,
		},
	}
	for _, cc := range cases {
		t.Run(cc.caseName, func(t *testing.T) {
			err := WrapErrorWithStatus(cc.err, cc.lang)
			assert.Error(t, err)
			assert.Equal(t, cc.expected, err.Error())
		})
	}
}
