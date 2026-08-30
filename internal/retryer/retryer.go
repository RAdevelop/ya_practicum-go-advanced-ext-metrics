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

/*
RetryLinear - повторяем вызов ф-ции attempts раз, с линейным ростом интервала в stepSeconds секунд.
Если attempts задать равным nil, то будет бесконечное повторение.
Если stepSeconds = 0, то интервалы между попытами будут равны baseInterval
*/
func RetryLinear[T any](ctx context.Context, fn func(ctx context.Context) (T, error), stepSeconds uint, attempts *int) (T, error) {

	linearCalc := func(sleepInterval time.Duration, stepSeconds uint) time.Duration {
		sleepInterval += time.Duration(stepSeconds) * time.Second
		return sleepInterval
	}

	return retryFn(ctx, fn, stepSeconds, attempts, linearCalc)
}

/*
RetryExponential - повторяем вызов ф-ции attempts раз, с ростом по экспоненте интервала в stepSeconds секунд.
Если attempts задать равным nil, то будет бесконечное повторение.
Если stepSeconds = 0, то интервалы между попытами будут равны baseInterval
*/
func RetryExponential[T any](ctx context.Context, fn func(ctx context.Context) (T, error), stepSeconds uint, attempts *int) (T, error) {
	exponentialCalc := func(sleepInterval time.Duration, stepSeconds uint) time.Duration {
		sleepInterval *= time.Duration(stepSeconds)
		return sleepInterval
	}
	return retryFn(ctx, fn, stepSeconds, attempts, exponentialCalc)
}

func retryFn[T any](ctx context.Context, fn func(ctx context.Context) (T, error), stepSeconds uint, attempts *int, sleepIntervalCalc func(sleepInterval time.Duration, stepSeconds uint) time.Duration) (T, error) {
	var result T
	var err error

	// Инициализируем счётчик попыток
	var attemptsCounter *int
	if attempts != nil {
		attemptsCounter = new(int)
		*attemptsCounter = *attempts + 1 // +1, потому что "0-я" попытка это для потенциального успешного выполнения
	}

	sleepInterval := baseInterval
	sleepDuration := 0 * time.Millisecond

	for {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
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
				result, err = fn(ctx)

				// Если ошибки нет, или она не из "повторяемых" — выходим
				if err == nil {
					return result, nil
				} else if !IsRetriableError(err) {
					var result T
					return result, err
				}

				if attemptsCounter != nil {
					*attemptsCounter--
				}
			} else {
				var result T
				return result, fmt.Errorf("stop retry after %d attempts: %w", *attempts, err)
			}

			// Готовим паузу для следующей итерации
			sleepDuration = sleepInterval + jitter()
		}
	}
}

// jitter - Добавляем случайность ±20%
func jitter() time.Duration {
	return time.Duration(rand.Float64() * float64(baseInterval) * 0.2)
}
