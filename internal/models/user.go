package models

type User struct {
	Id      string  `json:"id" db:"id"`
	Balance float64 `json:"balance" db:"balance"`
}

type UserChangeBalanceQuery struct {
	Id    string  `json:"id" swaggertype:"string" format:"base64" example:"34be95d0-9a41-11ec-b909-0242ac120003"`
	Money float64 `json:"money" swaggertype:"number" format:"base64" example:"100"`
}

type UserTransferBalanceQuery struct {
	FromId string  `json:"from_id" swaggertype:"string" format:"base64" example:"34be95d0-9a41-11ec-b909-0242ac120003"`
	ToId   string  `json:"to_id" swaggertype:"string" format:"base64" example:"34be95d0-9a41-11ec-b909-0242ac120004"`
	Money  float64 `json:"money" swaggertype:"number" format:"base64" example:"50"`
}
