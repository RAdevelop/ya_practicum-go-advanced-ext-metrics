package validator

import (
	"math"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidator_ValidateName(t *testing.T) {
	tests := []struct {
		name   string
		hasErr bool
		names  []string
	}{
		{
			name:   "valid name",
			hasErr: false,
			names: []string{
				"name",
				"nameValid",
				"nameValid123",
				"name123Valid",
				"nameValid0",
				"nameValid-",
				"nameValid_",
				"nameValid_-",
				"nameValid_1-2_3",
				"name-Valid_1-2_3",
				"name-1_2_0_Valid_1-2_3",
				"name-1.2_0.Valid_1-2_3.",
			},
		},
		{
			name:   "invalid name",
			hasErr: true,
			names: []string{
				"",
				"0",
				"1",
				"123",
				"1234",
				"1name123Valid",
				"-nameInValid",
				"-nameInValid-",
				".nameInValid",
				".nameInValid.",
				"_nameInValid",
				"_nameInValid_",
				"-1nameInValid",
				"-1nameInValid-",
				".1nameInValid",
				".1nameInValid.",
				"_1nameInValid",
				"_1nameInValid_",
				"-",
				".",
				"-",
				"-1",
				".2",
				"-3",
				"----",
				"____",
				"....",
				"-123",
				"_123",
				".123",
			},
		},
	}
	validator := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			for _, name := range tt.names {
				err := validator.ValidateName(name)
				if tt.hasErr {
					assert.ErrorIs(t, err, ErrNameInvalid, "validateName should return an error for name: %s", name)
				} else {
					assert.NoError(t, err, "validateName should not return an error for name: %s", name)
				}
			}
		})
	}
}

func TestValidator_ValidateValueInt64(t *testing.T) {
	tests := []struct {
		inputValue       string
		expectedOutValue int64
		err              error
	}{
		{
			inputValue:       "0",
			expectedOutValue: 0,
			err:              nil,
		},
		{
			inputValue:       "-0",
			expectedOutValue: 0,
			err:              nil,
		},
		{
			inputValue:       "123",
			expectedOutValue: 123,
			err:              nil,
		},
		{
			inputValue:       " 123",
			expectedOutValue: 123,
			err:              nil,
		},
		{
			inputValue:       "123 ",
			expectedOutValue: 123,
			err:              nil,
		},
		{
			inputValue:       " 123 ",
			expectedOutValue: 123,
			err:              nil,
		},
		{
			inputValue:       strconv.FormatInt(math.MaxInt64, 10),
			expectedOutValue: math.MaxInt64,
			err:              nil,
		},
		{
			inputValue:       strconv.FormatInt(math.MinInt64, 10),
			expectedOutValue: math.MinInt64,
			err:              nil,
		},
		{
			inputValue:       "9223372036854775808",
			expectedOutValue: 0,
			err:              ErrValueInt64,
		},
		{
			inputValue:       "-9223372036854775809",
			expectedOutValue: 0,
			err:              ErrValueInt64,
		},
		{
			inputValue:       "85070591730234615847396907784232501249",
			expectedOutValue: 0,
			err:              ErrValueInt64,
		},
		{
			inputValue:       "-85070591730234615847396907784232501249",
			expectedOutValue: 0,
			err:              ErrValueInt64,
		},
		{
			inputValue:       "-",
			expectedOutValue: 0,
			err:              ErrValueInt64,
		},
		{
			inputValue:       "",
			expectedOutValue: 0,
			err:              ErrValueInt64,
		},
		{
			inputValue:       "string",
			expectedOutValue: 0,
			err:              ErrValueInt64,
		},
		{
			inputValue:       "строка",
			expectedOutValue: 0,
			err:              ErrValueInt64,
		},
		{
			inputValue:       "123строка",
			expectedOutValue: 0,
			err:              ErrValueInt64,
		},
		{
			inputValue:       "строка123",
			expectedOutValue: 0,
			err:              ErrValueInt64,
		},
		{
			inputValue:       "string123",
			expectedOutValue: 0,
			err:              ErrValueInt64,
		},
		{
			inputValue:       "123string",
			expectedOutValue: 0,
			err:              ErrValueInt64,
		},
	}
	validator := New()
	for _, tt := range tests {
		t.Run(tt.inputValue, func(t *testing.T) {

			value, err := validator.ValidateValueInt64(tt.inputValue)

			if tt.err != nil {
				assert.ErrorIs(t, err, tt.err)
				assert.Equal(t, tt.expectedOutValue, value)
			} else {
				assert.NoError(t, err, "validateValueInt64 should not return an error for value: %s", tt.inputValue)
				assert.Equal(t, tt.expectedOutValue, value)
			}
		})
	}
}

