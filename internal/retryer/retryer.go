package retryer

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"
)

const (
	baseInterval     = 1 * time.Second
	maxSleepInterval = 10 * time.Second
)

type Retryer struct{}

func NewRetryer() *Retryer {
	return &Retryer{}
}

/*
RetryLinear - повторяем вызов ф-ции attempts раз, с интервалом в stepSeconds секунд.
Если attempts задать равным nil, то будет бесконечное повторение.
Если stepSeconds = 0, то интервалы между попытами будут равны baseInterval
*/
func (r *Retryer) RetryLinear(ctx context.Context, fn func(ctx context.Context) error, stepSeconds int, attempts *int) error {

	// Инициализируем счётчик попыток
	var attemptsCounter *int
	if attempts != nil {
		attemptsCounter = new(int)
		*attemptsCounter = *attempts + 1
	}

	sleepInterval := baseInterval
	sleepDuration := 0 * time.Millisecond

	var err error

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:

			// Если была ошибка — делаем паузу
			if sleepDuration > 0 {
				time.Sleep(sleepDuration)

				// Увеличиваем интервал (линейно)
				sleepInterval += time.Duration(stepSeconds) * time.Second
				if sleepInterval > maxSleepInterval {
					sleepInterval = maxSleepInterval
				}
				sleepDuration = 0
			}

			// Проверяем, есть ли ещё попытки
			hasAttempts := attemptsCounter == nil || *attemptsCounter > 0
			if hasAttempts {
				err = fn(ctx)
			} else {
				return fmt.Errorf("stop retry after %d attempts: %w", *attempts, err)
			}

			// Если ошибки нет, или она не из "повторяемых" — выходим
			if err == nil {
				return nil
			}

			if attemptsCounter != nil {
				*attemptsCounter--
			}

			// Готовим паузу для следующей итерации
			sleepDuration = r.jitter(sleepInterval)
		}
	}
}

// jitter - Добавляем случайность ±20%
func (r *Retryer) jitter(sleepInterval time.Duration) time.Duration {
	jitter := time.Duration(rand.Float64() * float64(sleepInterval) * 0.2)
	return sleepInterval + jitter
}
