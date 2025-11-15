package errors

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/e1esm/casino-transaction-system/api-gateway/src/internal/svcerr"
	"github.com/stretchr/testify/assert"
)

func TestParseSvcErrToResp(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedErrStr string
		httpStatus     int
	}{
		{
			name:           "Nil error",
			err:            nil,
			expectedErrStr: "",
			httpStatus:     http.StatusOK,
		},
		{
			name:           "Bad request",
			err:            fmt.Errorf("%w: field is empty", svcerr.ErrBadField),
			expectedErrStr: svcerr.ErrBadField.Error(),
			httpStatus:     http.StatusBadRequest,
		},
		{
			name:           "Not found",
			err:            fmt.Errorf("%w: entity not found", svcerr.ErrNotFound),
			expectedErrStr: svcerr.ErrNotFound.Error(),
			httpStatus:     http.StatusNotFound,
		},
		{
			name:           "Internal server error",
			err:            fmt.Errorf("unknown error"),
			expectedErrStr: "",
			httpStatus:     http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, txt := ParseSvcErrToResp(tt.err)

			assert.Equal(t, tt.httpStatus, code)
			assert.Contains(t, txt, tt.expectedErrStr)
		})
	}
}
