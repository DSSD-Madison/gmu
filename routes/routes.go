package routes

import (
	"github.com/labstack/echo/v4"
)

// InitRoutes registers all the application routes
func InitRoutes(e *echo.Echo, h Handler) {
	// --- General Routes ---
	e.GET("/", h.Home)

	// --- Search Routes ---
	e.GET("/search", h.Search)
	e.POST("/search/suggestions", h.SearchSuggestions)

	// --- PDF Upload and Metadata Routes ---
	// Page to display the upload form
	e.GET("/upload", h.PDFUploadPage)

	// Action endpoint to handle the actual file upload POST request
	e.POST("/upload", h.HandlePDFUpload) // <<< CORRECTED HANDLER

	// Page to display the metadata edit form, identified by fileId
	e.GET("/edit-metadata/:fileId", h.PDFMetadataEditPage) // <<< ADDED ROUTE

	// Action endpoint to handle the saving of edited metadata
	e.POST("/save-metadata", h.HandleMetadataSave) // <<< ADDED ROUTE
}
