package retryer

import (
	"errors"
	"os"

	"github.com/jackc/pgx/v5/pgconn"
)

// ErrorClassification тип для классификации ошибок
type ErrorClassification int

const (
	// NonRetriable - операцию не следует повторять
	NonRetriable ErrorClassification = iota

	// Retriable - операцию можно повторить
	Retriable
)

func IsRetriableError(err error) bool {
	if err == nil {
		return false
	}

	// Проверяем и конвертируем в pgconn.PgError, если это возможно
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		return isRetriablePgError(pgErr) == Retriable
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

func isRetriablePgError(pgErr *pgconn.PgError) ErrorClassification {

	if classification, ok := pgErrCode[pgErr.Code]; ok {
		return classification
	}

	// По умолчанию считаем ошибку неповторяемой
	return NonRetriable
}

// pgErrCode - Коды ошибок PostgreSQL: https://www.postgresql.org/docs/current/errcodes-appendix.html
var pgErrCode = map[string]ErrorClassification{

	"08000": Retriable, //connection_exception	Общая ошибка соединения — можно переподключиться.
	"08003": Retriable, //connection_does_not_exist	Соединение, к которому пытаются обратиться, не существует (обычно уже закрыто). Переподключение решит проблему.
	"08006": Retriable, //connection_failure	Соединение потеряно (обрыв). Повторная попытка установит новое соединение.
	"08001": Retriable, //sqlclient_unable_to_establish_sqlconnection	Клиент не смог установить соединение с сервером. Возможные причины: сетевые проблемы, сервер не запущен. Повтор попытки через интервал имеет смысл.
	"08004": Retriable, //	sqlserver_rejected_establishment_of_sqlconnection	Сервер отклонил попытку соединения (например, из-за перегрузки или превышения лимита подключений). Повтор через некоторое время может быть успешным.
	"08007": Retriable, //	transaction_resolution_unknown	Статус транзакции неизвестен (например, после сбоя соединения). Это не совсем ошибка соединения, но повтор транзакции с новым соединением — стандартная стратегия.
	"08P01": Retriable, //protocol_violation	Нарушение протокола связи. Может быть вызвано временным сбоем в сети или несовместимостью версий. Переподключение часто помогает.

	"40001": Retriable, //	serialization_failure | ошибка сериализации (конфликт параллельных транзакций)
	"40P01": Retriable, //	deadlock_detected | обнаружен взаимоблокировка (deadlock)

	"57P03": Retriable, // cannot_connect_now | невозможно подключиться сейчас

	"HV00N": Retriable, //  fdw_unable_to_establish_connection | ошибка подключения к внешнему источнику
}
