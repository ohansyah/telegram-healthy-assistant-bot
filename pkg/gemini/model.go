package gemini

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/google/generative-ai-go/genai"
)

var promptPrefix = `You are an expert Nutritionist and Food Safety Officer. Your task is to analyze the provided food product data (Ingredients and/or Nutrition Facts) for safety and health implications.
`

var promptSuffix = `
**Instructions:**
1. Analyze the ingredient quality.
2. If specific nutrition numbers (grams/calories) are not provided, estimate whether they are Low/Medium/High based on the ingredient order.
3. Keep the output concise and scannable.

Please output the response using this format:

📊 Nutrition Estimation
Profile: [High Carb / High Protein / Balanced / High Fat]
Key Levels: Estimate Sugar and Sodium levels (Low/Moderate/High).

⚠️ Ingredient Analysis
Red Flags: List any ultra-processed additives, preservatives, artificial colors, or hidden sugars.
Allergens/Sensitivities: Highlight common allergens (Dairy, Soy, Nuts, Gluten, etc.).
Positive Callouts: Mention beneficial whole-food ingredients.

✅ Verdict
Rating: [Healthy / Moderate / Unhealthy]
Summary: A one-sentence conclusion on whether this should be consumed daily or strictly limited.
`

func AnalyzeText(ctx context.Context, model *genai.GenerativeModel, content string) (string, error) {

	ingredient := `
	**Input Data:**
	` + content + `
	`

	// The Prompt
	promptText := promptPrefix + ingredient + promptSuffix

	// Create the request parts: Text Prompt + Image Data
	prompt := []genai.Part{
		genai.Text(promptText),
	}

	res, err := executeContent(ctx, model, prompt)
	if err != nil {
		return "", err
	}

	return res, nil
}

func AnalyzeImage(ctx context.Context, model *genai.GenerativeModel, imgData []byte) (string, error) {
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
	promptText := promptPrefix + promptSuffix

	// Create the request parts: Text Prompt + Image Data
	prompt := []genai.Part{
		genai.ImageData(format, imgData),
		genai.Text(promptText),
	}

	res, err := executeContent(ctx, model, prompt)
	if err != nil {
		return "", err
	}

	return res, nil
}

func executeContent(ctx context.Context, model *genai.GenerativeModel, prompt []genai.Part) (string, error) {
	log.Println("Executing content generation request...")

	// Execute request
	resp, err := model.GenerateContent(ctx, prompt...)
	if err != nil {
		return "", err
	}

	log.Println("Content generation response received. Processing...")

	// Extract text from response
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		log.Println("Extracting text from response...")
		var sb strings.Builder
		for _, part := range resp.Candidates[0].Content.Parts {
			log.Println("Processing part:")
			if txt, ok := part.(genai.Text); ok {
				log.Println("Appending text part to result.")
				sb.WriteString(string(txt))
			}
		}
		log.Println("Text extraction complete.")
		return sb.String(), nil
	}

	return "No analysis returned.", nil
}
