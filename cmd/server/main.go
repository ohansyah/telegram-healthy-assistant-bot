package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"ohansyah.com/telegram-healthy-assistant-bot/config"
	"ohansyah.com/telegram-healthy-assistant-bot/internal/initializer"
	"ohansyah.com/telegram-healthy-assistant-bot/pkg/gemini"
)

func init() {
	initializer.LoadEnv()
	initializer.InitTelegram()
}

func main() {

	// INIT: Gemini Client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(config.Get().Gemini.Key))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel(config.Get().Gemini.Model)

	// POLLING: Start listening for updates
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := initializer.BotTelegram.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID

		// --- Text Message Handling ---
		if update.Message.Text != "" {

			// --- SEND TO GEMINI ---
			logAndSend(initializer.BotTelegram, chatID, "Start analysis...")
			analysisResult, err := gemini.AnalyzeText(ctx, model, update.Message.Text)
			if err != nil {
				logAndSend(initializer.BotTelegram, chatID, "Error during analysis: "+err.Error())
				continue
			}

			// -- SEND RESULT BACK TO USER ---
			reply := tgbotapi.NewMessage(chatID, analysisResult)
			reply.ParseMode = "Markdown"
			initializer.BotTelegram.Send(reply)
		}

		// --- Image Message Handling ---
		if len(update.Message.Photo) > 0 {
			logAndSend(initializer.BotTelegram, chatID, "Photo received, Start Processing!")

			photo := update.Message.Photo[len(update.Message.Photo)-1]
			fileConfig, err := initializer.BotTelegram.GetFile(tgbotapi.FileConfig{FileID: photo.FileID})
			if err != nil {
				logAndSend(initializer.BotTelegram, chatID, err.Error())
				continue
			}

			// --- Download the file ---
			// Note: fileConfig.Link() generates the URL: https://api.telegram.org/file/bot<token>/<file_path>
			downloadURL := fileConfig.Link(config.Get().Telegram.Token)
			resp, err := http.Get(downloadURL)
			if err != nil {
				logAndSend(initializer.BotTelegram, chatID, "Error download: "+err.Error())
				continue
			}

			imgData, err := io.ReadAll(resp.Body)
			if err != nil {
				logAndSend(initializer.BotTelegram, chatID, "Error Read Image "+err.Error())
				continue
			}

			// --- SEND TO GEMINI ---
			logAndSend(initializer.BotTelegram, chatID, "Start analysis...")
			analysisResult, err := gemini.AnalyzeImage(ctx, model, imgData)
			if err != nil {
				logAndSend(initializer.BotTelegram, chatID, "Error during analysis: "+err.Error())
				continue
			}

			// -- SEND RESULT BACK TO USER ---
			log.Println("Sending analysis result to user...")
			reply := tgbotapi.NewMessage(chatID, analysisResult)
			reply.ParseMode = "Markdown"
			msg, err := initializer.BotTelegram.Send(reply)
			if err != nil {
				log.Println("Error sending message: ", err)
			} else {
				log.Println("Message sent successfully: ", msg.MessageID)
			}
		}

		// avoid overload if many updates
		time.Sleep(3 * time.Second)
	}
}

// create reusable function for logging & sending messages
func logAndSend(bot *tgbotapi.BotAPI, chatID int64, message string) {
	log.Println(message)
	msg := tgbotapi.NewMessage(chatID, message)
	bot.Send(msg)
}
