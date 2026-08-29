// Package perrors - предопределенные ошибки для проекта
package perrors

import "errors"

var ErrMetricListEmpty = errors.New("metric list is empty")
var ErrMetricUpdate = errors.New("metric update error")
var ErrMetricIsNil = errors.New("metric is nil")
var ErrMetricUnknownType = errors.New("unknown metric type")
var ErrMetricEmptyID = errors.New("metric id is empty")
var ErrMetricEmptyValue = errors.New("metric value is empty")
var ErrMetricNotFound = errors.New("metric not found")
