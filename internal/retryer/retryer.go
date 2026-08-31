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
Серегй, спасибо за оценку, и в частности по дженерикам :)

Меня вот смущает то, что если метод, который надо будет обернуть в Retry*, будет возвращать больше 2-х параметров - что посоветуете на это счет?
На N+1 возвращаемых параметров создавать "дубль" функций Retry* - как-то не очень хочется. Это я на перспективу если...

Интересно знать возможные подходы.

Делать что-то вроде:

func (s *SomeStruct) SomeMethod(ctx context.Context, v1 V1, v2 V2, ..., vN VN) (t1 Type1, t1 Type2, tN TypeN, , error) {
	var t1 Type1
	var t2 Type2
	...
	var tN TypeN
	var err error

	_, err = retryer.RetryLinear(ctx, func(ctx context.Context) (struct{}, error) {

		t1, t2, ..., tN, err = s.callMe(ctx, v1, v2, ..., vN)
		return struct{}{}, err
	}, 2, new(3))
	//Возврат нужных параметров
	return t1, t2, ..., tN, err
}

Но тогда как будто и дженерики не нужны для самих Retry*?!
*/

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
				timer := time.NewTimer(sleepDuration)
				select {
				case <-timer.C:
					timer.Stop()
					sleepDuration = 0

					// Увеличиваем интервал
					sleepInterval = sleepIntervalCalc(sleepInterval, stepSeconds)
					if sleepInterval > maxSleepInterval {
						sleepInterval = maxSleepInterval
					}
				case <-ctx.Done():
					timer.Stop()
					var result T
					return result, ctx.Err()
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
