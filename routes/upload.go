package routes

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
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
	ctx := c.Request().Context()

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

	// 🧠 Check if document already exists in DB by S3 path
	existingDoc, err := h.db.FindDocumentByS3Path(ctx, s3Path)
	if err == nil {
		return web.Render(c, http.StatusOK, components.DuplicateUploadResponse(existingDoc.ID.String()))
	}

	file, err := fileHeader.Open()
	if err != nil {
		return web.Render(c, http.StatusOK, components.ErrorMessage(fmt.Sprintf("Error opening file: %v", err)))
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			h.logger.Error("Error closing file: %v", err)
		}
	}(file)

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return web.Render(c, http.StatusOK, components.ErrorMessage(fmt.Sprintf("Error reading file: %v", err)))
	}

	metadata, err := h.bedrock.ProcessPdfAndExtractMetadata(context.Background(), fileBytes)
	if err != nil {
		return web.Render(c, http.StatusOK, components.ErrorMessage(err.Error()))
	}

	format := "2006-01-02"

	parse, err := time.Parse(format, metadata.PublishDate)
	if err != nil {
		parse = time.Now()
		fmt.Println(err)
	}
	sqlTime := sql.NullTime{
		Time:  parse,
		Valid: true,
	}

	prettyJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		log.Printf("⚠️ Error formatting metadata as JSON: %v", err)
		fmt.Printf("Raw Metadata: %+v\n", metadata) // Print raw struct if formatting fails
	} else {
		fmt.Println("--- Extracted Metadata ---")
		fmt.Println(string(prettyJSON))
		fmt.Println("--- End Extracted Metadata ---")
	}

	if err := h.db.InsertUploadedDocument(ctx, db.InsertUploadedDocumentParams{
		ID:          fileID,
		S3File:      s3Path,
		FileName:    originalFilename,
		Abstract:    sql.NullString{String: metadata.Abstract, Valid: true},
		PublishDate: sqlTime,
		Title:       metadata.Title,
	}); err != nil {
		log.Printf("DB insert failed: %v", err)
		errorMessage := "Could not save file metadata to database"
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
		return web.Render(c, http.StatusInternalServerError, components.UploadResponse(false, errorMessage))
	}

	h.modifyAuthors(ctx, metadata.AuthorName)
	h.modifyCategories(ctx, metadata.CategoryName)
	h.modifyKeywords(ctx, metadata.KeywordName)
	h.modifyRegions(ctx, metadata.RegionName)

	err = h.updateManyToManyFieldsMetadata(
		ctx,
		uuid.NullUUID{UUID: fileID, Valid: true},
		metadata.AuthorName,
		metadata.KeywordName,
		metadata.CategoryName,
		metadata.RegionName,
	)

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

	err = h.updateManyToManyFieldsMetadata(ctx, documentID, authorStrs, keywordStrs, categoryStrs, regionStrs)
	if err != nil {
		return web.Render(c, http.StatusOK, components.ErrorMessage(fmt.Sprintf("[ERROR] Error updating document metadata: %v", err)))
	}

	log.Printf("[INFO] Metadata updated successfully for fileId '%s'", docID.String())
	return web.Render(c, http.StatusOK, components.SuccessMessage(fmt.Sprintf("Metadata updated successfully for fileId '%s'", docID)))
}

func (h *Handler) updateManyToManyFieldsMetadata(
	ctx context.Context,
	documentID uuid.NullUUID,
	authorStrs []string,
	keywordStrs []string,
	categoryStrs []string,
	regionStrs []string,
) error {
	authors := util.ResolveIDs(ctx, h.db, authorStrs, util.GetOrCreateAuthor)
	keywords := util.ResolveIDs(ctx, h.db, keywordStrs, util.GetOrCreateKeyword)
	categories := util.ResolveIDs(ctx, h.db, categoryStrs, util.GetOrCreateCategory)
	regions := util.ResolveIDs(ctx, h.db, regionStrs, util.GetOrCreateRegion)

	fmt.Println(documentID)

	for _, authorID := range authors {
		err := h.db.InsertDocAuthor(ctx, db.InsertDocAuthorParams{
			ID:       uuid.New(),
			DocID:    documentID,
			AuthorID: uuid.NullUUID{UUID: authorID, Valid: true},
		})
		if err != nil {
			fmt.Println("[ERROR] Failed to insert into doc_authors: %v", err)
			return err
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
			return err
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
			return err
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
			return err
		}
	}
	return nil
}

func (h *Handler) modifyAuthors(ctx context.Context, authorStrs []string) error {
	for i, authorStr := range authorStrs {
		authorStrs[i] = "new:" + authorStr
	}
	return nil
}

func (h *Handler) modifyRegions(ctx context.Context, regionStrs []string) error {
	for i, regionStr := range regionStrs {
		regionStrs[i] = "new:" + regionStr
	}
	return nil
}

func (h *Handler) modifyCategories(ctx context.Context, categoriesStrs []string) error {
	for i, categoriesStr := range categoriesStrs {
		categoriesStrs[i] = "new:" + categoriesStr
	}
	return nil
}

func (h *Handler) modifyKeywords(ctx context.Context, keywordsStrs []string) error {
	for i, keywordStr := range keywordsStrs {
		keywordsStrs[i] = "new:" + keywordStr
	}
	return nil
}
