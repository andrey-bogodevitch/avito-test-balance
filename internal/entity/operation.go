package entity

import "time"

type Operation struct {
	Amount      int64
	CreatedAt   time.Time
	Description string
	SenderID    *int64
	RecipientID *int64
}
