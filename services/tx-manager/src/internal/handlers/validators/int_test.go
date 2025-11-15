package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateGreaterOrEqualTo(t *testing.T) {
	tests := []struct {
		name         string
		input        []int
		target       int
		expectedResp bool
	}{
		{
			name:         "empty input",
			input:        []int{},
			target:       100,
			expectedResp: true,
		},
		{
			name:         "single less input",
			input:        []int{1},
			target:       70,
			expectedResp: false,
		},
		{
			name:         "single greater input",
			input:        []int{100},
			target:       70,
			expectedResp: true,
		},
	}

	for _, tt := range tests {
		resp := ValidateGreaterOrEqualTo(tt.target, tt.input...)
		assert.Equal(t, tt.expectedResp, resp, tt.name)
	}
}
