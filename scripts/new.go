package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime" // Assuming this is the package for Bedrock
)

func ExtractTextFromPDFPages(pdfPath string, firstPage int, lastPage int) (string, error) {
	// --- Input Validation ---
	if firstPage < 1 {
		return "", fmt.Errorf("firstPage must be 1 or greater, got %d", firstPage)
	}
	if lastPage > 0 && lastPage < firstPage {
		return "", fmt.Errorf("lastPage (%d) cannot be less than firstPage (%d)", lastPage, firstPage)
	}

	// --- Check for pdftotext ---
	toolPath, err := exec.LookPath("pdftotext")
	if err != nil {
		return "", fmt.Errorf("pdftotext command not found in PATH: %w. Please install poppler-utils", err)
	}

	// --- Prepare Command Arguments ---
	cleanPath := filepath.Clean(pdfPath)
	args := []string{}

	// Add page range flags
	args = append(args, "-f", fmt.Sprintf("%d", firstPage)) // Always specify first page

	// Only add last page if it's specified positively
	if lastPage > 0 {
		args = append(args, "-l", fmt.Sprintf("%d", lastPage))
	}
	// else: pdftotext defaults to the end of the document if -l is omitted

	// Add input file path and output specifier ('-' for stdout)
	args = append(args, cleanPath, "-")

	// --- Execute Command ---
	cmd := exec.Command(toolPath, args...) // Pass args slice

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Printf("Running command: %s %s\n", toolPath, strings.Join(args, " ")) // Log the command

	err = cmd.Run()
	if err != nil {
		stderrString := stderr.String()
		// Provide more context if pdftotext exited with an error
		if stderrString == "" {
			return "", fmt.Errorf("pdftotext failed for %q (pages %d-%d): %w", cleanPath, firstPage, lastPage, err)
		}
		return "", fmt.Errorf("pdftotext failed for %q (pages %d-%d): %w\nStderr: %s", cleanPath, firstPage, lastPage, err, stderrString)
	}

	// Return the captured standard output as a string
	return strings.TrimSpace(stdout.String()), nil
}

type TextGenerationConfig struct {
	MaxTokenCount int `json:"maxTokenCount,omitempty"`
}

func main() {
	// Load the AWS configuration from default locations
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Create a new Bedrock client
	client := bedrockruntime.NewFromConfig(cfg)

	pdfText, err := ExtractTextFromPDFPages("../../Downloads/0010/0011.pdf", 1, 5)

	// Create Titan Text Struct
	type TitanTextInput struct {
		InputText            string                `json:"inputText"`
		TextGenerationConfig *TextGenerationConfig `json:"textGenerationConfig,omitempty"`
	}
	prompt := "Find the information from the following PDF and place it in this JSON format Here is the metadata we need\n- `title` (TEXT, Not Null)\n- `abstract` (TEXT)\n- `category` (VARCHAR(100)) # article, research paper, etc…. \n- `publish_date` (DATE)\n- `source` (VARCHAR(255)) # bucket (don’t worry too much about this)\n- `region_name (VARCHAR(255), Unique, Not Null) # there are multiple\n- `keyword_name` (VARCHAR(255), Unique, Not Null) # there are multiple\n- `author_name` (VARCHAR(255), Unique, Not Null) # there are multiple \n Here is the PDF: \n " + pdfText
	//prompt := "Summarize this text: \n" + pdfText
	titanPayload := TitanTextInput{
		InputText: prompt,
		TextGenerationConfig: &TextGenerationConfig{
			MaxTokenCount: 512,
		},
	}
	bodyBytes, err := json.Marshal(titanPayload)
	// Prepare your model invocation input (example)
	input := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String("amazon.titan-text-lite-v1"), // Replace with your model ID
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
		Body:        bodyBytes,
	}

	if err != nil {
		// Handle marshalling error
		log.Fatalf("failed to marshal input payload: %v", err)
	}

	log.Printf("DEBUG: Sending Body: %s\n", string(bodyBytes))
	// Make the API call to invoke the model
	output, err := client.InvokeModel(context.TODO(), input)
	if err != nil {
		log.Fatalf("failed to invoke model: %v", err)
	}

	// Assuming the response struct has fields like OutputText or similar
	if output != nil {
		// The response might contain the output text or other relevant data
		fmt.Printf("Model Response: %s\n", string(output.Body)) // Example of accessing the output
	}
}
