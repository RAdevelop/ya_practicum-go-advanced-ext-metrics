// Package audit - аудит получаемых метрик.
package audit

import "context"

// Auditor — приёмник событий аудита (Observer).
type Auditor interface {
	Save(ctx context.Context, event Event) error
	Close() error
}
