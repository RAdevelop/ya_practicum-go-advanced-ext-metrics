package database

import (
	"context"
	"fmt"

	configDB "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/config/db"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
)

/*
Сергей, приветствую!
Или можно в PR отвечать? :)
```s-shpak
	Миграции сейчас накатываются вручную через make migrate-up -- отдельным шагом до старта сервера. Лучше запускать их программно прямо внутри NewDB, вызывая библиотеку наподобие golang-migrate/migrate сразу после создания пула соединений.

	При ручном шаге инициализация схемы выпадает из жизненного цикла бинарника: если запустить сервер без предварительного migrate-up, он поднимется, но упадёт при первом обращении к несуществующей таблице. Встроенный в NewDB запуск миграций гарантирует, что схема актуальна к моменту, когда сервер начинает принимать трафик, -- это особенно важно в CI/CD-пайплайнах и контейнерных деплоях, где бинарник не предполагает ручного сопровождения

	SQL-файлы миграций лежат в migrations/, но при сборке бинарника туда не попадают: приложение при деплое ждёт, что нужная директория окажется рядом с исполняемым файлом. Чтобы бинарник был самодостаточным, встрой файлы через embed: объяви //go:embed migrations/*.sql и передай embed.FS в функцию, которая запускает миграции.

	Альтернатива -- сделать путь к директории с миграциями параметром конфигурации (флаг или переменная окружения) и передавать его в NewDB. Это тоже приемлемо, но менее удобно: появляется дополнительный артефакт деплоя, версию которого нужно синхронизировать с бинарником
```

Для деплоя в git-е я добавил запуск миграций. Так что, все там разворачивается до тестов.

У меня 20+ лет в web разработке. Это я для информации, а не чтобы мериться... :)
В мире Go, уж очень мало опыта, тут не срокю :) Поэтому стараюсь все Ваши замечания, примечения исправлять у себя.

За время этого опыта для себя вывел золоту середину - миграции запускаются отедльно, это дает:
- возможность накатывать миграции без неоходимости деплоя самого приложения (независимо от языка разработки)
	- Иначе, в такой ситуации, всегда и приложение пересобирать надо.
	- Например, возникнет необходимость создать триггер, индекс, права доступа, check огранияения и тп в БД. К приоржении он прямого отношения не имеет, но потребует пересобирать все...
- "пользователю", от имени которого запускаются миграции, выдаются свои привелегии, а "пользователю" приложения - свои.
	- разграничении ответственности, безопасность и тп
- возможность регулировать условия отката миграций в случае ошибок
	- не знаю, как тут дела обстоят с тем, когда это будет внутри приложения происходить (я про собрал бинарник, накатились миграции, но что-то пошло не так, и надо откатить бинарник...)
- конечно же, при ручном (да и любом, на самом деле) подходе к миграциям - чтобы они были "с обратной совместимостью" для кода приложения, и код приложения тоже был с обратной совместимостью
- чтобы код приложения не падал, создаю (и настаивал с коллегами на этом в командах), интеграционные тесты (я вообще сторонник TDD)
- ктстаи, если минрации добавлять через embed, размер же бинарника будет расти?
	- по поводу размера бинарника, пробовал найти информацию, как правильно исключить mocks.gen.go файлы, что-то пока не нашел
- да и просто ревьюить и проверять результат наката отката мигарции проще - не надо понимать само приложение для этого
- при необходимости (мало вероятно, и все же), можно добавить шаг в CI/CD между результатом выполнения миграций и началом тестирования/сборки приложения

PS. Пр миграцями + embed - если это отнесете к достаточно критичным замечаниям, я обещаю в относительно свободное время разрбраться с этим подходом! :)
*/

//go:generate mockery
type Database interface {
	Close()
	Ping(context.Context) error
}

type ExecutorAble interface {
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

/*
DB - выполняем запросы к БД
ВАЖНО! Есть методы (contextWithTx, txFromContext), которые обновляют исходный контекст!

Очень интересно будет узнать мнение по такому подходу работы с транзакциями - передавать через контекст.
Почему пошел по такому пути - совсем не хотелось менять интерфейс internal/service/metric/memory.go.
Так как для работы с хранилищем в памяти транзакция не требуется. С другой стороны, можно ее там тоже реализовать с
блокировкой/разблокировкой при изменении свойств, в которых хранятся метрики...
*/
type DB struct {
	pool *pgxpool.Pool
}

func NewDB(ctx context.Context, cfg configDB.ConfigProvider, logger logger.Logger) (*DB, error) {

	config, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.MaxConns() > 0 {
		config.MaxConns = cfg.MaxConns()
	}
	if cfg.MinConns() > 0 {
		config.MinConns = cfg.MinConns()
	}
	if cfg.MaxConnLifetime() > 0 {
		config.MaxConnLifetime = cfg.MaxConnLifetime()
	}
	if cfg.MaxConnIdleTime() > 0 {
		config.MaxConnIdleTime = cfg.MaxConnIdleTime()
	}

	// Подключаем адаптер логгера
	if logger != nil {
		config.ConnConfig.Tracer = &tracelog.TraceLog{
			Logger:   newPgxLoggerAdapter(logger, tracelog.LogLevelError),
			LogLevel: tracelog.LogLevelError,
		}
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	return &DB{
		pool: pool,
	}, nil
}

func (db *DB) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

// Ping - проверка доступности БД
func (db *DB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

// txKey — типизированный ключ, чтобы избежать коллизий в context.
type txKey struct{}

// contextWithTx кладёт транзакцию в контекст.
func (db *DB) contextWithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// txFromContext достаёт транзакцию из контекста.
func (db *DB) txFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}

func (db *DB) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	// Создаём контекст с транзакцией
	txCtx := db.contextWithTx(ctx, tx)

	// Передаём txCtx внутрь callback fn — именно его будут использовать storage.*
	if err = fn(txCtx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx) // Commit/Rollback используют исходный ctx для дедлайна или отмены
}

func (db *DB) Executor(ctx context.Context) ExecutorAble {
	// Пытаемся достать транзакцию из контекста
	if tx, ok := db.txFromContext(ctx); ok && tx != nil {
		return tx
	}
	// Иначе используем пул соединений
	return db.pool
}
