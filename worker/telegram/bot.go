package telegram

import (
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gitlab.com/logiq.one/agenda_3dru_bot/config"
	"log"
	"net/mail"
	"sync"
)

const commandStart = "start"

const pinAsk = "Введите пин-код"
const alreadyRegistered = "Вы уже зарегестрированы!"
const incorrectPin = "Неверный пинкод, повторите попытку:"
const inputEmail = "Введите адрес электронной почты"
const incorrectEmail = "Неверный адрес электронной почты!\n"
const initiatorRegistered = "Вы зарегистрированы. Введите Ваш вопрос на повестку планёрки одним сообщением. Приём вопросов осуществляется каждую неделю с понедельника по четверг"
const secretaryRegistered = "Вы зарегистрированы. Подверждённые вопросы на планёрку Вам будут отправлены каждую пятницу в 10.00"

type Bot struct {
	Users         sync.Map
	PinInitiator  string
	PinSecretary  string
	MainChannelId int64
	Api           *tgbotapi.BotAPI
	Updates       tgbotapi.UpdatesChannel
}

type UserData struct {
	ChatId      int64
	Role        Role
	PrevMessage string
	Email       string
	Registered  bool
}

type Role int32

const (
	Initiator Role = iota + 1
	Secretary
)

func New(cfg *config.App) *Bot {
	api, err := tgbotapi.NewBotAPI(cfg.Telegram.ApiToken)
	if err != nil {
		log.Fatalf("%v", err)
	}

	api.Debug = cfg.Telegram.Debug

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// TODO Выберем всех юзеров из базы!

	return &Bot{
		MainChannelId: cfg.Telegram.MainChannelId,
		Api:           api,
		PinInitiator:  cfg.Pin.Initiator,
		PinSecretary:  cfg.Pin.Secretary,
		Updates:       api.GetUpdatesChan(u),
	}
}

//func (b *Bot) SendMsg(msg string) {
//
//}

//func (b *Bot) CreatePool() {
//	poll := tgbotapi.SendPollConfig{
//		BaseChat: tgbotapi.BaseChat{
//			ChatID: b.MainChannelId,
//		},
//		Question:        "Чё каво, беяч?",
//		Type:            "quiz",
//		Options:         []string{"его", "того", "этого", "хз"},
//		CorrectOptionID: 1,
//	}
//
//	msg, err := b.Api.Send(poll)
//	if err != nil {
//		log.Println(err)
//	} else {
//		log.Println(msg)
//	}
//}

func (b *Bot) Listen() {
	for update := range b.Updates {
		if update.Message == nil {
			continue
		}

		// Если сообщение из главного (mainChannelId) канала - значит это голосование
		// И атор - не бот
		if update.Message.Chat.ID == b.MainChannelId {
			b.polling(&update)
			continue
		}

		username := update.Message.From.UserName
		u, userExist := b.Users.Load(username)
		var userData UserData
		if userExist {
			log.Println("USER-EXIST ", userExist)
			log.Println("USER-DATA ", userData)
			userData = u.(UserData)
		}

		// Если это команда /start и юзер существует и зарегестрирован
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
				ChatId:      update.Message.Chat.ID,
				PrevMessage: text,
			}
			b.Users.Store(username, userData)
			log.Printf("%s STORED: %v", username, userData)

			if _, err := b.Api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, text)); err != nil {
				log.Println(err)
			}

			continue
		}

		// Юзер существует, но не зарегестрирован, а предыдущее сообщение было запросом ПИН-кода (или повторного ввода пин-кода)
		if userExist && !userData.Registered && (userData.PrevMessage == pinAsk || userData.PrevMessage == incorrectPin) {
			var text string

			log.Println("update.Message.Text", update.Message.Text)
			log.Println("b.PinInitiator", b.PinInitiator)
			log.Println("b.PinSecretary", b.PinSecretary)
			log.Println("update.Message.Text == b.PinSecretary", update.Message.Text == b.PinSecretary)
			log.Println("update.Message.Text == b.PinInitiator", update.Message.Text == b.PinInitiator)

			if update.Message.Text == b.PinInitiator || update.Message.Text == b.PinSecretary {
				text = inputEmail

				var role Role = Initiator
				if update.Message.Text == b.PinSecretary {
					role = Secretary
				}

				// Обновим данные юзера
				b.Users.Store(
					username,
					UserData{
						ChatId:      update.Message.Chat.ID,
						Role:        role,
						PrevMessage: text,
					},
				)
			} else {
				text = incorrectPin
			}

			if _, err := b.Api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, text)); err != nil {
				log.Println(err)
			}
		}

		// Юзер существует, но не зарегестрирован, а предыдущее сообщение было запросом емэйла
		if userExist && !userData.Registered && (userData.PrevMessage == inputEmail) {
			var text string

			_, err := mail.ParseAddress(update.Message.Text)
			if err != nil {
				text = incorrectEmail + err.Error()
			} else if userData.Role == Initiator {
				text = initiatorRegistered
			} else if userData.Role == Secretary {
				text = secretaryRegistered
			}

			// Обновим данные юзера
			b.Users.Store(
				username,
				UserData{
					ChatId:      update.Message.Chat.ID,
					Role:        userData.Role,
					PrevMessage: text,
					Email:       update.Message.Text,
					Registered:  true,
				},
			)

			// TODO запишем в базу!!!

			if _, err = b.Api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, text)); err != nil {
				log.Println(err)
			}
		}

		// Юзер существует, зарегестрирован и является инициатором
		if userExist && userData.Registered && userData.Role == Initiator {
			// Записать вопрос в БД
			// Отправить вопрос в главный чат на голосование
			log.Println("Записываю в БД и отправляю на голосование")
		}
	}
}

func (b *Bot) register(update *tgbotapi.Update) {
	// Chek user map

	// If exist - do nothing

	// if not Ask PinCode

	// user, chatID:

	// update.Message.From.user.String()
	// update.Message.Chat.ID

	// Registration

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, pinAsk)
	msg.ReplyToMessageID = update.Message.MessageID

	m, err := b.Api.Send(msg)
	if err != nil {
		log.Println(err)
	} else {
		log.Println(m)
	}

	// Check pin code
}

func (b *Bot) polling(update *tgbotapi.Update) {
	//log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
	log.Println("...polling...")
}
