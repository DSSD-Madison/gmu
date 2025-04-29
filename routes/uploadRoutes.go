package routes

import (
	"github.com/DSSD-Madison/gmu/pkg/handlers"
	"github.com/DSSD-Madison/gmu/pkg/middleware"
	"github.com/labstack/echo/v4"
)

func RegisterUploadRoutes(e *echo.Echo, uploadHandler *handlers.UploadHandler) {
	// Page to display the upload form
	e.GET("/upload", uploadHandler.PDFUploadPage, middleware.RequireAuth)

	// Action endpoint to handle the actual file upload POST request
	e.POST("/upload", uploadHandler.HandlePDFUpload, middleware.RequireAuth) // <<< CORRECTED HANDLER

	// Page to display the metadata edit form, identified by fileId
	e.GET("/edit-metadata/:fileId", uploadHandler.PDFMetadataEditPage, middleware.RequireAuth) // <<< ADDED ROUTE

	// Action endpoint to handle the saving of edited metadata
	e.POST("/save-metadata", uploadHandler.HandleMetadataSave, middleware.RequireAuth)

	e.POST("/toggle-delete", uploadHandler.ToggleDelete, middleware.RequireAuth)

	e.GET("/latest", uploadHandler.LatestDocumentsPage, middleware.RequireAuth)

	e.POST("/documents-search", uploadHandler.SearchDocumentsPage, middleware.RequireAuth)

}
