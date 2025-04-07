package util

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DSSD-Madison/gmu/pkg/db"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBHandler struct {
	Config  *Config
	Pool    *pgxpool.Pool
	SqlDB   *sql.DB
	Querier *db.Queries
}

func (h *DBHandler) Close() error {
	h.Pool.Close()
	err := h.SqlDB.Close()
	if err != nil {
		return err
	}
	return nil
}

func NewDBHandler() (*DBHandler, error) {

	dbConfig, err := LoadConfig()

	databaseURL := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		dbConfig.DBUser, dbConfig.DBPassword, dbConfig.DBHost, dbConfig.DBName,
	)

	// Connect to PostgreSQL using pgxpool
	dbpool, err := pgxpool.Connect(context.Background(), databaseURL)
	if err != nil {
		return nil, err
	}

	// Create a *sql.DB instance using the pgx driver
	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}

	dbQuerier := db.New(sqlDB)

	return &DBHandler{
		Config:  dbConfig,
		Pool:    dbpool,
		SqlDB:   sqlDB,
		Querier: dbQuerier,
	}, nil
}
