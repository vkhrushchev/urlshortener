package repository

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/vkhrushchev/urlshortener/internal/app/db"
	"github.com/vkhrushchev/urlshortener/internal/app/entity"
	"github.com/vkhrushchev/urlshortener/internal/middleware"
)

const (
	sqlInsertRow           = "INSERT INTO short_url(uuid, short_url, original_url, user_id, is_deleted) VALUES($1, $2, $3, $4, $5)"
	sqlSelectByShortURL    = "SELECT su.uuid, su.short_url, su.original_url, su.user_id, su.is_deleted FROM short_url su WHERE su.short_url = $1"
	sqlSelectByOriginalURL = "SELECT su.uuid, su.short_url, su.original_url, su.user_id, su.is_deleted FROM short_url su WHERE su.original_url = $1"
	sqlSelectByUserID      = "SELECT su.uuid, su.short_url, su.original_url, su.user_id, su.is_deleted FROM short_url su WHERE su.user_id = $1"
	sqlUpdateIsDeleted     = "UPDATE short_url SET is_deleted = true WHERE is_deleted = false AND short_url = $1 AND user_id = $2"
)

// DBShortURLRepository структура для хранения ссылки на db.DBLookup.
//
// Реализует интерфейс IShortURLRepository для хранения коротких ссылок в БД
type DBShortURLRepository struct {
	dbLookup *db.DBLookup
}

// NewDBShortURLRepository создает экземпляр структуры DBShortURLRepository
func NewDBShortURLRepository(dbLookup *db.DBLookup) *DBShortURLRepository {
	return &DBShortURLRepository{dbLookup: dbLookup}
}

// GetShortURLByShortURI возвращает короткую ссылку по shortURI
func (r *DBShortURLRepository) GetShortURLByShortURI(ctx context.Context, shortURI string) (entity.ShortURLEntity, error) {
	dbLookup := r.dbLookup.GetDB()

	sqlRow := dbLookup.QueryRowContext(
		ctx,
		sqlSelectByShortURL,
		shortURI,
	)

	shortURLEntity := entity.ShortURLEntity{}
	err := sqlRow.Scan(
		&shortURLEntity.UUID,
		&shortURLEntity.ShortURI,
		&shortURLEntity.LongURL,
		&shortURLEntity.UserID,
		&shortURLEntity.Deleted,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.ShortURLEntity{}, ErrNotFound
		}

		log.Errorw("repository: unexpected error", "err", err)
		return entity.ShortURLEntity{}, ErrUnexpected
	}

	return shortURLEntity, nil
}

// SaveShortURL сохраняет короткую ссылку
func (r *DBShortURLRepository) SaveShortURL(ctx context.Context, shortURLEntity *entity.ShortURLEntity) (*entity.ShortURLEntity, error) {
	dbLookup := r.dbLookup.GetDB()

	_, err := dbLookup.ExecContext(
		ctx,
		sqlInsertRow,
		shortURLEntity.UUID,
		shortURLEntity.ShortURI,
		shortURLEntity.LongURL,
		shortURLEntity.UserID,
		shortURLEntity.Deleted,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				sqlRow := dbLookup.QueryRowContext(ctx, sqlSelectByOriginalURL, shortURLEntity.LongURL)
				if sqlRow.Err() != nil {
					log.Errorw("repository: unexpected error", "err", err)
					return nil, ErrUnexpected
				}

				err = sqlRow.Scan(
					&shortURLEntity.UUID,
					&shortURLEntity.ShortURI,
					&shortURLEntity.LongURL,
					&shortURLEntity.UserID,
					&shortURLEntity.Deleted,
				)
				if err != nil {
					log.Errorw("repository: unexpected error", "err", err)
					return nil, ErrUnexpected
				}

				return shortURLEntity, ErrConflict
			}
		}

		log.Errorw("repository: unexpected error", "err", err)
		return nil, ErrUnexpected
	}

	return shortURLEntity, nil
}

// SaveShortURLs сохраняет короткие ссылки пачкой
func (r *DBShortURLRepository) SaveShortURLs(ctx context.Context, shortURLEntities []entity.ShortURLEntity) ([]entity.ShortURLEntity, error) {
	dbLookup := r.dbLookup.GetDB()

	tx, err := dbLookup.Begin()
	if err != nil {
		log.Errorw("repository: unexpected error", "err", err)
		return nil, ErrUnexpected
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			log.Errorw("repository: error when rollback transaction", "rollbackErr", rollbackErr)
		}
	}()

	stmt, err := tx.PrepareContext(ctx, sqlInsertRow)
	if err != nil {
		log.Errorw("repository: unexpected error", "err", err)
		return nil, ErrUnexpected
	}

	for _, shortURLEntity := range shortURLEntities {
		_, err = stmt.ExecContext(
			ctx,
			shortURLEntity.UUID,
			shortURLEntity.ShortURI,
			shortURLEntity.LongURL,
			shortURLEntity.UserID,
			shortURLEntity.Deleted,
		)
		if err != nil {
			log.Errorw("repository: unexpected error", "err", err)
			return nil, ErrUnexpected
		}
	}

	if err = tx.Commit(); err != nil {
		log.Errorw("repository: unexpected error", "err", err)
		return nil, ErrUnexpected
	}

	return shortURLEntities, nil
}

