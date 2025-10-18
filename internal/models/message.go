package models

import "time"

type Message struct {
	ID        int       `db:"id"`
	UserID    int       `db:"user_id"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
	IsBot     bool      `db:"is_bot"`
}
