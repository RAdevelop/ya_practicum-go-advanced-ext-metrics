package converter

import (
	"fmt"
	"strconv"
)

/*
NumericToString - конвертация чисел в строку

Примечание: чтобы не увеличивать кодовую базу, пока не стал делать проверки на +/- бесконечности для параметра "v"
*/
func NumericToString(v any) string {
	switch val := v.(type) {
	case int:
		return strconv.FormatInt(int64(val), 10)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	case int64:
		return strconv.FormatInt(val, 10)
	case uint64:
		return strconv.FormatUint(val, 10)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	default:
		return ""
	}
}

func ToFloat64(v any) (float64, error) {

	if v == nil {
		return 0, fmt.Errorf("cannot convert nil to float64")
	}

	var f float64
	switch val := v.(type) {
	case int:
		f = float64(val)
	case uint:
		f = float64(val)
	case int32:
		f = float64(val)
	case uint32:
		f = float64(val)
	case int64:
		f = float64(val)
	case uint64:
		f = float64(val)
	case float32:
		f = float64(val)
	case float64:
		f = val
	case string:
		var err error
		f, err = strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot parse string to float64: %w", err)
		}
	default:
		return 0, fmt.Errorf("unsupported type: %T", v)
	}

	return f, nil
}
