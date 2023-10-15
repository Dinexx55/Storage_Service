package postgres

import (
	"StorageService/internal/config"
	"fmt"
	"github.com/jmoiron/sqlx"
)

func ConnectToPostgresDB(config *config.DB) (*sqlx.DB, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
	)
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

type Repository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) *Repository {
	repo := &Repository{
		db: db,
	}

	return repo
}

func (r *Repository) Close() error {
	return r.db.Close()
}

type Store struct {
	ID          int    `db:"id"`
	Name        string `db:"name"`
	Address     string `db:"address"`
	OwnerName   string `db:"owner_name"`
	OpeningTime string `db:"opening_time"`
	ClosingTime string `db:"closing_time"`
}

func (r *Repository) CreateStore(store *Store) error {
	query := `
        INSERT INTO stores (name, address, owner_name, opening_time, closing_time)
        VALUES (:name, :address, :owner_name, :opening_time, :closing_time)
    `
	_, err := r.db.NamedExec(query, store)
	return err
}
