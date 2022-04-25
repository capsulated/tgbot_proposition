package main

import (
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gitlab.com/logiq.one/agenda_3dru_bot/config"
	"gitlab.com/logiq.one/agenda_3dru_bot/telegram"
	"gitlab.com/logiq.one/agenda_3dru_bot/vault"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log := logrus.New()

	err := godotenv.Load(".env.dev")
	if err != nil {
		log.Panic(err)
	}

	cfg := config.New()

	vlt, err := vault.New(cfg)
	if err != nil {
		log.Panic(err)
	}
	defer vlt.DbPool.Close()

	bot := telegram.New(cfg)

	//poll := tgbotapi.SendPollConfig{BaseChat, "Question", []x}
	//bot.Send(poll)

	//cron := scheduler.New()
	//
	//go cron.Run(func (){
	//	log.Println(time.Now())
	//})

	//bot.CreatePool()
	go bot.Listen()

	//rs := raise.New()

	//wrk := worker.New(log, vlt, bot, rs)
	//go wrk.Work()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGABRT)
	for {
		select {
		case sig := <-sigs:
			log.Warnf("OS cmd received signal %s", sig.String())
			switch sig {
			//case syscall.SIGHUP:
			//	log.Info("Logfile rotate")
			//	break
			case syscall.SIGINT, syscall.SIGQUIT:
				log.Info("Graceful stop")
				os.Exit(1)
			case syscall.SIGABRT:
				os.Exit(1)
			}
			break
		}
	}
}
