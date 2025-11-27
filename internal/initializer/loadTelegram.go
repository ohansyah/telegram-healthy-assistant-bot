package initializer

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ohansyah.com/telegram-healthy-assistant-bot/config"
)

var BotTelegram *tgbotapi.BotAPI

func InitTelegram() {
	// INIT: Telegram Bot
	var err error
	BotTelegram, err = tgbotapi.NewBotAPI(config.Get().Telegram.Token)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}
	log.Printf("Bot authorized as %s", BotTelegram.Self.UserName)

}
