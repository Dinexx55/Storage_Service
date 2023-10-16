package model

type StoreVersion struct {
	VersionID     int    `db:"version_id"`
	StoreID       string `db:"store_id"`
	VersionNumber int    `db:"version_number" binding:"required"`
	CreatorLogin  string `db:"creator_login" binding:"required"`
	OwnerName     string `db:"owner_name" binding:"required"`
	OpeningTime   string `db:"opening_time" binding:"required"`
	ClosingTime   string `db:"closing_time" binding:"required"`
	CreatedAt     string `db:"created_at" binding:"required"`
	IsLast        bool   `db:"is_last" binding:"required"`
}
