package validator

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var ErrNameInvalid = errors.New("invalid name")
var ErrValueInt64 = errors.New("invalid value int64")
var ErrValueFloat64 = errors.New("invalid value float64")

var nameRegexp = regexp.MustCompile(`^[a-zA-Z]{3}[a-zA-Z0-9_.-]*$`)

type Validator struct {
}

func New() *Validator {
	return &Validator{}
}

func (v Validator) ValidateName(name string) error {
	if !nameRegexp.MatchString(name) {
		return fmt.Errorf("%w: %q", ErrNameInvalid, name)
	}
	return nil
}

func (v Validator) ValidateValueInt64(value string) (int64, error) {

	value = strings.TrimSpace(value)
	val, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("value: %q, %w: %w", value, ErrValueInt64, err)
	}
	return val, nil
}

func (v Validator) ValidateValueFloat64(value string) (float64, error) {
	value = strings.TrimSpace(value)
	val, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("value: %q, %w: %w", value, ErrValueFloat64, err)
	}
	return val, nil
}
