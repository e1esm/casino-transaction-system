package errors

import (
	"errors"
	"testing"

	"github.com/e1esm/casino-transaction-system/tx-manager/src/internal/svcerr"
	"github.com/stretchr/testify/assert"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestErrorCastingAndParsing(t *testing.T) {
	tests := []struct {
		name           string
		fn             func() (error, bool)
		expectedCode   codes.Code
		expectedMsg    string
		expectedSecond bool
	}{
		{
			name: "CastInvalidRequest returns InvalidArgument",
			fn: func() (error, bool) {
				err := errors.New("invalid input")
				return CastInvalidRequest(err), false
			},
			expectedCode:   codes.InvalidArgument,
			expectedMsg:    "invalid input",
			expectedSecond: false,
		},
		{
			name: "CastNotFound returns NotFound",
			fn: func() (error, bool) {
				err := errors.New("missing tx")
				return CastNotFound(err), false
			},
			expectedCode:   codes.NotFound,
			expectedMsg:    "missing tx",
			expectedSecond: false,
		},
		{
			name: "Parse handles BadRequest error",
			fn: func() (error, bool) {
				return ParseSvcErrToProto(svcerr.ErrBadField)
			},
			expectedCode:   codes.InvalidArgument,
			expectedMsg:    svcerr.ErrBadField.Error(),
			expectedSecond: false,
		},
		{
			name: "Parse handles NotFound",
			fn: func() (error, bool) {
				return ParseSvcErrToProto(svcerr.ErrNotFound)
			},
			expectedCode:   codes.NotFound,
			expectedMsg:    svcerr.ErrNotFound.Error(),
			expectedSecond: false,
		},
		{
			name: "Parse handles internal errors",
			fn: func() (error, bool) {
				return ParseSvcErrToProto(errors.New("internal error"))
			},
			expectedCode:   codes.Internal,
			expectedMsg:    "",
			expectedSecond: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err, second := tt.fn()

			assert.Equal(t, tt.expectedSecond, second, tt.name)
			assert.Equal(t, tt.expectedSecond, second, tt.name)

			st, ok := status.FromError(err)

			assert.True(t, ok, tt.name)
			assert.Equal(t, tt.expectedCode, st.Code(), tt.name)
			assert.Equal(t, tt.expectedMsg, st.Message(), tt.name)
		})
	}
}
