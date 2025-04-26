package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/pkg/awskendra"
	db "github.com/DSSD-Madison/gmu/pkg/db/generated"
	"github.com/DSSD-Madison/gmu/pkg/db/util"
	"github.com/DSSD-Madison/gmu/pkg/logger"
	"github.com/DSSD-Madison/gmu/pkg/services"
	"github.com/DSSD-Madison/gmu/web"
	"github.com/DSSD-Madison/gmu/web/components"
)

// UploadHandler TODO: Separate Metadata logic and export into services
type UploadHandler struct {
	log            logger.Logger
	bedrockManager services.BedrockManager
	fileManager    *services.FilemanagerService
	sessionManager services.SessionManager
	db             *db.Queries
}

func NewUploadHandler(log logger.Logger, db *db.Queries, bedrockManager services.BedrockManager, fms *services.FilemanagerService, sessionManager services.SessionManager) *UploadHandler {
	handlerLogger := log.With("Handler", "Upload")
	return &UploadHandler{
		log:            handlerLogger,
		bedrockManager: bedrockManager,
		sessionManager: sessionManager,
		db:             db,
		fileManager:    fms,
	}
}

func (uh *UploadHandler) PDFUploadPage(c echo.Context) error {
	csrf := c.Get("csrf").(string)
	isAuthorized := uh.sessionManager.IsAuthenticated(c)
	isMaster := uh.sessionManager.IsMaster(c)
	return web.Render(c, http.StatusOK, components.PDFUpload(csrf, isAuthorized, isMaster))
}

const dateFormat = "2006-01-02"

func (uh *UploadHandler) HandlePDFUpload(c echo.Context) error {
	ctx := c.Request().Context()

	fileHeader, err := c.FormFile("pdf")
	if err != nil {
		return uh.renderError(c, http.StatusOK, "Failed to get file: %v", err)
	}
	clientMime := fileHeader.Header.Get("Content-Type")
	filename := fileHeader.Filename
	fileID := uuid.New()
	s3Key := filename
	s3Path := fmt.Sprintf("s3://manually-uploaded-bep/%s", filename)

	// Check for duplicate
	if existing, err := uh.db.FindDocumentByS3Path(ctx, s3Path); err == nil {
		return web.Render(c, http.StatusOK, components.DuplicateUploadResponse(existing.ID.String()))
	}

	// Read file bytes
	fileBytes, err := uh.readMultipartFile(fileHeader)
	if err != nil {
		return uh.renderError(c, http.StatusOK, "Error reading file: %v", err)
	}

	// Upload to S3
	if err := uh.fileManager.UploadFile(ctx, s3Key, fileBytes, clientMime); err != nil {
		return uh.renderError(c, http.StatusOK, "Error uploading file: %v", err)
	}

	// Extract metadata
	metadata, err := uh.bedrockManager.ExtractPDFMetadata(ctx, fileBytes)
	if err != nil || metadata == nil {
		uh.cleanupOnError(ctx, filename)
		return uh.renderError(c, http.StatusOK, "Error extracting metadata: %v", err)
	}

	// Parse publish date
	publishDate := uh.parsePublishDate(ctx, metadata.PublishDate)

	// Insert document record
	if err := uh.db.InsertUploadedDocument(ctx, db.InsertUploadedDocumentParams{
		ID:          fileID,
		S3File:      s3Path,
		FileName:    filename,
		Title:       metadata.Title,
		Abstract:    sql.NullString{String: metadata.Abstract, Valid: true},
		PublishDate: publishDate,
	}); err != nil {
		uh.cleanupOnError(ctx, s3Key)
		return uh.renderError(c, 200, "Could not save file metadata to database")
	}

	// Update many-to-many joins
	if err := uh.addAndSaveAssociations(ctx, fileID, metadata); err != nil {
		uh.cleanupOnError(ctx, s3Key)
		err := uh.db.DeleteDocumentByID(ctx, fileID)
		if err != nil {
			uh.log.ErrorContext(ctx, "Error deleting document from db: %v", err)
		}
		return uh.renderError(c, http.StatusOK, "Could not save associated metadata")
	}

	// Redirect to metadata editor
	c.Response().Header().Set("HX-Redirect", fmt.Sprintf("/edit-metadata/%s", fileID))
	return c.NoContent(http.StatusOK)
}

