package routes

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/DSSD-Madison/gmu/db"
	"github.com/DSSD-Madison/gmu/handlers"
	"github.com/DSSD-Madison/gmu/models"
	"github.com/labstack/echo/v4"
)

func addImagesToResults(results models.KendraResults, c echo.Context, queries *db.Queries) (error) {
	uris, err := ConvertToS3URIs(results)
	if err != nil {
		return err
	}
	documentMap, err := handlers.GetDocuments(c, queries, uris)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	for key, kendraResult := range results.Results {
		s3URI := ConvertToS3URI(kendraResult.Link)
		if s3URI == "" {
			continue
		}

		document, found := documentMap[s3URI]
		if !found {
			return c.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("document not found for S3 URI: %s", s3URI)})
		}

		if document.S3FilePreview.Valid {
			image, err := ConvertS3URIToURL(document.S3FilePreview.String)
			if err != nil {
				return err
			}
			kendraResult.Image = image
		} else {
			kendraResult.Image = "https://placehold.co/120x120/webp"
		}
		results.Results[key] = kendraResult
	}
	
	return nil
}

func ConvertToS3URIs(kendraResults models.KendraResults) ([]string, error) {
	s3URIs := []string{}
	for _, result := range kendraResults.Results {
		fmt.Println(result.Link)
		s3URI := ConvertToS3URI(result.Link)
		if s3URI == "" {
			continue
		}
		s3URIs = append(s3URIs, s3URI)
	}
	return s3URIs, nil
}

func ConvertToS3URI(inputURL string) (string) {
	// Parse the URL
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		log.Printf("invalid URL: %s", inputURL)
		return ""
	}

	// Ensure the URL uses HTTP or HTTPS
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		log.Printf("unsupported URL scheme: %s", inputURL)
		return ""
	}

	// Extract the hostname (should be in the form of "<bucket>.s3.amazonaws.com")
	hostParts := strings.Split(parsedURL.Host, ".s3.amazonaws.com")
	if len(hostParts) != 2 || hostParts[0] == "" {
		log.Printf("URL format is incorrect: %s", inputURL)
		return ""
	}

	bucket := hostParts[0]

	// Decode the path (converts "%20" → " ")
	filePath, err := url.PathUnescape(parsedURL.Path)
	if err != nil {
		log.Printf("error decoding URL path: %v", err)
		return ""
	}

	// Remove leading slash if present
	filePath = strings.TrimPrefix(filePath, "/")

	// Construct the S3 URI
	return fmt.Sprintf("s3://%s/%s", bucket, filePath)
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

