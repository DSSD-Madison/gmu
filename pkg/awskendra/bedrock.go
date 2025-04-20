package awskendra

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/ledongthuc/pdf"
	"math"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	digitLineRegex     = regexp.MustCompile(`^\s*\d+\s*$`)
	jsonRegex          = regexp.MustCompile(`(?s)\{.*}`)
	maxPdfPagesToParse = 10
	maxTokens          = 512 // Max output tokens for Claude
	temperature        = 0.3
	topP               = 1.0
)

// BedrockClient provides methods to interact with the AWS Bedrock Runtime service.
type BedrockClient struct {
	client     *bedrockruntime.Client
	config     Config
	categories []string
	regions    []string
	keywords   []string
}

type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ClaudeRequestBody struct {
	Messages         []ClaudeMessage `json:"messages"`
	MaxTokens        int             `json:"max_tokens"`
	Temperature      float64         `json:"temperature"`
	TopP             float64         `json:"top_p"`
	AnthropicVersion string          `json:"anthropic_version"`
}

type ClaudeResponseContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ClaudeResponseBody struct {
	Content      []ClaudeResponseContent `json:"content"`
	ID           string                  `json:"id"`
	Model        string                  `json:"model"`
	Role         string                  `json:"role"`
	StopReason   string                  `json:"stop_reason"`
	StopSequence *string                 `json:"stop_sequence"` // Use pointer for optional null
	Type         string                  `json:"type"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type ExtractedMetadata struct {
	Title        string   `json:"title"`
	Abstract     string   `json:"abstract"`
	Category     string   `json:"category"`     // Note: Python uses category_name for the list, this seems like a single category? Adjust if needed.
	PublishDate  string   `json:"publish_date"` // Keep as string for simplicity, parse later if needed
	Source       string   `json:"source"`
	RegionName   []string `json:"region_name"`   // Array of strings
	KeywordName  []string `json:"keyword_name"`  // Array of strings
	AuthorName   []string `json:"author_name"`   // Array of strings
	CategoryName []string `json:"category_name"` // Array of strings
}

func loadKeywordsFromFile(filepath string) ([]string, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read keyword file: %w", err)
	}
	lines := strings.Split(string(content), "\n")
	var loadedKeywords []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			loadedKeywords = append(loadedKeywords, trimmed)
		}
	}
	return loadedKeywords, nil
}

func NewBedrockClient(cfg Config) (*BedrockClient, error) {
	opts := aws.Config{
		Region:      cfg.Region,
		Credentials: cfg.Credentials,
	}

	brClient := bedrockruntime.NewFromConfig(opts)

	categories := []string{
		"article", "background paper", "blog post", "book", "brief", "case study", "dataset", "educational guide",
		"evaluation", "fact sheet", "government report", "organizational study", "paper", "policy brief", "policy paper",
		"project evaluation", "project evaluations", "report", "working paper",
	}

	regions := []string{
		"Afghanistan", "Africa", "Albania", "Angola", "Asia", "Bangladesh", "Benin", "Bosnia And Herzegovina",
		"Burkina Faso", "Burundi", "Cambodia", "Caribean", "Central African Republic Car", "Central America",
		"Democratic Republic Of Congo Drc", "Democratic Republic Of Congo Drc / Central African Republic Car",
		"Ecuador", "Egypt", "El Salvador", "Ethiopia", "Europe", "Georgia", "Ghana", "Global", "Guatemala",
		"Guinea", "Indonesia", "Indo Pacific", "Iraq", "Israel", "Jamaica", "Jerusalem", "Jordan", "Kenya",
		"Kosovo", "Kyrgyzstan", "Latin America", "Lebanon", "Liberia", "Macedonia", "Madagascar", "Mali",
		"Middle East", "Morocco", "Myanmar", "Nepal", "Nigeria", "North America", "Oceana", "Oceania",
		"Pakistan", "Papua New Guinea", "Peru", "Philippines", "Russia", "Rwanda", "Senegal", "Somalia",
		"South Africa", "South America", "South Sudan", "Sri Lanka", "Sudan", "Tajikistan", "Tanzania",
		"Timor Leste", "Uganda", "Ukraine", "West Bank", "Yemen", "Zambia", "Zimbabwe",
	}

	// Load keywords from file
	keywords, err := loadKeywordsFromFile(cfg.KeywordsFilePath)

	if err != nil {
		return nil, fmt.Errorf("failed to load keywords from file: %w", err)
	}

	return &BedrockClient{
		client:     brClient,
		config:     cfg,
		categories: categories,
		regions:    regions,
		keywords:   keywords,
	}, nil
}

// estimateCost approximates Bedrock API call cost.
func estimateCost(inputTokens int, outputTokens int) float64 {
	inputCost := float64(inputTokens) / 1000.0 * 0.00025   // Haiku input cost
	outputCost := float64(outputTokens) / 1000.0 * 0.00125 // Haiku output cost
	totalCost := inputCost + outputCost
	// Round to 6 decimal places
	return math.Round(totalCost*1e6) / 1e6
}

// cleanText removes lines that are just numbers or too short.
func cleanText(text string) string {
	lines := strings.Split(text, "\n")
	var cleanedLines []string
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if digitLineRegex.MatchString(trimmedLine) {
			continue // Skip lines that are just numbers
		}
		if len(trimmedLine) < 6 {
			continue // Skip very short lines
		}
		cleanedLines = append(cleanedLines, trimmedLine)
	}
	return strings.Join(cleanedLines, "\n")
}

// extractTextFromPdf extracts text from the first N pages of a PDF.
func extractTextFromPdf(pdfBytes []byte, maxPages int) (string, error) {
	reader := bytes.NewReader(pdfBytes)
	pdfReader, err := pdf.NewReader(reader, int64(len(pdfBytes)))
	if err != nil {
		return "", fmt.Errorf("failed to create PDF reader: %w", err)
	}

	numPages := pdfReader.NumPage()
	if numPages == 0 {
		return "", fmt.Errorf("PDF has no pages")
	}

	pagesToRead := maxPages
	if numPages < maxPages {
		pagesToRead = numPages
	}

	var textBuilder strings.Builder
	for i := 1; i <= pagesToRead; i++ { // pdf library pages are 1-indexed
		page := pdfReader.Page(i)
		if page.V.IsNull() {
			fmt.Printf("⚠️ Warning: Skipping potentially invalid page %d", i)
			continue
		}
		content, err := page.GetPlainText(nil)
		if err != nil {
			// Log error but try to continue with other pages
			fmt.Printf("⚠️ Warning: Failed to get text from page %d: %v", i, err)
			continue
		}
		textBuilder.WriteString(content)
		textBuilder.WriteString("\n") // Add newline between pages like Python code
	}

	return textBuilder.String(), nil
}

// callClaudeHaiku sends the prompt to Bedrock and gets the response.
func (c BedrockClient) callClaudeHaiku(prompt string) (string, int, int, error) {
	requestBody := ClaudeRequestBody{
		Messages: []ClaudeMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens:        maxTokens,
		Temperature:      temperature,
		TopP:             topP,
		AnthropicVersion: "bedrock-2023-05-31",
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", 0, 0, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// inputTokens := estimateTokens(prompt) // Estimate before sending if needed

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // Add a timeout
	defer cancel()

	resp, err := c.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		Body:        jsonBody,
		ModelId:     aws.String(c.config.ModelID),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
	})
	if err != nil {
		return "", 0, 0, fmt.Errorf("failed to invoke Bedrock model: %w", err)
	}

	var responseBody ClaudeResponseBody
	err = json.Unmarshal(resp.Body, &responseBody)
	if err != nil {
		// Sometimes the response might not be perfect JSON, try raw output
		fmt.Printf("⚠️ Failed to unmarshal Bedrock JSON response: %v. Raw response: %s", err, string(resp.Body))
		// Attempt to return the raw body if unmarshalling fails but content might exist
		if len(resp.Body) > 0 {
			return string(resp.Body), 0, 0, fmt.Errorf("failed to unmarshal response, but got raw body")
		}
		return "", 0, 0, fmt.Errorf("failed to unmarshal Bedrock response body: %w", err)
	}

	if len(responseBody.Content) == 0 || responseBody.Content[0].Type != "text" {
		return "", responseBody.Usage.InputTokens, responseBody.Usage.OutputTokens, fmt.Errorf("unexpected response format from Bedrock: %s", string(resp.Body))
	}

	outputText := strings.TrimSpace(responseBody.Content[0].Text)
	inputTokens := responseBody.Usage.InputTokens
	outputTokens := responseBody.Usage.OutputTokens

	return outputText, inputTokens, outputTokens, nil
}

// extractFirstJson tries to find and parse the first JSON object in a string.
func extractFirstJson(text string) (*ExtractedMetadata, error) {
	match := jsonRegex.FindString(text)
	if match == "" {
		return nil, fmt.Errorf("no JSON object found in the text")
	}

	var metadata ExtractedMetadata
	decoder := json.NewDecoder(strings.NewReader(match))
	decoder.DisallowUnknownFields() // Optional: Be stricter about fields

	err := decoder.Decode(&metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JSON object: %w. Text was: %s", err, match)
	}
	return &metadata, nil
}

// buildPrompt constructs the prompt for Claude.
func (c BedrockClient) buildPrompt(text string) string {
	// Use fmt.Sprintf for easy formatting, similar to f-strings
	return fmt.Sprintf(`
You are an assistant extracting structured metadata from an academic policy document.

Prefer to select from the following known lists if relevant:

CATEGORIES:
%s

REGIONS:
%s

KEYWORDS:
%s

Return only a valid JSON object with the following fields:
- "title" (string, required)
- "abstract" (string)
- "category" (string, max 100 characters): e.g., article, research paper, etc.
- "publish_date" (date)
- "source" (string, max 255 characters): use "bucket" as a placeholder
- "region_name" (array of unique strings, required, max 10)
- "keyword_name" (array of unique strings, required, max 10)
- "author_name" (array of unique strings, required, max 10)
- "category_name" (array of unique strings, required, max 10)

Do not explain. Do not say "Here is the JSON". Do not use Markdown. Just return the JSON object.
TEXT:
%s
`, strings.Join(c.categories, ", "), strings.Join(c.regions, ", "), strings.Join(c.keywords, ", "), text)
}

// ProcessPdfAndExtractMetadata fetches PDF from S3, extracts text, calls LLM, and parses metadata.
func (c BedrockClient) ProcessPdfAndExtractMetadata(ctx context.Context, pdfBytes []byte) (*ExtractedMetadata, error) {
	rawText, err := extractTextFromPdf(pdfBytes, maxPdfPagesToParse)
	if err != nil {
		return nil, fmt.Errorf("failed to extract text from PDF: %w", err)
	}
	if rawText == "" {
		return nil, fmt.Errorf("no text could be extracted from the first %d pages", maxPdfPagesToParse)
	}

	cleanedText := cleanText(rawText)
	prompt := c.buildPrompt(cleanedText)

	claudeResponseText, actualInputTokens, actualOutputTokens, err := c.callClaudeHaiku(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to call Claude: %w", err)
	}
	fmt.Printf("☁️ Claude response received.")
	fmt.Printf("📊 Actual Tokens -> Input: %d, Output: %d", actualInputTokens, actualOutputTokens)
	fmt.Printf("💸 Actual cost for this doc: $%.6f", estimateCost(actualInputTokens, actualOutputTokens))

	metadata, err := extractFirstJson(claudeResponseText)
	if err != nil {
		fmt.Printf("Raw Claude response on JSON parse failure:\n%s", claudeResponseText)
		return nil, fmt.Errorf("failed to extract JSON from Claude response: %w", err)
	}

	return metadata, nil
}
