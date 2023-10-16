package postgres

import (
	"StorageService/internal/config"
	"StorageService/internal/model"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"strconv"
)

func ConnectToPostgresDB(config *config.DB, logger *zap.Logger) (*sqlx.DB, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
	)
	logger.Info("connection string :" + connStr)
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

	_, err = tx.Exec("SET TRANSACTION ISOLATION LEVEL SERIALIZABLE")
	if err != nil {
		tx.Rollback()
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
        RETURNING store_id
    `

	var storeID int
	namedQuery, args, err := sqlx.Named(storeQuery, store)
	if err != nil {
		return err
	}
	err = tx.QueryRowx(tx.Rebind(namedQuery), args...).Scan(&storeID)
	if err != nil {
		return err
	}
	storeIdstr := strconv.Itoa(storeID)

	version := model.StoreVersion{
		StoreID:       storeIdstr,
		VersionNumber: 1,
		CreatorLogin:  store.CreatorLogin,
		OwnerName:     store.OwnerName,
		OpeningTime:   store.OpeningTime,
		ClosingTime:   store.ClosingTime,
		CreatedAt:     store.CreatedAt,
		IsLast:        true,
	}
	versionQuery := `
        INSERT INTO store_versions (store_id, version_number, creator_login, owner_name,
                                    opening_time, closing_time, created_at, is_last)
        VALUES ( :store_id, :version_number, :creator_login, :owner_name,
                :opening_time, :closing_time, :created_at, :is_last)
    `
	_, err = tx.NamedExec(versionQuery, version)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) CreateStoreVersion(storeVersion model.StoreVersion) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}

	_, err = tx.Exec("SET TRANSACTION ISOLATION LEVEL SERIALIZABLE")
	if err != nil {
		tx.Rollback()
		return err
	}

	var previousVersion model.StoreVersion
	err = tx.Get(&previousVersion, "SELECT * FROM store_versions WHERE store_id = $1 AND is_last = true", storeVersion.StoreID)
	if err != nil && err != sql.ErrNoRows {
		tx.Rollback()
		return err
	}

	if previousVersion.StoreID != "" {
		previousVersion.IsLast = false
		_, err = tx.Exec("UPDATE store_versions SET is_last = false WHERE store_id = $1 AND is_last = true", storeVersion.StoreID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	storeVersion.VersionNumber = previousVersion.VersionNumber + 1

	_, err = tx.Exec(`INSERT INTO store_versions (store_id, version_number, creator_login,
                            owner_name, opening_time, closing_time, created_at, is_last)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		storeVersion.StoreID, storeVersion.VersionNumber, storeVersion.CreatorLogin, storeVersion.OwnerName,
		storeVersion.OpeningTime, storeVersion.ClosingTime, storeVersion.CreatedAt, storeVersion.IsLast)

	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) DeleteStore(storeId string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("SET TRANSACTION ISOLATION LEVEL SERIALIZABLE")
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	err = r.DeleteStoreVersions(storeId)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	query := `
        DELETE FROM stores
        WHERE store_id = $1
    `
	_, err = tx.Exec(query, storeId)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return nil
}

func (r *Repository) DeleteStoreVersion(versionId string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("SET TRANSACTION ISOLATION LEVEL SERIALIZABLE")
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	query := `
        DELETE FROM store_versions
        WHERE version_id = $1
    `
	_, err = tx.Exec(query, versionId)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return nil
}

func (r *Repository) GetStoreByID(storeId string) (*model.Store, error) {
	query := `
        SELECT store_id, name, address, creator_login, owner_name, opening_time, closing_time, created_at
        FROM stores
        WHERE store_id = $1
    `
	store := &model.Store{}
	err := r.db.Get(store, query, storeId)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return store, nil
}

func (r *Repository) GetStoreVersionHistory(storeId string) ([]*model.StoreVersion, error) {
	query := `
        SELECT version_id, store_id, version_number, creator_login, owner_name, opening_time, closing_time, created_at, is_last
        FROM store_versions
        WHERE store_id = $1
        ORDER BY created_at DESC
    `
	storeVersions := []*model.StoreVersion{}
	err := r.db.Select(&storeVersions, query, storeId)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return storeVersions, nil
}

func (r *Repository) GetStoreVersionByID(versionId string) (*model.StoreVersion, error) {
	query := `
        SELECT version_id, store_id, version_number, creator_login, owner_name, opening_time, closing_time, created_at, is_last
        FROM store_versions
        WHERE version_id = $1
    `
	storeVersion := &model.StoreVersion{}
	err := r.db.Get(storeVersion, query, versionId)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return storeVersion, nil
}

func (r *Repository) GetStoreVersionForStore(storeId, versionId string) (*model.StoreVersion, error) {
	query := `
        SELECT version_id, store_id, version_number, creator_login, owner_name, opening_time, closing_time, created_at, is_last
        FROM store_versions
        WHERE version_id = $1 AND store_id = $2
    `
	storeVersion := &model.StoreVersion{}
	err := r.db.Get(storeVersion, query, versionId, storeId)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return storeVersion, nil
}

func (r *Repository) DeleteStoreVersions(storeId string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("SET TRANSACTION ISOLATION LEVEL SERIALIZABLE")
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	query := `
        DELETE FROM store_versions
        WHERE store_id = $1
    `
	_, err = tx.Exec(query, storeId)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return nil
}