// GetShortURLsByUserID возвращает список коротких ссылок по userID
func (r *DBShortURLRepository) GetShortURLsByUserID(ctx context.Context, userID string) ([]entity.ShortURLEntity, error) {
	dbLookup := r.dbLookup.GetDB()

	rows, err := dbLookup.QueryContext(ctx, sqlSelectByUserID, userID)
	if err != nil {
		log.Errorw("repository: unexpected error", "err", err)
		return nil, ErrUnexpected
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Errorw("repository: unexpected error", "err", err)
		}
	}()

	result := make([]entity.ShortURLEntity, 0)
	for rows.Next() {
		resultEntry := entity.ShortURLEntity{}
		if err := rows.Scan(
			&resultEntry.UUID,
			&resultEntry.ShortURI,
			&resultEntry.LongURL,
			&resultEntry.UserID,
			&resultEntry.Deleted,
		); err != nil {
			log.Errorw("repository: unexpected error", "err", err)
			return nil, ErrUnexpected
		}

		result = append(result, resultEntry)
	}

	if err := rows.Err(); err != nil {
		log.Errorw("repository: unexpected error", "err", err)
		return nil, ErrUnexpected
	}

	return result, nil
}

// DeleteShortURLsByShortURIs удаляет короткие ссылки по списку shortURI
func (r *DBShortURLRepository) DeleteShortURLsByShortURIs(ctx context.Context, shortURIs []string) error {
	dbLookup := r.dbLookup.GetDB()

	tx, err := dbLookup.Begin()
	if err != nil {
		log.Errorw("repository: unexpected error", "err", err)
		return ErrUnexpected
	}

	shortURIsCh := deleteByShortURIsGenerator(shortURIs)

	stmt, err := tx.PrepareContext(ctx, sqlUpdateIsDeleted)
	if err != nil {
		log.Errorw("repository: unexpected error", "err", err)
		return ErrUnexpected
	}
	userID := ctx.Value(middleware.UserIDContextKey).(string)
	deleteByShortURIsTaskResultChannels := deleteByShortURIsTaskFanOut(ctx, stmt, &userID, shortURIsCh)
	deleteByShortURIsTaskFanIn(tx, deleteByShortURIsTaskResultChannels)

	return nil
}

func deleteByShortURIsGenerator(shortURIs []string) chan string {
	inputCh := make(chan string)

	go func() {
		defer close(inputCh)

		for _, data := range shortURIs {
			inputCh <- data
		}
	}()

	return inputCh
}

func deleteByShortURIsTask(ctx context.Context, stmt *sql.Stmt, userID *string, shortURIsCh chan string) chan int {
	deleteByShortURIsTaskResultCh := make(chan int)

	go func() {
		defer close(deleteByShortURIsTaskResultCh)

		for shortURI := range shortURIsCh {
			res, err := stmt.ExecContext(ctx, shortURI, userID)
			if err != nil {
				log.Errorw("repository: unexpected error", "err", err)
				continue
			}

			deletedCount, err := res.RowsAffected()
			if err != nil {
				log.Errorw("repository: unexpected error", "err", err)
			}

			log.Infow("repository: row marked as deleted", "shortURI", shortURI, "userID", userID)

			deleteByShortURIsTaskResultCh <- int(deletedCount)
		}
	}()

	return deleteByShortURIsTaskResultCh
}

func deleteByShortURIsTaskFanOut(ctx context.Context, stmt *sql.Stmt, userID *string, shortURIsCh chan string) []chan int {
	workerCount := 10
	deleteByShortURIsTaskResultChannels := make([]chan int, workerCount)
	for i := 0; i < workerCount; i++ {
		deleteByShortURIsTaskResultCh := deleteByShortURIsTask(ctx, stmt, userID, shortURIsCh)
		deleteByShortURIsTaskResultChannels[i] = deleteByShortURIsTaskResultCh
	}

	return deleteByShortURIsTaskResultChannels
}

func deleteByShortURIsTaskFanIn(tx *sql.Tx, deleteByShortURIsTaskResultChannels []chan int) {
	go func() {
		var wg sync.WaitGroup
		for _, deleteByShortURIsTaskResultCh := range deleteByShortURIsTaskResultChannels {
			deleteByShortURIsTaskResultChClosure := deleteByShortURIsTaskResultCh
			wg.Add(1)

			go func() {
				defer wg.Done()

				for range deleteByShortURIsTaskResultChClosure {
					// просто считываем данные из канала
				}
			}()
		}

		wg.Wait()

		if err := tx.Commit(); err != nil {
			log.Errorw("repository: unexpected error", "err", err)
		}
	}()
}
