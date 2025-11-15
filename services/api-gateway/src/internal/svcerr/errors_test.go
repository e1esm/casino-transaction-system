package svcerr

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsBadRequest(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedResp bool
	}{
		{
			name:         "no error was passed",
			err:          nil,
			expectedResp: false,
		},
		{
			name:         "bad request error was passed",
			err:          fmt.Errorf("%w: missing id", ErrBadField),
			expectedResp: true,
		},
		{
			name:         "not found error was passed",
			err:          fmt.Errorf("%w: entry was not found", ErrNotFound),
			expectedResp: false,
		},
	}

	for _, tt := range tests {
		assert.Equal(t, IsBadRequest(tt.err), tt.expectedResp, tt.name)
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedResp bool
	}{
		{
			name:         "no error was passed",
			err:          nil,
			expectedResp: false,
		},
		{
			name:         "bad request error was passed",
			err:          fmt.Errorf("%w: missing id", ErrBadField),
			expectedResp: false,
		},
		{
			name:         "not found error was passed",
			err:          fmt.Errorf("%w: entry was not found", ErrNotFound),
			expectedResp: true,
		},
	}

	for _, tt := range tests {
		assert.Equal(t, IsNotFound(tt.err), tt.expectedResp, tt.name)
	}
}
