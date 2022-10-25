package models

import "time"

type Transaction struct {
	Id        string    `json:"id" db:"id"`
	ToId      *string   `json:"to_id,omitempty" db:"to_id"`
	FromId    *string   `json:"from_id,omitempty" db:"from_id"`
	Money     float64   `json:"money" db:"money"`
	Operation string    `json:"operation" db:"operation"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type AllTransactionsGetQuery struct {
	Id       string `json:"id" swaggertype:"string" format:"base64" example:"34be95d0-9a41-11ec-b909-0242ac120003"`
	SortType string `json:"sort_type" swaggertype:"string" format:"base64" example:"date_asc"`
	Limit    int    `json:"limit" swaggertype:"number" format:"base64" example:"10"`
	Page     int    `json:"page" swaggertype:"number" format:"base64" example:"1"`
}
