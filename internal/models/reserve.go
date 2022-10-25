package models

import "time"

type Reserve struct {
	Id           string     `json:"id" db:"id"`
	UserId       string     `json:"user_id" db:"user_id"`
	ServiceId    string     `json:"service_id" db:"service_id"`
	OrderId      string     `json:"order_id" db:"order_id"`
	Amount       float64    `json:"amount" db:"amount"`
	Status       string     `json:"status" db:"status"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	RecognizedAt *time.Time `json:"recognized_at,omitempty" db:"recognized_at"`
}

type ReserveMoneyQuery struct {
	UserId    string  `json:"user_id" db:"user_id" swaggertype:"string" format:"base64" example:"34be95d0-9a41-11ec-b909-0242ac120003"`
	ServiceId string  `json:"service_id" db:"service_id" swaggertype:"string" format:"base64" example:"someserviceid1"`
	OrderId   string  `json:"order_id" db:"order_id" swaggertype:"string" format:"base64" example:"someorderid1"`
	Amount    float64 `json:"amount" db:"amount" swaggertype:"number" format:"base64" example:"20"`
}
