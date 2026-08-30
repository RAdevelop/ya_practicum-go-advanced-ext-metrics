package retryer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

/*
ВНИМАНИЕ!!!
Медленные тесты - так как проверяют паузы во время выполнения!
*/
func TestRetryer_RetryLinear(t *testing.T) {

	t.Log("ВНИМАНИЕ!!! Медленные тесты - так как проверяют паузы во время выполнения!")

	type given struct {
		stepSeconds uint
		attempts    *int
		result      any
	}
	type want struct {
		hasError    bool
		countFnCall int //сколько раз была вызвана ф-ция
		elapsedTime time.Duration
		result      any
	}

	tests := []struct {
		name  string
		given given
		want  want
	}{
		{
			name: "countFnCall must be 1 without attempts",
			given: given{
				stepSeconds: 1,
				attempts:    new(1),
				result:      []int{1, 2, 3},
			},
			want: want{
				hasError:    false,
				countFnCall: 1,
				elapsedTime: time.Duration(0) * time.Second, //0с, потому что нет попыток
				result:      []int{1, 2, 3},
			},
		},
		{
			name: "countFnCall must be 4 after 3 attempts and error in result",
			given: given{
				stepSeconds: 2,
				attempts:    new(3),
				result:      "123",
			},
			want: want{
				hasError: true,
				// потому что отсчет попыток идет после 0-го вызова ф-ции (он мог быть удачным)
				countFnCall: 4,
				//10с, потому что 3 попытки с шагом в stepSeconds: 1-я пауза в 1 сек, 2-я пауза в 3 сек, 2-я пауза в 5 сек: 1+3+5=9сек + погрешность jitter
				elapsedTime: time.Duration(9) * time.Second,
				result:      "",
			},
		},
		{
			name: "don't run an endless test",
			given: given{
				attempts: nil, // в коде теста выведем лог, чтобы не запустить бесконечный тест
			},
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.given.attempts == nil {
				t.Log("don't run an endless test")
				return
			}

			counter := 0

			var zeroCallTime time.Time
			var elapsed time.Duration

			result, err := RetryLinear(ctx, func(ctx context.Context) (any, error) {

				counter++

				if counter == 1 {
					//первый вызов ф-ции будем считать, когда отчитывать паузы в попытках
					zeroCallTime = time.Now()
				}

				if counter > 1 {
					elapsed = time.Since(zeroCallTime)
				}

				if tt.want.hasError {
					return "", errors.New("ErrorRetryLinear")
				}

				return tt.given.result, nil

			}, tt.given.stepSeconds, tt.given.attempts)

			if tt.want.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want.result, result)
			assert.Equal(t, tt.want.countFnCall, counter)

			elapsedRound := elapsed.Round(time.Millisecond)
			assert.LessOrEqual(t, tt.want.elapsedTime, elapsedRound)
			assert.GreaterOrEqual(t, tt.want.elapsedTime+time.Second, elapsedRound)
		})
	}
}

func TestRetryer_RetryExponential(t *testing.T) {

	t.Log("ВНИМАНИЕ!!! Медленные тесты - так как проверяют паузы во время выполнения!")

	type given[T any] struct {
		stepSeconds uint
		attempts    *int
		result      T
	}
	type want[T any] struct {
		hasError    bool
		countFnCall int //сколько раз была вызвана ф-ция
		elapsedTime time.Duration
		result      T
	}

	tests := []struct {
		name  string
		given given[any]
		want  want[any]
	}{
		{
			name: "countFnCall must be 1 without attempts",
			given: given[any]{
				stepSeconds: 1,
				attempts:    new(1),
				result:      "123",
			},
			want: want[any]{
				hasError:    false,
				countFnCall: 1,
				elapsedTime: time.Duration(0) * time.Second, //0с, потому что нет попыток
				result:      "123",
			},
		},
		{
			name: "countFnCall must be 3 after 2 attempts and error in result",
			given: given[any]{
				stepSeconds: 2,
				attempts:    new(2),
				result:      0,
			},
			want: want[any]{
				hasError:    true,
				countFnCall: 3,
				elapsedTime: time.Duration(3) * time.Second, //0с, потому что нет попыток
				result:      0,
			},
		},
		{
			name: "don't run an endless test",
			given: given[any]{
				attempts: nil, // в коде теста выведем лог, чтобы не запустить бесконечный тест
			},
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.given.attempts == nil {
				t.Log("don't run an endless test")
				return
			}

			counter := 0

			var zeroCallTime time.Time
			var elapsed time.Duration

			result, err := RetryExponential(ctx, func(ctx context.Context) (any, error) {

				counter++

				if counter == 1 {
					//первый вызов ф-ции будем считать, когда отчитывать паузы в попытках
					zeroCallTime = time.Now()
				}

				if counter > 1 {
					elapsed = time.Since(zeroCallTime)
				}

				if tt.want.hasError {
					return tt.given.result, errors.New("RetryExponential")
				}

				return tt.given.result, nil

			}, tt.given.stepSeconds, tt.given.attempts)

			if tt.want.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want.result, result)

			assert.Equal(t, tt.want.countFnCall, counter)
			elapsedRound := elapsed.Round(time.Millisecond)
			assert.LessOrEqual(t, tt.want.elapsedTime, elapsedRound)
			assert.GreaterOrEqual(t, tt.want.elapsedTime+time.Second, elapsedRound)
		})
	}
}
