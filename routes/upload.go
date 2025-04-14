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

	db "github.com/DSSD-Madison/gmu/pkg/db/generated"
	"github.com/DSSD-Madison/gmu/pkg/db/util"
	"github.com/DSSD-Madison/gmu/pkg/middleware"
	"github.com/DSSD-Madison/gmu/web"
	"github.com/DSSD-Madison/gmu/web/components"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *Handler) PDFUploadPage(c echo.Context) error {
	csrf := c.Get("csrf").(string)
	isAuthorized, isMaster := middleware.GetSessionFlags(c)
	return web.Render(c, http.StatusOK, components.PDFUpload(csrf, isAuthorized, isMaster))
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

	selectedAuthors := util.ToAuthorPairs(allAuthors, authorNames)
	selectedKeywords := util.ToKeywordPairs(allKeywords, keywordNames)
	selectedRegions := util.ToRegionPairs(allRegions, regionNames)
	selectedCategories := util.ToCategoryPairs(allCategories, categoryNames)

	csrf, ok := c.Get("csrf").(string)
	if !ok {
		log.Println("CSRF token not found in context")
	}
	isAuthorized, isMaster := middleware.GetSessionFlags(c)
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
		isAuthorized,
		isMaster,
	))

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
	
	authors := util.ResolveIDs(ctx, h.db, authorStrs, util.GetOrCreateAuthor)
	keywords := util.ResolveIDs(ctx, h.db, keywordStrs, util.GetOrCreateKeyword)
	categories := util.ResolveIDs(ctx, h.db, categoryStrs, util.GetOrCreateCategory)
	regions := util.ResolveIDs(ctx, h.db, regionStrs, util.GetOrCreateRegion)
	
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
