package telegram

import (
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gitlab.com/logiq.one/agenda_3dru_bot/config"
	"gitlab.com/logiq.one/agenda_3dru_bot/vault"
	"log"
	"net/mail"
	"sync"
)

type Bot struct {
	Users         sync.Map
	Vlt           *vault.Postgres
	PinInitiator  string
	PinSecretary  string
	MainChannelId int64
	Api           *tgbotapi.BotAPI
	Updates       tgbotapi.UpdatesChannel
}

type UserData struct {
	TelegramUsername string
	Role             Role
	PrevMessage      string
	Email            string
	Registered       bool
	Id               int64
}

type Role int32

const (
	Initiator Role = iota + 1
	Secretary
)

const commandStart = "start"

const pinAsk = "Введите пин-код"
const alreadyRegistered = "Вы уже зарегестрированы!"
const incorrectPin = "Неверный пинкод, повторите попытку:"
const inputEmail = "Введите адрес электронной почты"
const incorrectEmail = "Неверный адрес электронной почты!"
const initiatorRegistered = "Вы зарегистрированы. Введите Ваш вопрос на повестку планёрки одним сообщением. Приём вопросов осуществляется каждую неделю с понедельника по четверг"
const secretaryRegistered = "Вы зарегистрированы. Подверждённые вопросы на планёрку Вам будут отправлены каждую пятницу в 10.00"
const questionSent = "Спасибо, Ваш вопрос отправлен. Можете отправить следующий вопрос."

func New(cfg *config.App, vault *vault.Postgres) *Bot {
	api, err := tgbotapi.NewBotAPI(cfg.Telegram.ApiToken)
	if err != nil {
		log.Fatalf("%v", err)
	}

	api.Debug = cfg.Telegram.Debug

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	bot := &Bot{
		MainChannelId: cfg.Telegram.MainChannelId,
		Api:           api,
		Vlt:           vault,
		PinInitiator:  cfg.Pin.Initiator,
		PinSecretary:  cfg.Pin.Secretary,
		Updates:       api.GetUpdatesChan(u),
	}

	users, err := vault.ListUsers()
	if err != nil {
		log.Fatalf("%v", err)
	}

	for _, user := range *users {
		bot.Users.Store(
			user.ChatId,
			UserData{
				TelegramUsername: user.TelegramUsername,
				Role:             Role(user.RoleId),
				Email:            user.Email,
				Registered:       true,
				Id:               user.Id,
			},
		)
	}

	return bot
}

func (b *Bot) Listen() {
	for update := range b.Updates {
		// Если кто-то проголосовал
		if update.Poll != nil && update.Poll.Options != nil && len(update.Poll.Options) > 0 {
			// Обновить в БД строку с вопросом
			var yes, no, archive int
			for _, v := range update.Poll.Options {
				switch v.Text {
				case "Да":
					yes = v.VoterCount
				case "Нет":
					no = v.VoterCount
				case "Архив":
					archive = v.VoterCount
				}
			}
			err := b.Vlt.VoteInitiative(update.Poll.Question, yes, no, archive)
			if err != nil {
				log.Println(err)
			}
			continue
		}

		// Если сообщение пустое
		if update.Message == nil {
			continue
		}

		// Если сообщение из главного (mainChannelId) канала - значит это голосование
		// И автор - не бот
		if update.Message.Chat.ID == b.MainChannelId {
			continue
		}

		userId := update.Message.From.ID
		u, userExist := b.Users.Load(userId)
		var userData UserData
		if userExist {
			userData = u.(UserData)
		}

		// Если это команда /start и юзер существует и зарегестрирован (сообщение: Вы уже зареганы!)
		if update.Message.Command() == commandStart && userExist && userData.Registered {
			if _, err := b.Api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, alreadyRegistered)); err != nil {
				log.Println(err)
			}
			continue
		}

		// Если не зареган - попросим ввести пин-код и запишем данные о нём
		if update.Message.Command() == commandStart && !userExist {
			text := pinAsk

			// Создадим нового юзера
			userData = UserData{
				TelegramUsername: update.Message.From.UserName,
				PrevMessage:      text,
			}
			b.Users.Store(update.Message.From.ID, userData)

			if _, err := b.Api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, text)); err != nil {
				log.Println(err)
			}
			continue
		}

		// Юзер существует, но не зарегестрирован, а предыдущее сообщение было запросом ПИН-кода (или повторного ввода пин-кода)
		if userExist && !userData.Registered && (userData.PrevMessage == pinAsk || userData.PrevMessage == incorrectPin) {
			var text string

			if update.Message.Text == b.PinInitiator || update.Message.Text == b.PinSecretary {
				text = inputEmail

				var role = Initiator
				if update.Message.Text == b.PinSecretary {
					role = Secretary
				}

				// Обновим данные юзера
				b.Users.Store(
					userId,
					UserData{
						TelegramUsername: update.Message.From.UserName,
						Role:             role,
						PrevMessage:      text,
					},
				)
			} else {
				text = incorrectPin
			}

			if _, err := b.Api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, text)); err != nil {
				log.Println(err)
			}
			continue
		}

		// Юзер существует, но не зарегестрирован, а предыдущее сообщение было запросом емэйла
		if userExist && !userData.Registered && (userData.PrevMessage == inputEmail) {
			var text string

			email, err := mail.ParseAddress(update.Message.Text)
			if err != nil {
				text = incorrectEmail
			} else if userData.Role == Initiator {
				text = initiatorRegistered
			} else if userData.Role == Secretary {
				text = secretaryRegistered
			}

			id, err := b.Vlt.CreateUser(int32(userData.Role), email.Address, update.Message.From.UserName, update.Message.Chat.ID)
			if err != nil {
				log.Println(err)
				continue
			}

			// Обновим данные юзера
			b.Users.Store(
				userId,
				UserData{
					TelegramUsername: update.Message.From.UserName,
					Role:             userData.Role,
					PrevMessage:      text,
					Email:            email.Address,
					Registered:       true,
					Id:               id,
				},
			)

			if _, err = b.Api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, text)); err != nil {
				log.Println(err)
			}
			continue
		}

		// Юзер существует, зарегестрирован и является инициатором
		if userExist && userData.Registered && userData.Role == Initiator {
			_, err := b.Vlt.CreateInitiative(userData.Id, update.Message.Text)
			if err != nil {
				log.Println(err)
				continue
			}

			// Отправить вопрос в главный чат на голосование
			err = b.SendPool(update.Message.Text)
			if err != nil {
				log.Println(err)
				continue
			}

			if _, err = b.Api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, questionSent)); err != nil {
				log.Println(err)
			}
		}
	}
}
