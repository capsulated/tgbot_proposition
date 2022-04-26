package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gitlab.com/logiq.one/agenda_3dru_bot/model"
	"log"
)

func (b *Bot) SendPool(question string) error {
	poll := tgbotapi.SendPollConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID: b.MainChannelId,
		},
		Question: question,
		Options:  []string{"Да", "Нет", "Архив"},
	}

	_, err := b.Api.Send(poll)
	return err
}

func (b *Bot) SendWonInitiatives(wonInitiatives *[]model.WonInitiative) error {
	secretaries, err := b.Vlt.ListSecretaries()
	if err != nil {
		return err
	}

	var questions string
	var emails string
	var empty struct{}
	e := make(map[string]struct{})
	for i, initiative := range *wonInitiatives {
		_, ok := e[initiative.Email]
		if !ok {
			e[initiative.Email] = empty
			emails += fmt.Sprintf("%s ", initiative.Email)
		}
		questions += fmt.Sprintf("%d. %s\n", i+1, initiative.Question)
	}

	for _, secretary := range *secretaries {
		if _, err = b.Api.Send(tgbotapi.NewMessage(secretary.ChatId, questions)); err != nil {
			log.Println(err)
		}
		if _, err = b.Api.Send(tgbotapi.NewMessage(secretary.ChatId, emails)); err != nil {
			log.Println(err)
		}
	}

	return nil
}
