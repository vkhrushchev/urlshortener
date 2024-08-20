package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/vkhrushchev/urlshortener/internal/app/db"
	"github.com/vkhrushchev/urlshortener/internal/app/dto"
	"github.com/vkhrushchev/urlshortener/internal/middleware"
	"github.com/vkhrushchev/urlshortener/internal/util"
)

var ErrConflictOnUniqueConstraint = errors.New("storage_db: unique constraint conflict")

type DBStorage struct {
	dbLookup *db.DBLookup
}

func NewDBStorage(dbLookup *db.DBLookup) *DBStorage {
	return &DBStorage{
		dbLookup: dbLookup,
	}
}

func (s *DBStorage) GetURLByShortURI(ctx context.Context, shortURI string) (longURL string, found bool, err error) {
	db := s.dbLookup.GetDB()

	sqlRow := db.QueryRowContext(ctx, "SELECT su.original_url FROM short_url su WHERE su.short_url = $1", shortURI)

	err = sqlRow.Scan(&longURL)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return "", false, fmt.Errorf("db: error when get original URL by short URI")
		}
		return "", false, nil
	}

	return longURL, true, nil
}

func (s *DBStorage) SaveURL(ctx context.Context, longURL string) (shortURI string, err error) {
	userID := ctx.Value(middleware.UserIDContextKey)
	db := s.dbLookup.GetDB()

	shortURI = util.RandStringRunes(10)
	_, err = db.ExecContext(
		ctx,
		"INSERT INTO short_url(uuid, short_url, original_url, user_id) VALUES($1, $2, $3, $4)",
		uuid.New().String(),
		shortURI,
		longURL,
		userID,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				row := db.QueryRowContext(ctx, "SELECT su.short_url FROM short_url su WHERE su.original_url = $1", longURL)
				if row.Err() != nil {
					err = fmt.Errorf("db: error when search existed short URL: %v", err)
					return "", err
				}

				err = row.Scan(&shortURI)
				if err != nil {
					err = fmt.Errorf("db: error when scan existed short URL: %v", err)
					return "", err
				}

				return shortURI, ErrConflictOnUniqueConstraint
			}
		}

		err = fmt.Errorf("db: error when save short URL: %v", err)
		return "", err
	}

	return shortURI, nil
}

func (s *DBStorage) SaveURLBatch(ctx context.Context, entries []*dto.StorageShortURLEntry) ([]*dto.StorageShortURLEntry, error) {
	userID := ctx.Value(middleware.UserIDContextKey)
	db := s.dbLookup.GetDB()

	tx, err := db.Begin()
	if err != nil {
		err = fmt.Errorf("db: error when begin transaction: %v", err)
		return nil, err
	}
	defer func() {
		err := tx.Rollback()
		if !errors.Is(err, pgx.ErrTxClosed) {
			log.Errorw(
				"db: error when rollback transaction",
				"error",
				err.Error())
		}
	}()

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO short_url(uuid, short_url, original_url) VALUES($1, $2, $3, $4)")
	if err != nil {
		err = fmt.Errorf("db: error when create prepared statement: %v", err)
		return nil, err
	}

	for _, entry := range entries {
		shortURI := util.RandStringRunes(10)
		_, err = stmt.ExecContext(ctx, entry.UUID, shortURI, entry.LongURL, userID)
		if err != nil {
			err = fmt.Errorf("db: error when save entry to 'short_url' table: %v", err)
			return nil, err
		}
		entry.ShortURI = shortURI
	}

	if err = tx.Commit(); err != nil {
		err = fmt.Errorf("db: error when save entry to 'short_url' table: %v", err)
		return nil, err
	}

	return entries, nil
}

func (s *DBStorage) GetURLByUserID(ctx context.Context, userID string) ([]*dto.StorageShortURLEntry, error) {
	db := s.dbLookup.GetDB()

	rows, err := db.QueryContext(ctx, "SELECT su.uuid, su.short_url, su.original_url FROM short_url su WHERE su.user_id = $1", userID)
	if err != nil {
		err = fmt.Errorf("db: error when execute_query: %w", err)
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Errorw("db: error when close sql.Rows: %v", err)
		}
	}()

	result := make([]*dto.StorageShortURLEntry, 0)
	for rows.Next() {
		resultEntry := dto.StorageShortURLEntry{}
		if err := rows.Scan(&resultEntry.UUID, &resultEntry.ShortURI, &resultEntry.LongURL); err != nil {
			err = fmt.Errorf("db: error when parse sql.Rows: %w", err)
			return nil, err
		}
		result = append(result, &resultEntry)
	}

	return result, nil
}