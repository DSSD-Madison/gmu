package routes

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	db "github.com/DSSD-Madison/gmu/pkg/db/generated"
	"github.com/DSSD-Madison/gmu/pkg/db/util"
	"github.com/DSSD-Madison/gmu/pkg/logger"
	"github.com/DSSD-Madison/gmu/pkg/middleware"
	"github.com/DSSD-Madison/gmu/web"
	"github.com/DSSD-Madison/gmu/web/components"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type UploadHandler struct {
	log logger.Logger
	db  *db.Queries
}

func NewUploadHandler(log logger.Logger, db *db.Queries) *UploadHandler {
	handlerLogger := log.With("Handler", "Upload")
	return &UploadHandler{
		log: handlerLogger,
		db:  db,
	}
}

func (uh *UploadHandler) PDFUploadPage(c echo.Context) error {
	csrf := c.Get("csrf").(string)
	isAuthorized, isMaster := middleware.GetSessionFlags(c)
	return web.Render(c, http.StatusOK, components.PDFUpload(csrf, isAuthorized, isMaster))
}

func (uh *UploadHandler) HandlePDFUpload(c echo.Context) error {
	ctx := c.Request().Context()

	fileHeader, err := c.FormFile("pdf")
	if err != nil {
		uh.log.ErrorContext(c.Request().Context(), "Error getting uploaded file", "error", err)
		errorMessage := fmt.Sprintf("Failed to get file: %v", err)
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
		return web.Render(c, http.StatusBadRequest, components.UploadResponse(false, errorMessage))
	}

	originalFilename := fileHeader.Filename
	fileID := uuid.New()
	s3Path := fmt.Sprintf("s3://your-bucket-name/documents/%s", originalFilename)
	title := strings.TrimSuffix(originalFilename, path.Ext(originalFilename))

	// 🧠 Check if document already exists in DB by S3 path
	existingDoc, err := uh.db.FindDocumentByS3Path(ctx, s3Path)
	if err == nil {
		return web.Render(c, http.StatusOK, components.DuplicateUploadResponse(existingDoc.ID.String()))
	}

	if err := uh.db.InsertUploadedDocument(ctx, db.InsertUploadedDocumentParams{
		ID:       fileID,
		S3File:   s3Path,
		FileName: originalFilename,
		Title:    title,
	}); err != nil {
		uh.log.ErrorContext(c.Request().Context(), "DB insert failed", "error", err)
		errorMessage := "Could not save file metadata to database"
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
		return web.Render(c, http.StatusInternalServerError, components.UploadResponse(false, errorMessage))
	}

	redirectPath := fmt.Sprintf("/edit-metadata/%s", fileID.String())
	c.Response().Header().Set("HX-Redirect", redirectPath)
	return c.NoContent(http.StatusOK)
}

func (uh *UploadHandler) PDFMetadataEditPage(c echo.Context) error {
	fileId := c.Param("fileId")
	if fileId == "" {
		uh.log.ErrorContext(c.Request().Context(), "Missing fileId")
		return c.Redirect(http.StatusTemporaryRedirect, "/upload")
	}
	docUUID, err := uuid.Parse(fileId)
	if err != nil {
		return err
	}

	doc, err := uh.db.FindDocumentByID(context.Background(), docUUID)
	if err != nil {
		return err
	}

	allAuthors, _ := uh.db.ListAllAuthors(c.Request().Context())
	allKeywords, _ := uh.db.ListAllKeywords(c.Request().Context())
	allRegions, _ := uh.db.ListAllRegions(c.Request().Context())
	allCategories, _ := uh.db.ListAllCategories(c.Request().Context())

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
		uh.log.WarnContext(c.Request().Context(), "CSRF token not found in context")
	}
	isAuthorized, isMaster := middleware.GetSessionFlags(c)

	s3Link := util.ConvertS3URIToURL(doc.S3File)
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
		s3Link,
	))

}

