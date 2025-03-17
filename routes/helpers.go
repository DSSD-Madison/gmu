package routes

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/DSSD-Madison/gmu/models"
)

func ConvertToS3URIs(kendraResults models.KendraResults) ([]string, error) {
	fmt.Println("before conversion")
	
	s3URIs := []string{}
	for _, result := range kendraResults.Results {
		fmt.Println(result.Link)
		s3URI, err := ConvertToS3URI(result.Link)
		if err != nil {
			return nil, err
		}
		s3URIs = append(s3URIs, s3URI)
	}
	return s3URIs, nil
}

func ConvertToS3URI(inputURL string) (string, error) {
	// Parse the URL
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %s", inputURL)
	}

	// Ensure the URL uses HTTP or HTTPS
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", fmt.Errorf("unsupported URL scheme: %s", inputURL)
	}

	// Extract the hostname (should be in the form of "<bucket>.s3.amazonaws.com")
	hostParts := strings.Split(parsedURL.Host, ".s3.amazonaws.com")
	if len(hostParts) != 2 || hostParts[0] == "" {
		return "", fmt.Errorf("URL format is incorrect: %s", inputURL)
	}

	bucket := hostParts[0]

	// Decode the path (converts "%20" → " ")
	filePath, err := url.PathUnescape(parsedURL.Path)
	if err != nil {
		return "", fmt.Errorf("error decoding URL path: %v", err)
	}

	// Remove leading slash if present
	filePath = strings.TrimPrefix(filePath, "/")

	// Construct the S3 URI
	return fmt.Sprintf("s3://%s/%s", bucket, filePath), nil
}

func ConvertS3URIToURL(s3URI string) (string, error) {
	if !strings.HasPrefix(s3URI, "s3://") {
		return "", fmt.Errorf("invalid S3 URI: %s", s3URI)
	}

	uriParts := strings.TrimPrefix(s3URI, "s3://")


	parts := strings.SplitN(uriParts, "/", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("S3 URI format is incorrect: %s", s3URI)
	}

	bucket := parts[0]
	filePath := parts[1]

	// Encode the file path (spaces → %20, special chars encoded)
	encodedPath := url.PathEscape(filePath)

	// Construct the HTTPS URL
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucket, encodedPath), nil
}