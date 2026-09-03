package retryer

import (
	"errors"
	"net"
	"os"
	"syscall"

	"github.com/jackc/pgx/v5/pgconn"
)

// errorClassification тип для классификации ошибок
type errorClassification int

const (
	// nonRetriable - операцию не следует повторять
	nonRetriable errorClassification = iota

	// retriable - операцию можно повторить
	retriable
)

func IsRetriableError(err error) bool {
	if err == nil {
		return false
	}

	if isRetriableNetworkError(err) == retriable {
		return true
	}

	// Проверяем и конвертируем в pgconn.PgError, если это возможно
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		return isRetriablePgError(pgErr) == retriable
	}

	// Файловые ошибки (постоянные)
	if errors.Is(err, os.ErrNotExist) ||
		errors.Is(err, os.ErrExist) ||
		errors.Is(err, os.ErrPermission) ||
		errors.Is(err, os.ErrInvalid) ||
		errors.Is(err, os.ErrClosed) {
		return false
	}

	// По умолчанию считаем ошибку неповторяемой
	return false
}

func isRetriablePgError(pgErr *pgconn.PgError) errorClassification {

	if classification, ok := pgErrCode[pgErr.Code]; ok {
		return classification
	}

	// По умолчанию считаем ошибку неповторяемой
	return nonRetriable
}

// pgErrCode - Коды ошибок PostgreSQL: https://www.postgresql.org/docs/current/errcodes-appendix.html
var pgErrCode = map[string]errorClassification{

	"08000": retriable, //connection_exception	Общая ошибка соединения — можно переподключиться.
	"08003": retriable, //connection_does_not_exist	Соединение, к которому пытаются обратиться, не существует (обычно уже закрыто). Переподключение решит проблему.
	"08006": retriable, //connection_failure	Соединение потеряно (обрыв). Повторная попытка установит новое соединение.
	"08001": retriable, //sqlclient_unable_to_establish_sqlconnection	Клиент не смог установить соединение с сервером. Возможные причины: сетевые проблемы, сервер не запущен. Повтор попытки через интервал имеет смысл.
	"08004": retriable, //	sqlserver_rejected_establishment_of_sqlconnection	Сервер отклонил попытку соединения (например, из-за перегрузки или превышения лимита подключений). Повтор через некоторое время может быть успешным.
	"08007": retriable, //	transaction_resolution_unknown	Статус транзакции неизвестен (например, после сбоя соединения). Это не совсем ошибка соединения, но повтор транзакции с новым соединением — стандартная стратегия.
	"08P01": retriable, //protocol_violation	Нарушение протокола связи. Может быть вызвано временным сбоем в сети или несовместимостью версий. Переподключение часто помогает.

	"40001": retriable, //	serialization_failure | ошибка сериализации (конфликт параллельных транзакций)
	"40P01": retriable, //	deadlock_detected | обнаружен взаимоблокировка (deadlock)

	"57P03": retriable, // cannot_connect_now | невозможно подключиться сейчас

	"HV00N": retriable, //  fdw_unable_to_establish_connection | ошибка подключения к внешнему источнику
}

func isRetriableNetworkError(err error) errorClassification {

	if netErr, ok := errors.AsType[net.Error](err); ok {
		if netErr.Temporary() || netErr.Timeout() {
			return retriable
		}
	}

	if opErr, ok := errors.AsType[*net.OpError](err); ok {
		if errors.Is(opErr.Err, syscall.ECONNREFUSED) ||
			errors.Is(opErr.Err, syscall.ECONNRESET) ||
			errors.Is(opErr.Err, syscall.ETIMEDOUT) ||
			errors.Is(opErr.Err, syscall.EHOSTDOWN) ||
			errors.Is(opErr.Err, syscall.EHOSTUNREACH) ||
			errors.Is(opErr.Err, syscall.ENETDOWN) ||
			errors.Is(opErr.Err, syscall.ENETUNREACH) {
			return retriable
		}
	}

	// Системные ошибки напрямую
	if errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.ETIMEDOUT) ||
		errors.Is(err, syscall.EHOSTDOWN) ||
		errors.Is(err, syscall.EHOSTUNREACH) ||
		errors.Is(err, syscall.ENETDOWN) ||
		errors.Is(err, syscall.ENETUNREACH) {
		return retriable
	}

	return nonRetriable
}
