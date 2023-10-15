package postgres

import (
	"StorageService/internal/config"
	"StorageService/internal/model"
	"database/sql"
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

func (r *Repository) CreateStore(store model.Store) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	storeQuery := `
        INSERT INTO stores (name, address, creator_login, owner_name, opening_time, closing_time, created_at)
        VALUES (:name, :address, :creator_login, :owner_name, :opening_time, :closing_time, :created_at)
    `
	_, err = tx.NamedExec(storeQuery, store)
	if err != nil {
		return err
	}

	var storeID string
	err = tx.Get(&storeID, "SELECT LAST_INSERT_ID()")
	if err != nil {
		return err
	}

	version := model.StoreVersion{
		StoreID:        storeID,
		CreatorLogin:   store.CreatorLogin,
		StoreOwnerName: store.OwnerName,
		OpeningTime:    store.OpeningTime,
		ClosingTime:    store.ClosingTime,
		CreatedAt:      store.CreatedAt,
	}
	versionQuery := `
        INSERT INTO store_versions (store_id, creator_login, owner_name, opening_time, closing_time, created_at)
        VALUES ( :store_id, :creator_login, :owner_name, :opening_time, :closing_time, :created_at)
    `
	_, err = tx.NamedExec(versionQuery, version)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) CreateStoreVersion(storeVersion model.StoreVersion) error {
	query := `
        INSERT INTO store_versions (store_id, name, address, owner_name, opening_time, closing_time)
        VALUES (:store_id, :name, :address, :owner_name, :opening_time, :closing_time)
    `
	_, err := r.db.NamedExec(query, storeVersion)
	return err
}

func (r *Repository) DeleteStore(storeId string) error {
	err := r.DeleteStoreVersions(storeId)
	if err != nil {
		return err
	}

	query := `
        DELETE FROM stores
        WHERE id = ?
    `
	_, err = r.db.Exec(query, storeId)
	return err
}

func (r *Repository) DeleteStoreVersion(versionId string) error {
	query := `
        DELETE FROM store_versions
        WHERE version_id = ?
    `
	_, err := r.db.Exec(query, versionId)
	return err
}

func (r *Repository) GetStoreByID(storeId string) (*model.Store, error) {
	query := `
        SELECT id, name, address, owner_name, opening_time, closing_time
        FROM stores
        WHERE id = ?
    `
	store := &model.Store{}
	err := r.db.Get(store, query, storeId)
	if err == sql.ErrNoRows {
		return nil, nil // Магазин не найден
	} else if err != nil {
		return nil, err // Произошла ошибка при выполнении запроса
	}
	return store, nil
}

func (r *Repository) GetStoreVersionHistory(storeId string) ([]*model.StoreVersion, error) {
	query := `
        SELECT version_id, shop_id, name, address, owner_name, opening_time, closing_time, created_at
        FROM store_versions
        WHERE shop_id = ?
        ORDER BY created_at DESC
    `
	storeVersions := []*model.StoreVersion{}
	err := r.db.Select(&storeVersions, query, storeId)
	if err == sql.ErrNoRows {
		return nil, nil // История версий не найдена
	} else if err != nil {
		return nil, err // Произошла ошибка при выполнении запроса
	}
	return storeVersions, nil
}

func (r *Repository) GetStoreVersionByID(versionId string) (*model.StoreVersion, error) {
	query := `
        SELECT version_id, shop_id, name, address, owner_name, opening_time, closing_time, created_at
        FROM store_versions
        WHERE version_id = ?
    `
	storeVersion := &model.StoreVersion{}
	err := r.db.Get(storeVersion, query, versionId)
	if err == sql.ErrNoRows {
		return nil, nil // Версия магазина не найдена
	} else if err != nil {
		return nil, err // Произошла ошибка при выполнении запроса
	}
	return storeVersion, nil
}

func (r *Repository) DeleteStoreVersions(storeId string) error {
	query := `
        DELETE FROM store_versions
        WHERE shop_id = ?
    `
	_, err := r.db.Exec(query, storeId)
	return err
}