func (uh *UploadHandler) readMultipartFile(fh *multipart.FileHeader) ([]byte, error) {
	file, err := fh.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

func (uh *UploadHandler) parsePublishDate(ctx context.Context, raw string) sql.NullTime {
	if raw == "" {
		return sql.NullTime{}
	}
	t, err := time.Parse(dateFormat, raw)
	if err != nil {
		uh.log.ErrorContext(ctx, "Invalid publish date format", "error", err)
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t, Valid: true}
}

func (uh *UploadHandler) addAndSaveAssociations(ctx context.Context, docID uuid.UUID, m *awskendra.ExtractedMetadata) error {
	uh.addNewToStr(ctx, m.AuthorName)
	uh.addNewToStr(ctx, m.CategoryName)
	uh.addNewToStr(ctx, m.KeywordName)
	uh.addNewToStr(ctx, m.RegionName)
	return uh.updateManyToManyFieldsMetadata(
		ctx,
		uuid.NullUUID{UUID: docID, Valid: true},
		m.AuthorName,
		m.KeywordName,
		m.CategoryName,
		m.RegionName,
	)
}

func (uh *UploadHandler) cleanupOnError(ctx context.Context, key string) {
	if err := uh.fileManager.DeleteFile(ctx, key, "manually-uploaded-bep"); err != nil {
		uh.log.ErrorContext(ctx, "Failed to delete S3 file during cleanup", "error", err)
	}
}

func (uh *UploadHandler) renderError(c echo.Context, code int, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return web.Render(c, code, components.ErrorMessage(msg))
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
	isAuthorized := uh.sessionManager.IsAuthenticated(c)
	isMaster := uh.sessionManager.IsMaster(c)

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
		doc.ToDelete,
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
		ToIndex:     sql.NullBool{Bool: true, Valid: true},
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

func (uh *UploadHandler) addNewToStr(ctx context.Context, strs []string) {
	for i, s := range strs {
		strs[i] = "new:" + s
	}
}

func (uh *UploadHandler) updateManyToManyFieldsMetadata(
	ctx context.Context,
	documentID uuid.NullUUID,
	authorStrs []string,
	keywordStrs []string,
	categoryStrs []string,
	regionStrs []string,
) error {
	authors := util.ResolveIDs(ctx, uh.db, authorStrs, util.GetOrCreateAuthor)
	keywords := util.ResolveIDs(ctx, uh.db, keywordStrs, util.GetOrCreateKeyword)
	categories := util.ResolveIDs(ctx, uh.db, categoryStrs, util.GetOrCreateCategory)
	regions := util.ResolveIDs(ctx, uh.db, regionStrs, util.GetOrCreateRegion)

	fmt.Println(documentID)

	for _, authorID := range authors {
		err := uh.db.InsertDocAuthor(ctx, db.InsertDocAuthorParams{
			ID:       uuid.New(),
			DocID:    documentID,
			AuthorID: uuid.NullUUID{UUID: authorID, Valid: true},
		})
		if err != nil {
			fmt.Printf("[ERROR] Failed to insert into doc_authors: %v\n", err)
			return err
		}
	}

	for _, keywordID := range keywords {
		err := uh.db.InsertDocKeyword(ctx, db.InsertDocKeywordParams{
			ID:        uuid.New(),
			DocID:     documentID,
			KeywordID: uuid.NullUUID{UUID: keywordID, Valid: true},
		})
		if err != nil {
			log.Printf("[ERROR] Failed to insert into doc_keywords: %v", err)
			return err
		}
	}

	for _, categoryID := range categories {
		err := uh.db.InsertDocCategory(ctx, db.InsertDocCategoryParams{
			ID:         uuid.New(),
			DocID:      documentID,
			CategoryID: uuid.NullUUID{UUID: categoryID, Valid: true},
		})
		if err != nil {
			log.Printf("[ERROR] Failed to insert into doc_categories: %v", err)
			return err
		}
	}

	for _, regionID := range regions {
		err := uh.db.InsertDocRegion(ctx, db.InsertDocRegionParams{
			ID:       uuid.New(),
			DocID:    documentID,
			RegionID: uuid.NullUUID{UUID: regionID, Valid: true},
		})
		if err != nil {
			log.Printf("[ERROR] Failed to insert into doc_regions: %v", err)
			return err
		}
	}
	return nil
}

func (uh *UploadHandler) ToggleDelete(c echo.Context) error {
	ctx := c.Request().Context()

	documentID := c.FormValue("fileId")
	markStr := c.FormValue("mark")

	toDelete, err := strconv.ParseBool(markStr)
	if err != nil {
		return web.Render(c, 200, components.ErrorMessage(err.Error()))
	}

	id, err := uuid.Parse(documentID)
	if err != nil {
		return web.Render(c, 200, components.ErrorMessage(err.Error()))
	}

	if err := uh.db.UpdateDocumentDeletionStatus(ctx, db.UpdateDocumentDeletionStatusParams{
		ID:       id,
		ToDelete: toDelete,
	}); err != nil {
		return web.Render(c, 200, components.ErrorMessage(err.Error()))
	}

	buttonText := "Delete"
	if toDelete {
		buttonText = "Undo Delete"
	}
	successMessage := "Unmarked for Deletion"
	if toDelete {
		successMessage = "Marked for Deletion"
	}

	return web.Render(
		c,
		200,
		components.ToggleDeleteButton(id.String(), !toDelete, buttonText),
		components.SuccessMessage(successMessage),
	)

}
