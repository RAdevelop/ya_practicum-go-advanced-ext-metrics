package retryer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

/*
ВНИМАНИЕ!
Медленные тесты - так как работают с паузами во время выполнения
*/
func TestRetryer_RetryLinear(t *testing.T) {

	type given struct {
		stepSeconds int
		attempts    *int
	}
	type want struct {
		hasError    bool
		countFnCall int //сколько раз была вызвана ф-ция
		elapsedTime time.Duration
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
			},
			want: want{
				hasError:    false,
				countFnCall: 1,
				elapsedTime: time.Duration(0) * time.Second, //0с, потому что нет попыток
			},
		},
		{
			name: "countFnCall must be 2 after 1 attempts and error in result",
			given: given{
				stepSeconds: 1,
				attempts:    new(1),
			},
			want: want{
				hasError:    true,
				countFnCall: 2,                              // потому что отсчет попыток идет после 0-го вызова ф-ции (он мог быть удачным)
				elapsedTime: time.Duration(1) * time.Second, //1с, потому что одна попытка с шагом в stepSeconds: 1-я пауза в 1 сек
			},
		},
		{
			name: "countFnCall must be 3 after 2 attempts and error in result",
			given: given{
				stepSeconds: 1,
				attempts:    new(2),
			},
			want: want{
				hasError:    true,
				countFnCall: 3,                              // потому что отсчет попыток идет после 0-го вызова ф-ции (он мог быть удачным)
				elapsedTime: time.Duration(3) * time.Second, //3с, потому что две попытки с шагом в stepSeconds: 1-я пауза в 1 сек, 2-я пауза в 2 сек: 1+2=3сек
			},
		},
		{
			name: "countFnCall must be 4 after 3 attempts and error in result",
			given: given{
				stepSeconds: 2,
				attempts:    new(3),
			},
			want: want{
				hasError: true,
				// потому что отсчет попыток идет после 0-го вызова ф-ции (он мог быть удачным)
				countFnCall: 4,
				//10с, потому что 3 попытки с шагом в stepSeconds: 1-я пауза в 1 сек, 2-я пауза в 3 сек, 2-я пауза в 5 сек: 1+3+5=9сек + погрешность jitter
				elapsedTime: time.Duration(10) * time.Second,
			},
		},
		{
			name: "don't run an endless test",
			given: given{
				attempts: nil, // в коде теста выведем лог, чтобы не запустить бесконечный тест
			},
		},
	}

	r := NewRetryer()
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

			err := r.RetryLinear(ctx, func(ctx context.Context) error {

				counter++

				if counter == 1 {
					//первый вызов ф-ции будем считать, когда отчитывать паузы в попытках
					zeroCallTime = time.Now()
				}

				if counter > 1 {
					elapsed = time.Since(zeroCallTime)
				}

				if tt.want.hasError {
					return errors.New("ErrorRetryLinear")
				}

				return nil

			}, tt.given.stepSeconds, tt.given.attempts)

			if tt.want.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want.countFnCall, counter)
			assert.LessOrEqual(t, tt.want.elapsedTime, elapsed.Round(time.Second))
		})
	}
}
