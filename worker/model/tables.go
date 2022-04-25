package model

import (
	"time"
)

type Role struct {
	Id         int32  `db:"id"`
	Nomination string `db:"nomination"`
}

type User struct {
	Id               int32     `db:"id"`
	RoleId           int32     `db:"role_id"`
	Email            string    `db:"email"`
	TelegramUsername string    `db:"telegram_username"`
	CreatedAt        time.Time `db:"created_at"`
}

type Initiative struct {
	Id        int32     `db:"id"`
	UserId    string    `db:"printer_id"`
	Question  string    `db:"question"`
	Yes       int32     `db:"yes"`
	No        int32     `db:"no"`
	Archive   int32     `db:"archive"`
	CreatedAt time.Time `db:"created_at"`
}