func TestValidator_ValidateValueFloat64(t *testing.T) {
	tests := []struct {
		inputValue       string
		expectedOutValue float64
		err              error
	}{
		{
			inputValue:       "0",
			expectedOutValue: 0,
			err:              nil,
		},
		{
			inputValue:       "-0",
			expectedOutValue: 0,
			err:              nil,
		},
		{
			inputValue:       "123",
			expectedOutValue: 123,
			err:              nil,
		},
		{
			inputValue:       " 123",
			expectedOutValue: 123,
			err:              nil,
		},
		{
			inputValue:       "123 ",
			expectedOutValue: 123,
			err:              nil,
		},
		{
			inputValue:       " 123 ",
			expectedOutValue: 123,
			err:              nil,
		},
		{
			inputValue:       "0.00000000000001",
			expectedOutValue: 0.00000000000001,
			err:              nil,
		},
		{
			inputValue:       "-0.00000000000001",
			expectedOutValue: -0.00000000000001,
			err:              nil,
		},
		{
			inputValue:       "123.00000000000001",
			expectedOutValue: 123.00000000000001,
			err:              nil,
		},
		{
			inputValue:       " 123.00000000000001",
			expectedOutValue: 123.00000000000001,
			err:              nil,
		},
		{
			inputValue:       "123.00000000000001 ",
			expectedOutValue: 123.00000000000001,
			err:              nil,
		},
		{
			inputValue:       " 123.00000000000001 ",
			expectedOutValue: 123.00000000000001,
			err:              nil,
		},
		{
			inputValue:       strconv.FormatFloat(math.MaxFloat64, 'g', -1, 64),
			expectedOutValue: math.MaxFloat64,
			err:              nil,
		},
		{
			inputValue:       strconv.FormatFloat(-math.MaxFloat64, 'g', -1, 64),
			expectedOutValue: -math.MaxFloat64,
			err:              nil,
		},
		{
			inputValue:       "9223372036854775808",
			expectedOutValue: 9223372036854775808,
			err:              nil,
		},
		{
			inputValue:       "-9223372036854775809",
			expectedOutValue: -9223372036854775809,
			err:              nil,
		},
		{
			inputValue:       "85070591730234615847396907784232501249",
			expectedOutValue: 85070591730234615847396907784232501249,
			err:              nil,
		},
		{
			inputValue:       "-85070591730234615847396907784232501249",
			expectedOutValue: -85070591730234615847396907784232501249,
			err:              nil,
		},
		{
			inputValue:       "-",
			expectedOutValue: 0,
			err:              ErrValueFloat64,
		},
		{
			inputValue:       "",
			expectedOutValue: 0,
			err:              ErrValueFloat64,
		},
		{
			inputValue:       "string",
			expectedOutValue: 0,
			err:              ErrValueFloat64,
		},
		{
			inputValue:       "строка",
			expectedOutValue: 0,
			err:              ErrValueFloat64,
		},
		{
			inputValue:       "123строка",
			expectedOutValue: 0,
			err:              ErrValueFloat64,
		},
		{
			inputValue:       "строка123",
			expectedOutValue: 0,
			err:              ErrValueFloat64,
		},
		{
			inputValue:       "string123",
			expectedOutValue: 0,
			err:              ErrValueFloat64,
		},
		{
			inputValue:       "123string",
			expectedOutValue: 0,
			err:              ErrValueFloat64,
		},
	}

	validator := New()
	for _, tt := range tests {
		t.Run(tt.inputValue, func(t *testing.T) {

			value, err := validator.ValidateValueFloat64(tt.inputValue)

			if tt.err != nil {
				assert.ErrorIs(t, err, tt.err)
				assert.Equal(t, tt.expectedOutValue, value)
			} else {
				assert.NoError(t, err, "validateValueInt64 should not return an error for value: %s", tt.inputValue)
				assert.Equal(t, tt.expectedOutValue, value)
			}
		})
	}
}
