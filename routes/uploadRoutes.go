package routes

import (
	"github.com/DSSD-Madison/gmu/pkg/handlers"
	"github.com/DSSD-Madison/gmu/pkg/services"
	"github.com/labstack/echo/v4"
)

func RegisterUploadRoutes(e *echo.Echo, uploadHandler *handlers.UploadHandler, sessionManager services.SessionManager) {
	// Page to display the upload form
	e.GET("/upload", uploadHandler.PDFUploadPage, sessionManager.RequireAuth)

	// Action endpoint to handle the actual file upload POST request
	e.POST("/upload", uploadHandler.HandlePDFUpload, sessionManager.RequireAuth) // <<< CORRECTED HANDLER

	// Page to display the metadata edit form, identified by fileId
	e.GET("/edit-metadata/:fileId", uploadHandler.PDFMetadataEditPage, sessionManager.RequireAuth) // <<< ADDED ROUTE

	// Action endpoint to handle the saving of edited metadata
	e.POST("/save-metadata", uploadHandler.HandleMetadataSave, sessionManager.RequireAuth)

	e.POST("/toggle-delete", uploadHandler.ToggleDelete, sessionManager.RequireAuth)

	e.GET("/latest", uploadHandler.LatestDocumentsPage, sessionManager.RequireAuth)

	e.POST("/documents-search", uploadHandler.SearchDocumentsPage, sessionManager.RequireAuth)
}
