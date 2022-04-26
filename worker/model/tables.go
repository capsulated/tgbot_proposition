package model

import (
	"time"
)

type Role struct {
	Id         int64  `db:"id"`
	Nomination string `db:"nomination"`
}

type User struct {
	Id               int64     `db:"id"`
	RoleId           int32     `db:"role_id"`
	Email            string    `db:"email"`
	TelegramUsername string    `db:"telegram_username"`
	ChatId           int64     `db:"chat_id"`
	CreatedAt        time.Time `db:"created_at"`
}

type Initiative struct {
	Id        int64     `db:"id"`
	UserId    int64     `db:"user_id"`
	Question  string    `db:"question"`
	Yes       int       `db:"yes"`
	No        int       `db:"no"`
	Archive   int       `db:"archive"`
	CreatedAt time.Time `db:"created_at"`
}

type WonInitiative struct {
	Question string `db:"question"`
	Email    string `db:"email"`
}
