package routes

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/DSSD-Madison/gmu/pkg/awskendra"
	"github.com/DSSD-Madison/gmu/pkg/db"
	"github.com/DSSD-Madison/gmu/web"
	"github.com/DSSD-Madison/gmu/web/components"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	// "github.com/lib/pq"
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
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
		return web.Render(c, http.StatusBadRequest, components.UploadResponse(false, errorMessage))
	}

	originalFilename := fileHeader.Filename
	fileID := uuid.New()

	s3Path := fmt.Sprintf("s3://your-bucket-name/documents/%s", originalFilename)
	title := strings.TrimSuffix(originalFilename, path.Ext(originalFilename))

	err = h.db.InsertUploadedDocument(c.Request().Context(), db.InsertUploadedDocumentParams{
		ID:       fileID,
		S3File:   s3Path,
		FileName: originalFilename,
		Title:    title,
	})
	if err != nil {
		log.Printf("DB insert failed: %v", err)
		errorMessage := "Could not save file metadata to database"
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
		return web.Render(c, http.StatusInternalServerError, components.UploadResponse(false, errorMessage))
	}

	redirectPath := fmt.Sprintf("/edit-metadata/%s", fileID.String())
	c.Response().Header().Set("HX-Redirect", redirectPath)
	return c.NoContent(http.StatusOK)
}

func (h *Handler) PDFMetadataEditPage(c echo.Context) error {
	fileId := c.Param("fileId")
	if fileId == "" {
		log.Println("Missing fileId")
		return c.Redirect(http.StatusTemporaryRedirect, "/upload")
	}

	docUUID, err := uuid.Parse(fileId)
	if err != nil {
		return err
	}

	// 1. Get doc info
	doc, err := h.db.FindDocumentByID(context.Background(), docUUID)
	if err != nil {
		return err
	}

	// 2. Get all available values
	allAuthors, _ := h.db.ListAllAuthors(c.Request().Context())
	allKeywords, _ := h.db.ListAllKeywords(c.Request().Context())
	allRegions, _ := h.db.ListAllRegions(c.Request().Context())
	allCategories, _ := h.db.ListAllCategories(c.Request().Context())


	authorNames := []string(doc.AuthorNames)
	

	keywordNames := []string(doc.KeywordNames)
	

	regionNames := []string(doc.RegionNames)


	categoryNames := []string(doc.CategoryNames)
	

	// 4. Convert selected values to []Pair
	selectedAuthors := toAuthorPairs(allAuthors, authorNames)
	selectedKeywords := toKeywordPairs(allKeywords, keywordNames)
	selectedRegions := toRegionPairs(allRegions, regionNames)

	csrf := c.Get("csrf").(string)

	// 5. Render form
	return web.Render(c, http.StatusOK, components.PDFMetadataEditForm(
		fileId,
		doc.FileName,
		doc.Title,
		doc.Abstract.String,
		strings.Join(categoryNames, ","),
		doc.PublishDate.Time.Format("2006-01-02"),
		doc.Source.String,
		selectedRegions,
		selectedKeywords,
		selectedAuthors,
		csrf,
		allRegions,
		allKeywords,
		allAuthors,
		allCategories,
	))
}


func toAuthorPairs(all []db.Author, selected []string) []components.Pair {
	var out []components.Pair
	for _, name := range selected {
		for _, item := range all {
			if strings.EqualFold(item.Name, name) {
				out = append(out, components.Pair{ID: item.ID.String(), Name: item.Name})
			}
		}
	}
	return out
}

func toKeywordPairs(all []db.Keyword, selected []string) []components.Pair {
	var out []components.Pair
	for _, name := range selected {
		for _, item := range all {
			if strings.EqualFold(item.Name, name) {
				out = append(out, components.Pair{ID: item.ID.String(), Name: item.Name})
			}
		}
	}
	return out
}

func toRegionPairs(all []db.Region, selected []string) []components.Pair {
	var out []components.Pair
	for _, name := range selected {
		for _, item := range all {
			if strings.EqualFold(item.Name, name) {
				out = append(out, components.Pair{ID: item.ID.String(), Name: item.Name})
			}
		}
	}
	return out
}

func toCategoryPairs(all []db.Category, selected []string) []components.Pair {
	var out []components.Pair
	for _, name := range selected {
		for _, item := range all {
			if strings.EqualFold(item.Name, name) {
				out = append(out, components.Pair{ID: item.ID.String(), Name: item.Name})
			}
		}
	}
	return out
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

	// var tempScanner pq.StringArray

	// if err := tempScanner.Scan(row.AuthorNames.(string)); err == nil {
	// 	res.Authors = tempScanner
	// }
	// if err := tempScanner.Scan(row.CategoryNames.(string)); err == nil {
	// 	res.Categories = tempScanner
	// }
	// if err := tempScanner.Scan(row.KeywordNames.(string)); err == nil {
	// 	res.Keywords = tempScanner
	// }
	// if err := tempScanner.Scan(row.RegionNames.(string)); err == nil {
	// 	res.Regions = tempScanner
	// }
	res.Authors = row.AuthorNames
	res.Categories = row.CategoryNames
	res.Keywords = row.KeywordNames
	res.Regions = row.RegionNames

	return res
}


