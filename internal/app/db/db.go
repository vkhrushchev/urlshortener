package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

var log = zap.Must(zap.NewProduction()).Sugar()

const createShortURLTableSQL = `create table if not exists short_url
(
	uuid varchar(36) not null constraint short_url_pk primary key,
	short_url varchar(20) not null,
	original_url text not null
);`
const createUniqueIndexOnOriginalURLSQL = `create unique index if not exists short_url_original_url_uindex on short_url (original_url);`
const addUserIDColumnSQL = `alter table short_url add if not exists user_id varchar(36) not null;`
const addIsDeletedColumn = `alter table short_url add if not exists is_deleted boolean not null;`

type DBLookup struct {
	db *sql.DB
}

func NewDBLookup(databaseDSN string) (*DBLookup, error) {
	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return nil, fmt.Errorf("db: error when open database: %s", err.Error())
	}

	return &DBLookup{
		db: db,
	}, nil
}

func (d *DBLookup) InitDB(ctx context.Context) error {
	log.Infow("db: run createShortUrlTableSQL...")
	_, err := d.db.ExecContext(ctx, createShortURLTableSQL)
	if err != nil {
		return fmt.Errorf("db: error when execute createShortUrlTableSQL: %v", err)
	}
	log.Infow("db: run createShortUrlTableSQL... success")

	log.Infow("db: run createUniqueIndexOnOriginalURLSQL...")
	_, err = d.db.ExecContext(ctx, createUniqueIndexOnOriginalURLSQL)
	if err != nil {
		return fmt.Errorf("db: error when execute createUniqueIndexOnOriginalURLSQL: %v", err)
	}
	log.Infow("db: run createUniqueIndexOnOriginalURLSQL... success")

	log.Infow("db: run addUserIDColumnSQL...")
	_, err = d.db.ExecContext(ctx, addUserIDColumnSQL)
	if err != nil {
		return fmt.Errorf("db: error when execute addUserIDColumnSQL: %v", err)
	}
	log.Infow("db: run addUserIDColumnSQL... success")

	log.Infow("db: run addIsDeletedColumn...")
	_, err = d.db.ExecContext(ctx, addIsDeletedColumn)
	if err != nil {
		return fmt.Errorf("db: error when execute addIsDeletedColumn: %v", err)
	}
	log.Infow("db: run addIsDeletedColumn... success")

	return nil
}

func (d *DBLookup) Ping(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	err := d.db.PingContext(ctx)
	if err != nil {
		log.Errorw("db: error when ping database connection: %v", err)
	}

	return err == nil
}

func (d *DBLookup) GetDB() *sql.DB {
	return d.db
}
