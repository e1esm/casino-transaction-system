package handlers

import (
	"testing"

	"github.com/e1esm/casino-transaction-system/api-gateway/src/internal/entities"
	"github.com/stretchr/testify/assert"
)

func TestParseFiltersStruct(t *testing.T) {
	tests := []struct {
		name     string
		filters  string
		expected entities.TransactionFilter
		isErr    bool
	}{
		{
			name:    "both filters provided",
			filters: "{\"UserID\":\"461805a5-d762-441b-91ac-961629f926e7\",\"Type\":\"bet\"}",
			expected: entities.TransactionFilter{
				UserID: "461805a5-d762-441b-91ac-961629f926e7",
				Type:   "bet",
			},
			isErr: false,
		},
		{
			name:     "filters not provided",
			filters:  "{}",
			expected: entities.TransactionFilter{},
			isErr:    false,
		},
		{
			name:     "invalid filters provided",
			filters:  "{\"Typ",
			expected: entities.TransactionFilter{},
			isErr:    true,
		},
	}

	for _, tt := range tests {
		resp, err := parseFiltersStruct(tt.filters)

		if tt.isErr {
			assert.NotNil(t, err)
			continue
		}

		assert.Equal(t, resp, tt.expected)
		assert.Nil(t, err)

	}
}

func TestStrToIntWithDefault(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		defaultV int64
		expected int64
	}{
		{
			name:     "success - using parsed value",
			str:      "15",
			expected: 15,
			defaultV: 1,
		},
		{
			name:     "error - using default value",
			str:      "abc",
			defaultV: 10,
			expected: 10,
		},
	}

	for _, tt := range tests {
		resp := strToIntWithDefault(tt.str, tt.defaultV)
		assert.Equal(t, tt.expected, resp, tt.name)
	}
}
