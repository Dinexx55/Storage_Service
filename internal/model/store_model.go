package model

type Store struct {
	StoreID      int    `db:"store_id"`
	Name         string `db:"name" binding:"required"`
	Address      string `db:"address" binding:"required"`
	CreatorLogin string `db:"creator_login" binding:"required"`
	OwnerName    string `db:"owner_name" binding:"required"`
	OpeningTime  string `db:"opening_time" binding:"required"`
	ClosingTime  string `db:"closing_time" binding:"required"`
	CreatedAt    string `db:"created_at" binding:"required"`
}
