package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/pkg/middleware"
)

// InitRoutes registers all the application routes
func InitRoutes(
	e *echo.Echo,
	homeHandler *HomeHandler,
	searchHandler *SearchHandler,
	suggestionsHandler *SuggestionsHandler,
	uploadHandler *UploadHandler,
	authenticationHandler *AuthenticationHandler,
	userManagementHandler *UserManagementHandler,
	databaseHandler *DatabaseHandler,
) {
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
	e.GET("/login", authenticationHandler.LoginPage)
	e.POST("/login", authenticationHandler.Login)

	// Logout Route
	e.GET("/logout", authenticationHandler.Logout) // for dev testing, remove when nav bar added
	e.POST("/logout", authenticationHandler.Logout)

	// Admin Routes
	e.GET("/admin/users", userManagementHandler.ManageUsersPage, middleware.RequireAuth)
	e.POST("/admin/users", userManagementHandler.CreateNewUser, middleware.RequireAuth)
	e.POST("/admin/users/delete", userManagementHandler.DeleteUser, middleware.RequireAuth)

	// --- Database Search Routes ---
	e.GET("/authors", databaseHandler.DatabaseSearchAuthors, middleware.RequireAuth)
	e.GET("/keywords", databaseHandler.DatabaseSearchKeywords, middleware.RequireAuth)
	e.GET("/regions", databaseHandler.DatabaseSearchRegions, middleware.RequireAuth)
	e.GET("/categories", databaseHandler.DatabaseSearchCategories, middleware.RequireAuth)
}
