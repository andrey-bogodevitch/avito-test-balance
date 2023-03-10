package entity

import "time"

type Operation struct {
	ID          int64     `json:"id"`
	Amount      int64     `json:"amount"`
	CreatedAt   time.Time `json:"created_at"`
	Description string    `json:"description"`
	SenderID    *int64    `json:"sender_id"`
	RecipientID *int64    `json:"recipient_id"`
}