func (uh *UploadHandler) HandleMetadataSave(c echo.Context) error {
	ctx := c.Request().Context()

	fileId := c.FormValue("fileId")
	title := c.FormValue("title")
	abstract := c.FormValue("abstract")
	publishDate := c.FormValue("publish_date")
	source := c.FormValue("source")

	form, err := c.FormParams()
	if err != nil {
		uh.log.ErrorContext(c.Request().Context(), "Failed to parse form params", "error", err)
		return web.Render(c, http.StatusOK, components.ErrorMessage("Failed to parse form. Please check log"))
	}
	authorStrs := form["author_names"]
	keywordStrs := form["keyword_names"]
	categoryStrs := form["category_names"]
	regionStrs := form["region_names"]

	docID, err := uuid.Parse(fileId)
	if err != nil {
		uh.log.ErrorContext(c.Request().Context(), "Invalid UUID in form", "error", err)
		return err
	}

	var parsedDate sql.NullTime
	if publishDate != "" {
		t, err := time.Parse("2006-01-02", publishDate)
		if err == nil {
			parsedDate = sql.NullTime{Time: t, Valid: true}
		}
	}

	err = uh.db.UpdateDocumentMetadata(ctx, db.UpdateDocumentMetadataParams{
		ID:          docID,
		Title:       title,
		Abstract:    sql.NullString{String: abstract, Valid: abstract != ""},
		PublishDate: parsedDate,
		Source:      sql.NullString{String: source, Valid: source != ""},
	})
	if err != nil {
		uh.log.ErrorContext(c.Request().Context(), "Error updating document metadata", "error", err)
		return web.Render(c, http.StatusOK, components.ErrorMessage(fmt.Sprintf("[ERROR] Error updating document metadata: %v", err)))
	}
	documentID := uuid.NullUUID{UUID: docID, Valid: true}

	uh.db.DeleteDocAuthorsByDocID(ctx, documentID)
	uh.db.DeleteDocKeywordsByDocID(ctx, documentID)
	uh.db.DeleteDocCategoriesByDocID(ctx, documentID)
	uh.db.DeleteDocRegionsByDocID(ctx, documentID)

	authors := util.ResolveIDs(ctx, uh.db, authorStrs, util.GetOrCreateAuthor)
	keywords := util.ResolveIDs(ctx, uh.db, keywordStrs, util.GetOrCreateKeyword)
	categories := util.ResolveIDs(ctx, uh.db, categoryStrs, util.GetOrCreateCategory)
	regions := util.ResolveIDs(ctx, uh.db, regionStrs, util.GetOrCreateRegion)

	uh.log.DebugContext(c.Request().Context(), "Author IDs", authors)
	uh.log.DebugContext(c.Request().Context(), "Keyword IDs", keywords)
	uh.log.DebugContext(c.Request().Context(), "Category IDs", categories)
	uh.log.DebugContext(c.Request().Context(), "Region IDs", regions)

	for _, authorID := range authors {
		err := uh.db.InsertDocAuthor(ctx, db.InsertDocAuthorParams{
			ID:       uuid.New(),
			DocID:    documentID,
			AuthorID: uuid.NullUUID{UUID: authorID, Valid: true},
		})
		if err != nil {
			uh.log.WarnContext(c.Request().Context(), "Failed to insert into doc_authors", "error", err)
		}
	}

	for _, keywordID := range keywords {
		err := uh.db.InsertDocKeyword(ctx, db.InsertDocKeywordParams{
			ID:        uuid.New(),
			DocID:     documentID,
			KeywordID: uuid.NullUUID{UUID: keywordID, Valid: true},
		})
		if err != nil {
			uh.log.WarnContext(c.Request().Context(), "Failed to insert into doc_keywords", "error", err)
		}
	}

	for _, categoryID := range categories {
		err := uh.db.InsertDocCategory(ctx, db.InsertDocCategoryParams{
			ID:         uuid.New(),
			DocID:      documentID,
			CategoryID: uuid.NullUUID{UUID: categoryID, Valid: true},
		})
		if err != nil {
			uh.log.WarnContext(c.Request().Context(), "Failed to insert into doc_categories", "error", err)
		}
	}

	for _, regionID := range regions {
		err := uh.db.InsertDocRegion(ctx, db.InsertDocRegionParams{
			ID:       uuid.New(),
			DocID:    documentID,
			RegionID: uuid.NullUUID{UUID: regionID, Valid: true},
		})
		if err != nil {
			uh.log.WarnContext(c.Request().Context(), "Failed to insert into doc_regions", "error", err)
		}
	}

	uh.log.InfoContext(c.Request().Context(), "Metadata updated successfully for fileId", "docID", docID.String())
	return web.Render(c, http.StatusOK, components.SuccessMessage(fmt.Sprintf("Metadata updated successfully for fileId '%s'", docID)))
}
