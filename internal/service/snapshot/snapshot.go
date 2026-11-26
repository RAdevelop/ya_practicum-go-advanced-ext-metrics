// Package snapshot - сервис по сохранению промежуточных данных с метриками.
package snapshot

import "context"

//go:generate mockery
type Able interface {
	Load(context.Context) error
	Save(context.Context) error
}