func (h *Handler) HandleMetadataSave(c echo.Context) error {
	ctx := c.Request().Context()

	// --- Get form values ---
	fileId := c.FormValue("fileId")
	title := c.FormValue("title")
	abstract := c.FormValue("abstract")
	publishDate := c.FormValue("publish_date")
	source := c.FormValue("source")
	authorStr := c.FormValue("author_names")
	keywordStr := c.FormValue("keyword_names")
	categoryStr := c.FormValue("category_names")
	regionStr := c.FormValue("region_names")

	// --- Parse UUID ---
	docID, err := uuid.Parse(fileId)
	if err != nil {
		log.Printf("[ERROR] Invalid UUID in form: %v", err)
		return c.Redirect(http.StatusSeeOther, "/upload")
	}

	// --- Parse date ---
	var parsedDate sql.NullTime
	if publishDate != "" {
		t, err := time.Parse("2006-01-02", publishDate)
		if err == nil {
			parsedDate = sql.NullTime{Time: t, Valid: true}
		}
	}

	// --- Update main document ---
	err = h.db.UpdateDocumentMetadata(ctx, db.UpdateDocumentMetadataParams{
		ID:          docID,
		Title:       title,
		Abstract:    sql.NullString{String: abstract, Valid: abstract != ""},
		PublishDate: parsedDate,
		Source:      sql.NullString{String: source, Valid: source != ""},
	})
	if err != nil {
		log.Printf("[ERROR] Error updating document metadata: %v", err)
		return c.Redirect(http.StatusSeeOther, "/upload")
	}
	documentID := uuid.NullUUID{UUID: docID, Valid: true}
	// --- Clear previous associations ---
	h.db.DeleteDocAuthorsByDocID(ctx, documentID)
	h.db.DeleteDocKeywordsByDocID(ctx, documentID)
	h.db.DeleteDocCategoriesByDocID(ctx, documentID)
	h.db.DeleteDocRegionsByDocID(ctx, documentID)

	// --- Step 1: Parse UUIDs from comma-separated form fields ---
	parseUUIDList := func(s string) []uuid.UUID {
		ids := []uuid.UUID{}
		for _, id := range strings.Split(s, ",") {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			u, err := uuid.Parse(id)
			if err == nil {
				ids = append(ids, u)
			} else {
				log.Printf("[WARN] Skipping invalid UUID: %s", id)
			}
		}
		return ids
	}

	authors := parseUUIDList(authorStr)
	keywords := parseUUIDList(keywordStr)
	categories := parseUUIDList(categoryStr)
	regions := parseUUIDList(regionStr)

	log.Printf("[DEBUG] Parsed author IDs: %v", authors)
	log.Printf("[DEBUG] Parsed keyword IDs: %v", keywords)
	log.Printf("[DEBUG] Parsed category IDs: %v", categories)
	log.Printf("[DEBUG] Parsed region IDs: %v", regions)

	// --- Insert new associations ---
	for _, authorID := range authors {
		log.Printf("[DEBUG] Inserting doc_author: doc_id=%s author_id=%s", docID, authorID)
		err := h.db.InsertDocAuthor(ctx, db.InsertDocAuthorParams{
			ID:       uuid.New(),
			DocID:    documentID,
			AuthorID: uuid.NullUUID{UUID: authorID, Valid: true},
		})
		if err != nil {
			log.Printf("[ERROR] Failed to insert into doc_authors: %v", err)
		}
	}

	for _, keywordID := range keywords {
		log.Printf("[DEBUG] Inserting doc_keyword: doc_id=%s keyword_id=%s", docID, keywordID)
		err := h.db.InsertDocKeyword(ctx, db.InsertDocKeywordParams{
			ID:        uuid.New(),
			DocID:     documentID,
			KeywordID: uuid.NullUUID{UUID: keywordID, Valid: true},
		})
		if err != nil {
			log.Printf("[ERROR] Failed to insert into doc_keywords: %v", err)
		}
	}

	for _, categoryID := range categories {
		log.Printf("[DEBUG] Inserting doc_category: doc_id=%s category_id=%s", docID, categoryID)
		err := h.db.InsertDocCategory(ctx, db.InsertDocCategoryParams{
			ID:         uuid.New(),
			DocID:      documentID,
			CategoryID: uuid.NullUUID{UUID: categoryID, Valid: true},
		})
		if err != nil {
			log.Printf("[ERROR] Failed to insert into doc_categories: %v", err)
		}
	}

	for _, regionID := range regions {
		log.Printf("[DEBUG] Inserting doc_region: doc_id=%s region_id=%s", docID, regionID)
		err := h.db.InsertDocRegion(ctx, db.InsertDocRegionParams{
			ID:       uuid.New(),
			DocID:    documentID,
			RegionID: uuid.NullUUID{UUID: regionID, Valid: true},
		})
		if err != nil {
			log.Printf("[ERROR] Failed to insert into doc_regions: %v", err)
		}
	}

	log.Printf("[INFO] Metadata updated successfully for fileId '%s'", docID.String())
	return c.Redirect(http.StatusSeeOther, "/upload")
}


