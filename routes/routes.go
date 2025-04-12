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
	e.GET("/upload", h.PDFUploadPage)
	e.POST("/upload", h.HandlePDFUpload)
	e.GET("/edit-metadata/:fileId", h.PDFMetadataEditPage)
	e.POST("/save-metadata", h.HandleMetadataSave)

	// --- Database Search Routes ---
	e.GET("/authors", h.DatabaseSearchAuthors)
	e.GET("/keywords", h.DatabaseSearchKeywords)
	e.GET("/regions", h.DatabaseSearchRegions)

}
