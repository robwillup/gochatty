package models

import "time"

type User struct {
	ID           int       `db:"id"`
	Username     string    `db:"username"`
	PasswordHash string    `db:"password_hash"`
	CreatedAd    time.Time `db:"created_at"`
}
