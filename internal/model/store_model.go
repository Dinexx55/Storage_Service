package model

type Store struct {
	Name         string `db:"name" binding:"required"`
	Address      string `db:"address" binding:"required"`
	CreatorLogin string `db:"creatorLogin" binding:"required"`
	OwnerName    string `db:"ownerName" binding:"required"`
	OpeningTime  string `db:"openingTime" binding:"required"`
	ClosingTime  string `db:"closingTime" binding:"required"`
	CreatedAt    string `db:"created_at" binding:"required"`
}
