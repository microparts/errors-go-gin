package ginerrors

import (
	"errors"
	"net/http"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/go-playground/validator.v9"
)

type GRPCValidationError struct {
	Violations []*errdetails.BadRequest_FieldViolation
}

var (
	httpToGRPCCodes = map[int]codes.Code{
		http.StatusNotFound:            codes.NotFound,
		http.StatusMethodNotAllowed:    codes.Unavailable,
		http.StatusInternalServerError: codes.Internal,
		http.StatusBadRequest:          codes.InvalidArgument,
	}

	grpcToHTTPCodes = map[codes.Code]int{
		codes.NotFound:        http.StatusNotFound,
		codes.Unavailable:     http.StatusMethodNotAllowed,
		codes.Internal:        http.StatusInternalServerError,
		codes.InvalidArgument: http.StatusBadRequest,
	}
)

func UnwrapRPCError(st *status.Status) interface{} {
	switch st.Code() {
	case codes.NotFound:
		return ErrRecordNotFound
	case codes.Unavailable:
		return ErrNoMethod
	case codes.Internal:
		return ErrServerError
	case codes.InvalidArgument:
		details := st.Details()
		violations := getViolations(details)

		return GRPCValidationError{Violations: violations}
	default:
		return errors.New(st.Message())
	}
}

func WrapErrorWithStatus(err error, lang string) error {
	l := langName(lang)
	httpCode, msg := getErrCode(err)

	code, ok := getGRPCCode(httpCode)
	if !ok {
		return err
	}

	st := status.New(code, msg)

	if code == codes.InvalidArgument {
		if errs, ok := err.(validator.ValidationErrors); ok {
			var e error
			br := &errdetails.BadRequest{}
			for _, ee := range errs {
				field := getFieldName(ee.Namespace(), ee.Field())
				msg := getErrMessage(validationRule(ee.ActualTag()), field, ee.Param(), l)
				violation := &errdetails.BadRequest_FieldViolation{
					Field:       string(field),
					Description: string(msg),
				}

				br.FieldViolations = append(br.FieldViolations, violation)
			}

			st, e = st.WithDetails(br)
			if e != nil {
				return err
			}
		}
	}

	return st.Err()
}

// getGRPCCode returns grpc error code by first value and its existence by second
func getGRPCCode(httpCode int) (codes.Code, bool) {
	c, ok := httpToGRPCCodes[httpCode]
	return c, ok
}

func getViolations(details []interface{}) []*errdetails.BadRequest_FieldViolation {
	if len(details) == 0 {
		return nil
	}

	violations := make([]*errdetails.BadRequest_FieldViolation, 0)
	for _, detail := range details {
		if d, ok := detail.(*errdetails.BadRequest); ok {
			violations = append(violations, d.FieldViolations...)
		}
	}

	return violations
}
