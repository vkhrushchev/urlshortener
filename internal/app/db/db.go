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

func (d *DBLookup) Ping(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	err := d.db.PingContext(ctx)
	if err != nil {
		log.Errorw("db: error when ping database connection: %v", err)
	}

	return err == nil
}
