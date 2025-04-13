package routes

import (
	"context"
	"fmt"
	"github.com/DSSD-Madison/gmu/pkg/awskendra"
	"github.com/DSSD-Madison/gmu/pkg/db"
	"github.com/lib/pq"
	"log"
	"net/http"
	"net/url"
	"strings" // For placeholder helper

	// Import UUID generator again
	"github.com/google/uuid" // Run: go get github.com/google/uuid
	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/web"
	"github.com/DSSD-Madison/gmu/web/components"
)

func (h *Handler) PDFUploadPage(c echo.Context) error {
	csrf := c.Get("csrf").(string)
	return web.Render(c, http.StatusOK, components.PDFUpload(csrf))
}

func (h *Handler) HandlePDFUpload(c echo.Context) error {
	fileHeader, err := c.FormFile("pdf")
	if err != nil {
		log.Printf("Error getting uploaded file: %v", err)
		errorMessage := fmt.Sprintf("Failed to get file: %v", err)
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML) // Ensure correct content type for HTMX swap
		return web.Render(c, http.StatusBadRequest, components.UploadResponse(false, errorMessage))
	}

	originalFilename := fileHeader.Filename
	log.Printf("Received file upload: Name=%s", originalFilename)

	// --- 2. Generate Placeholder File ID (UUID) ---
	fileId := uuid.NewString()
	log.Printf("Generated placeholder fileId (UUID): %s for original file: %s", fileId, originalFilename)

	// --- 3. Construct Redirect URL ---
	redirectPath := fmt.Sprintf("/edit-metadata/%s", fileId)
	redirectURL := url.URL{Path: redirectPath}
	query := redirectURL.Query()
	redirectURL.RawQuery = query.Encode()
	finalRedirectURL := redirectURL.String() // Get the final URL string

	c.Response().Header().Set("HX-Redirect", finalRedirectURL)
	return c.NoContent(http.StatusOK)
}

// PDFMetadataEditPage displays the form using fileId from the path parameter
func (h *Handler) PDFMetadataEditPage(c echo.Context) error {
	fileId := c.Param("fileId") // <<< Read fileId from path /edit-metadata/:fileId

	if fileId == "" {
		log.Println("Error: fileId missing in path for edit page")
		return c.Redirect(http.StatusTemporaryRedirect, "/upload")
	}
	parsedUUID, err := uuid.Parse(fileId)

	if err != nil {
		return err
	}

	doc, err := h.db.FindDocumentByID(context.Background(), parsedUUID)
	if err != nil {
		return err
	}

	res := parseDocument(doc)

	// --- 3. Render the Edit Form ---
	// Pass fileId as the primary identifier to the form component
	csrf := c.Get("csrf").(string)
	return web.Render(c, http.StatusOK, components.PDFMetadataEditForm(
		fileId, // <-- Pass fileId here
		res.Link,
		res.Title,
		res.Abstract,
		strings.Join(res.Categories, ","),
		res.PublishDate,
		res.Source,
		res.Regions,
		res.Keywords,
		res.Authors,
		csrf,
	))
}

func parseDocument(row db.FindDocumentByIDRow) awskendra.KendraResult {
	res := awskendra.KendraResult{}

	res.Title = row.Title

	parts := strings.Split(row.S3File, "/")
	if len(parts) >= 2 {
		res.Link = parts[3]
	}

	if row.PublishDate.Valid {
		res.PublishDate = row.PublishDate.Time.Format("2006-01-02")
	}
	if row.Abstract.Valid {
		res.Abstract = row.Abstract.String
	}
	if row.Source.Valid {
		res.Source = row.Source.String
	}
	var tempScanner pq.StringArray
	err := tempScanner.Scan(row.AuthorNames.(string))
	if err == nil {
		res.Authors = tempScanner
	}

	err = tempScanner.Scan(row.CategoryNames.(string))
	if err == nil {
		res.Categories = tempScanner
	}

	err = tempScanner.Scan(row.KeywordNames.(string))
	if err == nil {
		res.Keywords = tempScanner
	}

	err = tempScanner.Scan(row.RegionNames.(string))
	if err == nil {
		res.Regions = tempScanner
	}

	return res
}

// HandleMetadataSave processes the submitted metadata form (Placeholder)
func (h *Handler) HandleMetadataSave(c echo.Context) error {
	// --- 1. Get submitted form data ---
	// Get the fileId from the hidden form field
	fileId := c.FormValue("fileId") // <<< Should match hidden input name="fileId"
	title := c.FormValue("title")
	//abstract := c.FormValue("abstract")
	//category := c.FormValue("category")
	//publishDate := c.FormValue("publish_date")
	//source := c.FormValue("source")
	//regionNames := c.FormValue("region_names")
	//keywordNames := c.FormValue("keyword_names")
	authorNames := c.FormValue("author_names")

	// Basic validation
	if fileId == "" {
		log.Println("Error: fileId missing in metadata save form submission")
		return c.Redirect(http.StatusSeeOther, "/upload") // Redirect back
	}

	// Log the received values associated with the fileId
	log.Printf("Received metadata update for fileId '%s':", fileId) // <<< Log against fileId
	log.Printf("  Title: %s", title)
	// ... other fields ...
	log.Printf("  Author Names: %s", authorNames)

	// --- TODO: Implement Actual Saving Logic ---
	// 1. Use `fileId` (UUID) to identify the record/file in your database/system.
	// 2. Validate data.
	// 3. Split comma-separated strings.
	// 4. Perform database operations (UPDATE record identified by `fileId`, etc.).
	// ------------------------------------------

	log.Printf("Placeholder save complete for fileId '%s'. Redirecting.", fileId)
	return c.Redirect(http.StatusSeeOther, "/upload")
}
