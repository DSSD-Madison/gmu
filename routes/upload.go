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
	db "github.com/DSSD-Madison/gmu/pkg/db/generated"
	"github.com/DSSD-Madison/gmu/web"
	"github.com/DSSD-Madison/gmu/web/components"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
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

func toAuthorPairs(all []db.Author, selected []string) []components.Pair {
	seen := make(map[string]bool)
	var out []components.Pair
	for _, name := range selected {
		norm := strings.ToLower(name)
		if seen[norm] {
			continue
		}
		for _, a := range all {
			if strings.EqualFold(a.Name, name) {
				out = append(out, components.Pair{
					ID:   a.ID.String(),
					Name: a.Name,
				})
				seen[norm] = true
				break
			}
		}
	}
	return out
}

func toKeywordPairs(all []db.Keyword, selected []string) []components.Pair {
	seen := make(map[string]bool)
	var out []components.Pair
	for _, name := range selected {
		norm := strings.ToLower(name)
		if seen[norm] {
			continue
		}
		for _, k := range all {
			if strings.EqualFold(k.Name, name) {
				out = append(out, components.Pair{
					ID:   k.ID.String(),
					Name: k.Name,
				})
				seen[norm] = true
				break
			}
		}
	}
	return out
}

func toRegionPairs(all []db.Region, selected []string) []components.Pair {
	seen := make(map[string]bool)
	var out []components.Pair
	for _, name := range selected {
		norm := strings.ToLower(name)
		if seen[norm] {
			continue
		}
		for _, r := range all {
			if strings.EqualFold(r.Name, name) {
				out = append(out, components.Pair{
					ID:   r.ID.String(),
					Name: r.Name,
				})
				seen[norm] = true
				break
			}
		}
	}
	return out
}

func toCategoryPairs(all []db.Category, selected []string) []components.Pair {
	seen := make(map[string]bool)
	var out []components.Pair
	for _, name := range selected {
		norm := strings.ToLower(name)
		if seen[norm] {
			continue
		}
		for _, c := range all {
			if strings.EqualFold(c.Name, name) {
				out = append(out, components.Pair{
					ID:   c.ID.String(),
					Name: c.Name,
				})
				seen[norm] = true
				break
			}
		}
	}
	return out
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

	doc, err := h.db.FindDocumentByID(context.Background(), docUUID)
	if err != nil {
		return err
	}

	allAuthors, _ := h.db.ListAllAuthors(c.Request().Context())
	allKeywords, _ := h.db.ListAllKeywords(c.Request().Context())
	allRegions, _ := h.db.ListAllRegions(c.Request().Context())
	allCategories, _ := h.db.ListAllCategories(c.Request().Context())

	authorNames := []string(doc.AuthorNames)
	keywordNames := []string(doc.KeywordNames)
	regionNames := []string(doc.RegionNames)
	categoryNames := []string(doc.CategoryNames)

	selectedAuthors := toAuthorPairs(allAuthors, authorNames)
	selectedKeywords := toKeywordPairs(allKeywords, keywordNames)
	selectedRegions := toRegionPairs(allRegions, regionNames)
	selectedCategories := toCategoryPairs(allCategories, categoryNames)

	csrf, ok := c.Get("csrf").(string)
	if !ok {
		log.Println("CSRF token not found in context")
	}

	return web.Render(c, http.StatusOK, components.PDFMetadataEditForm(
		fileId,
		doc.FileName,
		doc.Title,
		doc.Abstract.String,
		doc.PublishDate.Time.Format("2006-01-02"),
		doc.Source.String,
		selectedRegions,
		selectedKeywords,
		selectedAuthors,
		selectedCategories,
		csrf,
		allRegions,
		allKeywords,
		allAuthors,
		allCategories,
	))

}

func parseDocument(row db.FindDocumentByIDRow) awskendra.KendraResult {
	res := awskendra.KendraResult{
		Title:      row.Title,
		Authors:    row.AuthorNames,
		Categories: row.CategoryNames,
		Keywords:   row.KeywordNames,
		Regions:    row.RegionNames,
	}

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

	return res
}

func (h *Handler) HandleMetadataSave(c echo.Context) error {
	ctx := c.Request().Context()

	fileId := c.FormValue("fileId")
	title := c.FormValue("title")
	abstract := c.FormValue("abstract")
	publishDate := c.FormValue("publish_date")
	source := c.FormValue("source")

	form, err := c.FormParams()
	if err != nil {
		log.Printf("[ERROR] Failed to parse form params: %v", err)
		return web.Render(c, http.StatusOK, components.ErrorMessage("Failed to parse form. Please check log"))
	}
	authorStrs := form["author_names"]
	keywordStrs := form["keyword_names"]
	categoryStrs := form["category_names"]
	regionStrs := form["region_names"]

	docID, err := uuid.Parse(fileId)
	if err != nil {
		log.Printf("[ERROR] Invalid UUID in form: %v", err)
		return err
	}

	var parsedDate sql.NullTime
	if publishDate != "" {
		t, err := time.Parse("2006-01-02", publishDate)
		if err == nil {
			parsedDate = sql.NullTime{Time: t, Valid: true}
		}
	}

	err = h.db.UpdateDocumentMetadata(ctx, db.UpdateDocumentMetadataParams{
		ID:          docID,
		Title:       title,
		Abstract:    sql.NullString{String: abstract, Valid: abstract != ""},
		PublishDate: parsedDate,
		Source:      sql.NullString{String: source, Valid: source != ""},
	})
	if err != nil {
		log.Printf("[ERROR] Error updating document metadata: %v", err)
		return web.Render(c, http.StatusOK, components.ErrorMessage(fmt.Sprintf("[ERROR] Error updating document metadata: %v", err)))
	}
	documentID := uuid.NullUUID{UUID: docID, Valid: true}

	h.db.DeleteDocAuthorsByDocID(ctx, documentID)
	h.db.DeleteDocKeywordsByDocID(ctx, documentID)
	h.db.DeleteDocCategoriesByDocID(ctx, documentID)
	h.db.DeleteDocRegionsByDocID(ctx, documentID)

	parseUUIDListFromSlice := func(list []string) []uuid.UUID {
		result := []uuid.UUID{}
		for _, id := range list {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			u, err := uuid.Parse(id)
			if err == nil {
				result = append(result, u)
			} else {
				log.Printf("[WARN] Skipping invalid UUID: %s", id)
			}
		}
		return result
	}

	authors := parseUUIDListFromSlice(authorStrs)
	keywords := parseUUIDListFromSlice(keywordStrs)
	categories := parseUUIDListFromSlice(categoryStrs)
	regions := parseUUIDListFromSlice(regionStrs)

	log.Printf("[DEBUG] Parsed author IDs: %v", authors)
	log.Printf("[DEBUG] Parsed keyword IDs: %v", keywords)
	log.Printf("[DEBUG] Parsed category IDs: %v", categories)
	log.Printf("[DEBUG] Parsed region IDs: %v", regions)

	for _, authorID := range authors {
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
	return web.Render(c, http.StatusOK, components.SuccessMessage(fmt.Sprintf("Metadata updated successfully for fileId '%s'", docID)))
}
