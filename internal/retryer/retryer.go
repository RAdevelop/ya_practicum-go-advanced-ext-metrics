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
RetryLinear - повторяем вызов ф-ции attempts раз, с линейным ростом интервала в stepSeconds секунд.
Если attempts задать равным nil, то будет бесконечное повторение.
Если stepSeconds = 0, то интервалы между попытами будут равны baseInterval
*/
func (r *Retryer) RetryLinear(ctx context.Context, fn func(ctx context.Context) error, stepSeconds uint, attempts *int) error {

	linearCalc := func(sleepInterval time.Duration, stepSeconds uint) time.Duration {
		sleepInterval += time.Duration(stepSeconds) * time.Second
		return sleepInterval
	}

	return r.retryFn(ctx, fn, stepSeconds, attempts, linearCalc)
}

/*
RetryExponential - повторяем вызов ф-ции attempts раз, с ростом по экспоненте интервала в stepSeconds секунд.
Если attempts задать равным nil, то будет бесконечное повторение.
Если stepSeconds = 0, то интервалы между попытами будут равны baseInterval
*/
func (r *Retryer) RetryExponential(ctx context.Context, fn func(ctx context.Context) error, stepSeconds uint, attempts *int) error {

	exponentialCalc := func(sleepInterval time.Duration, stepSeconds uint) time.Duration {
		sleepInterval *= time.Duration(stepSeconds)
		return sleepInterval
	}

	return r.retryFn(ctx, fn, stepSeconds, attempts, exponentialCalc)
}

func (r *Retryer) retryFn(ctx context.Context, fn func(ctx context.Context) error, stepSeconds uint, attempts *int, sleepIntervalCalc func(sleepInterval time.Duration, stepSeconds uint) time.Duration) error {

	// Инициализируем счётчик попыток
	var attemptsCounter *int
	if attempts != nil {
		attemptsCounter = new(int)
		*attemptsCounter = *attempts + 1 // +1, потому что "0-я" попытка это для потенциального успешного выполнения
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
				sleepDuration = 0

				// Увеличиваем интервал
				sleepInterval = sleepIntervalCalc(sleepInterval, stepSeconds)

				if sleepInterval > maxSleepInterval {
					sleepInterval = maxSleepInterval
				}
			}

			// Проверяем, есть ли ещё попытки
			hasAttempts := attemptsCounter == nil || *attemptsCounter > 0
			if hasAttempts {
				err = fn(ctx)

				if attemptsCounter != nil {
					*attemptsCounter--
				}
			} else {
				return fmt.Errorf("stop retry after %d attempts: %w", *attempts, err)
			}

			// Если ошибки нет, или она не из "повторяемых" — выходим
			if err == nil {
				return nil
			}

			// Готовим паузу для следующей итерации
			sleepDuration = sleepInterval + r.jitter()

		}
	}
}

// jitter - Добавляем случайность ±20%
func (r *Retryer) jitter() time.Duration {
	return time.Duration(rand.Float64() * float64(baseInterval) * 0.2)
}
