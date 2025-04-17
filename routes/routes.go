package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/pkg/middleware"
)

// InitRoutes registers all the application routes
func InitRoutes(e *echo.Echo, homeHandler *HomeHandler, searchHandler *SearchHandler, suggestionsHandler *SuggestionsHandler, loginHandler *LoginHandler, uploadHandler *UploadHandler, h *Handler) {
	e.GET("/", homeHandler.Home)

	// --- Search Routes ---
	e.GET("/search", searchHandler.Search)
	e.POST("/search/suggestions", suggestionsHandler.SearchSuggestions)

	// --- PDF Upload and Metadata Routes ---
	// Page to display the upload form
	e.GET("/upload", uploadHandler.PDFUploadPage, middleware.RequireAuth)

	// Action endpoint to handle the actual file upload POST request
	e.POST("/upload", uploadHandler.HandlePDFUpload, middleware.RequireAuth) // <<< CORRECTED HANDLER

	// Page to display the metadata edit form, identified by fileId
	e.GET("/edit-metadata/:fileId", uploadHandler.PDFMetadataEditPage, middleware.RequireAuth) // <<< ADDED ROUTE

	// Action endpoint to handle the saving of edited metadata
	e.POST("/save-metadata", uploadHandler.HandleMetadataSave, middleware.RequireAuth) // <<< ADDED ROUTE

	// Login Route
	e.GET("/login", loginHandler.LoginPage)
	e.POST("/login", loginHandler.Login)

	// Logout Route
	e.GET("/logout", loginHandler.Logout) // for dev testing, remove when nav bar added
	e.POST("/logout", loginHandler.Logout)

	// Admin Routes
	e.GET("/admin/users", h.ManageUsersPage, middleware.RequireAuth)
	e.POST("/admin/users", h.CreateNewUser, middleware.RequireAuth)
	e.POST("/admin/users/delete", h.DeleteUser, middleware.RequireAuth)

	// --- Database Search Routes ---
	e.GET("/authors", h.DatabaseSearchAuthors, middleware.RequireAuth)
	e.GET("/keywords", h.DatabaseSearchKeywords, middleware.RequireAuth)
	e.GET("/regions", h.DatabaseSearchRegions, middleware.RequireAuth)
	e.GET("/categories", h.DatabaseSearchCategories, middleware.RequireAuth)
}
