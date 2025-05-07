package util

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/lib/pq"

	"github.com/DSSD-Madison/gmu/internal/domain/search"
	"github.com/DSSD-Madison/gmu/internal/infra/database/handlers"
	db "github.com/DSSD-Madison/gmu/internal/infra/database/sqlc/generated"
)

func AddImagesToResults(ctx context.Context, results search.Results, queries *db.Queries) error {
	uris := ConvertToS3URIs(results)
	documentMap, err := handlers.GetDocuments(ctx, queries, uris)
	if err != nil {
		log.Printf("GetDocuments failed: %v\n", err)
	}

	for key, kendraResult := range results.Results {
		s3URI := ConvertToS3URI(kendraResult.Link)
		if s3URI == "" {
			continue
		}

		document, found := documentMap[s3URI]
		if !found {
			return fmt.Errorf("failed to find document")
		}

		if document.S3FilePreview.Valid {
			image := ConvertS3URIToURL(document.S3FilePreview.String)
			if image == "" {
				return fmt.Errorf("failed to find document")
			}
			kendraResult.Image = image
		} else {
			kendraResult.Image = "https://placehold.co/120x120/webp"
		}
		kendraResult.Abstract = document.Abstract.String
		if document.PublishDate.Valid {
			kendraResult.PublishDate = document.PublishDate.Time.Format("2006-01-02")
		}

		kendraResult.UUID = document.ID.String()

		var tempScanner pq.StringArray
		err := tempScanner.Scan(document.AuthorNames.(string))
		if err == nil {
			kendraResult.Authors = tempScanner
		}

		err = tempScanner.Scan(document.CategoryNames.(string))
		if err == nil {
			kendraResult.Categories = tempScanner
		}

		err = tempScanner.Scan(document.KeywordNames.(string))
		if err == nil {
			kendraResult.Keywords = tempScanner
		}

		err = tempScanner.Scan(document.RegionNames.(string))
		if err == nil {
			kendraResult.Regions = tempScanner
		}

		results.Results[key] = kendraResult
	}
	return nil
}

func ConvertToS3URIs(kendraResults search.Results) []string {
	s3URIs := []string{}
	for _, result := range kendraResults.Results {
		s3URI := ConvertToS3URI(result.Link)
		if s3URI == "" {
			continue
		}
		s3URIs = append(s3URIs, s3URI)
	}
	return s3URIs
}

func ConvertToS3URI(inputURL string) string {
	// Parse the URL
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		log.Printf("invalid URL: %s", inputURL)
		return ""
	}

	// Ensure the URL has a valid prefix
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

func ConvertS3URIToURL(s3URI string) string {
	if !strings.HasPrefix(s3URI, "s3://") {
		return ""
	}

	uriParts := strings.TrimPrefix(s3URI, "s3://")

	parts := strings.SplitN(uriParts, "/", 2)
	if len(parts) != 2 {
		return ""
	}

	bucket := parts[0]
	filePath := parts[1]

	// Encode the file path (spaces → %20, special chars encoded)
	encodedPath := url.PathEscape(filePath)

	// Construct the HTTPS URL
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucket, encodedPath)
}
