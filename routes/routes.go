package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/middleware"
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
	e.GET("/upload", h.PDFUploadPage, middleware.RequireAuth)

	// Action endpoint to handle the actual file upload POST request
	e.POST("/upload", h.HandlePDFUpload, middleware.RequireAuth) // <<< CORRECTED HANDLER

	// Page to display the metadata edit form, identified by fileId
	e.GET("/edit-metadata/:fileId", h.PDFMetadataEditPage, middleware.RequireAuth) // <<< ADDED ROUTE

	// Action endpoint to handle the saving of edited metadata
	e.POST("/save-metadata", h.HandleMetadataSave, middleware.RequireAuth) // <<< ADDED ROUTE

	// Login Route
	e.GET("/login", h.LoginPage)
	e.POST("/login", h.Login)

	// Logout Route
	e.GET("/logout", h.Logout) // for dev testing, remove when nav bar added
	e.POST("/logout", h.Logout) 

	// Admin Routes
	e.GET("/admin/users", h.ManageUsersPage, middleware.RequireAuth)
	e.POST("/admin/users", h.CreateNewUser, middleware.RequireAuth)
}
