package main

import (
	"gitlab.com/logiq.one/agenda_3dru_bot/config"
	"gitlab.com/logiq.one/agenda_3dru_bot/scheduler"
	"gitlab.com/logiq.one/agenda_3dru_bot/telegram"
	"gitlab.com/logiq.one/agenda_3dru_bot/vault"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.SetFlags(log.Lshortfile)

	cfg := config.New()

	vlt, err := vault.New(cfg)
	if err != nil {
		log.Panic(err)
	}

	bot := telegram.New(cfg, vlt)
	go bot.Listen()

	cron := scheduler.New(cfg, vlt, bot)
	go cron.Run()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGABRT)
	for {
		select {
		case sig := <-sigs:
			log.Printf("OS cmd received signal %s", sig.String())
			switch sig {
			//case syscall.SIGHUP:
			//	log.Info("Logfile rotate")
			//	break
			case syscall.SIGINT, syscall.SIGQUIT:
				log.Println("Graceful stop")
				os.Exit(1)
			case syscall.SIGABRT:
				os.Exit(1)
			}
			break
		}
	}
}
