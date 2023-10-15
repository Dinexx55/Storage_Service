package model

type StoreVersion struct {
	StoreID        string `db:"shop_id"`
	CreatorLogin   string `db:"creator_login" binding:"required"`
	StoreOwnerName string `db:"store_owner_name" binding:"required"`
	OpeningTime    string `db:"opening_time" binding:"required"`
	ClosingTime    string `db:"closing_time" binding:"required"`
	CreatedAt      string `db:"created_at" binding:"required"`
}
