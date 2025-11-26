package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"ohansyah.com/telegram-healthy-assistant-bot/config"
	"ohansyah.com/telegram-healthy-assistant-bot/internal/initializer"
)

func init() {
	initializer.LoadEnv()
}

func main() {
	// INIT: Telegram Bot
	bot, err := tgbotapi.NewBotAPI(config.Get().Telegram.Token)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}
	log.Printf("Bot authorized as %s", bot.Self.UserName)

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

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID

		// --- Text Message Handling ---
		if update.Message.Text != "" {
			reply := tgbotapi.NewMessage(chatID, "Received: "+update.Message.Text)
			bot.Send(reply)
			log.Printf("Received message from %s: %s", update.Message.From.UserName, update.Message.Text)
		}

		// --- Image Message Handling ---
		if len(update.Message.Photo) > 0 {
			logAndSend(bot, chatID, "Photo received, Start Processing!")

			photo := update.Message.Photo[len(update.Message.Photo)-1]
			fileConfig, err := bot.GetFile(tgbotapi.FileConfig{FileID: photo.FileID})
			if err != nil {
				logAndSend(bot, chatID, err.Error())
				continue
			}

			// --- Download the file ---
			// Note: fileConfig.Link() generates the URL: https://api.telegram.org/file/bot<token>/<file_path>
			downloadURL := fileConfig.Link(config.Get().Telegram.Token)
			resp, err := http.Get(downloadURL)
			if err != nil {
				logAndSend(bot, chatID, "Error download: "+err.Error())
				continue
			}

			imgData, err := io.ReadAll(resp.Body)
			if err != nil {
				logAndSend(bot, chatID, "Error Read Image "+err.Error())
				continue
			}

			// --- SEND TO GEMINI ---
			logAndSend(bot, chatID, "Start analysis...")
			analysisResult, err := analyzeWithGemini(ctx, model, imgData)
			if err != nil {
				logAndSend(bot, chatID, "Error during analysis: "+err.Error())
				continue
			}

			// -- SEND RESULT BACK TO USER ---
			reply := tgbotapi.NewMessage(chatID, analysisResult)
			reply.ParseMode = "Markdown"
			bot.Send(reply)
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

// analyzeWithGemini handles the prompt engineering and API call
func analyzeWithGemini(ctx context.Context, model *genai.GenerativeModel, imgData []byte) (string, error) {
	// Detect MIME type (Telegram usually sends JPEGs, but good to be safe)
	mimeType := http.DetectContentType(imgData)

	// Extract just the format (e.g. "jpeg") because genai.ImageData prepends "image/"
	// passing "image/jpeg" results in "image/image/jpeg" which causes Error 400
	var format string
	parts := strings.Split(mimeType, "/")
	if len(parts) > 1 {
		format = strings.Split(parts[1], ";")[0]
	} else {
		format = "jpeg" // Default fallback
	}

	// The Prompt
	promptText := `
	You are an expert Nutritionist and Food Safety Officer. 
	Analyze the provided image of a product label (Nutrition Facts / Informasi Nilai Gizi).
	
	Please output the response using this format:

	**📊 Nutrition Summary**
	* List the key stats (Calories, Fat, Sugar, Sodium/Salt, Protein).
	* Standardize unit to "per serving" if possible.

	**⚠️ Ingredient Analysis**
	* Read the ingredients list if visible.
	* Identify any "Malicious" or potentially harmful ingredients (e.g., High Fructose Corn Syrup, Trans Fats, Artificial Colors like Red 40, excessive preservatives).
	* If the product contains high sugar or sodium, flag it here.

	**✅ Verdict**
	* Give a one-sentence summary: Is this healthy or should it be consumed in moderation?
	`

	// Create the request parts: Text Prompt + Image Data
	prompt := []genai.Part{
		genai.ImageData(format, imgData),
		genai.Text(promptText),
	}

	// Execute request
	resp, err := model.GenerateContent(ctx, prompt...)
	if err != nil {
		return "", err
	}

	// Extract text from response
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		var sb strings.Builder
		for _, part := range resp.Candidates[0].Content.Parts {
			if txt, ok := part.(genai.Text); ok {
				sb.WriteString(string(txt))
			}
		}
		return sb.String(), nil
	}

	return "No analysis returned.", nil
}
