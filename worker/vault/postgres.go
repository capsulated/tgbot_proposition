package vault

import (
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"gitlab.com/logiq.one/agenda_3dru_bot/config"
	"gitlab.com/logiq.one/agenda_3dru_bot/model"
	"time"
)

type Postgres struct {
	Db *sqlx.DB
}

func New(cfg *config.App) (*Postgres, error) {
	db, err := sqlx.Open("pgx", cfg.Database.ConnStr)
	if err != nil {
		return nil, err
	}

	return &Postgres{
		Db: db,
	}, nil
}

func (p *Postgres) ListUsers() (*[]model.User, error) {
	var users []model.User
	err := p.Db.Select(&users, `SELECT id, role_id, email, telegram_username, chat_id, created_at FROM "user"`)

	if err != nil {
		return nil, err
	}

	return &users, nil
}

func (p *Postgres) ListSecretaries() (*[]model.User, error) {
	var users []model.User
	err := p.Db.Select(&users, `
		SELECT id, role_id, email, telegram_username, chat_id, created_at 
		FROM "user"
		WHERE role_id = 2
	`)

	if err != nil {
		return nil, err
	}

	return &users, nil
}

func (p *Postgres) CreateUser(roleId int32, email string, telegramUsername string, chatId int64) (int64, error) {
	var id int64
	userState := `INSERT INTO "user" (role_id, email, telegram_username, chat_id) VALUES ($1, $2, $3, $4) RETURNING id`
	err := p.Db.Get(&id, userState, roleId, email, telegramUsername, chatId)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (p *Postgres) CreateInitiative(userId int64, question string) (int64, error) {
	var id int64
	initiativeState := `INSERT INTO initiative (user_id, question) VALUES  ($1, $2) RETURNING id`
	err := p.Db.Get(&id, initiativeState, userId, question)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (p *Postgres) VoteInitiative(question string, yes int, no int, archive int) error {
	voteState := `UPDATE initiative SET yes = $1, no = $2, archive = $3 WHERE question = $4`
	_, err := p.Db.Exec(voteState, yes, no, archive, question)
	return err
}

func (p *Postgres) ListWonInitiatives(weekAgo time.Time) (*[]model.WonInitiative, error) {
	var wonInitiatives []model.WonInitiative
	err := p.Db.Select(&wonInitiatives, `
		SELECT initiative.question, "user".email
		FROM initiative
		LEFT JOIN "user" ON initiative.user_id = "user".id
		WHERE initiative.yes > initiative.no
		AND initiative.yes > initiative.archive
		AND "user".created_at > $1
	`, weekAgo)

	if err != nil {
		return nil, err
	}

	return &wonInitiatives, nil
}
