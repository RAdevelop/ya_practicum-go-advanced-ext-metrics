package converter

import "strconv"

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
