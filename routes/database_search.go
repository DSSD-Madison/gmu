package routes

import (
	"database/sql"
	"github.com/DSSD-Madison/gmu/web"
	"github.com/DSSD-Madison/gmu/web/components"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

func (h *Handler) DatabaseFieldSearch(c echo.Context, fieldName string) error {
	var idPrefix string

	switch fieldName {
	case "region_names":
		idPrefix = "regions"
	case "keyword_names":
		idPrefix = "keywords"
	case "author_names":
		idPrefix = "authors"
	default:
		h.logger.Error("DatabaseFieldSearch called with unsupported fieldName: %s", fieldName)
		return c.String(http.StatusInternalServerError, "Internal server configuration error.")
	}

	searchQuery := c.QueryParam("name")
	searchQuery = strings.TrimSpace(searchQuery) // Clean up input

	if searchQuery == "" {
		return web.Render(c, http.StatusOK, components.SuggestionList(idPrefix, fieldName, []string{}))
		// Alternatively: return c.NoContent(http.StatusOK) - depends on desired client behavior
	}

	ctx := c.Request().Context()
	var suggestions []string
	var dbErr error // Renamed to avoid shadowing in loops

	dbQuery := sql.NullString{String: searchQuery, Valid: true}

	switch fieldName {
	case "region_names":
		items, err := h.db.SearchRegionsByNamePrefix(ctx, dbQuery)
		if err != nil {
			h.logger.Error("Error searching regions for '%s': %v", searchQuery, err)
			dbErr = err // Store error, but continue (might return partial results if needed)
		} else {
			suggestions = make([]string, 0, len(items)) // Pre-allocate slice capacity
			for _, item := range items {
				suggestions = append(suggestions, item.Name) // Assuming Name field
			}
		}
	case "keyword_names":
		items, err := h.db.SearchKeywordsByNamePrefix(ctx, dbQuery)
		if err != nil {
			h.logger.Error("Error searching keywords for '%s': %v", searchQuery, err)
			dbErr = err
		} else {
			suggestions = make([]string, 0, len(items))
			for _, item := range items {
				suggestions = append(suggestions, item.Name) // Assuming Name field
			}
		}
	case "author_names":
		items, err := h.db.SearchAuthorsByNamePrefix(ctx, dbQuery)
		if err != nil {
			h.logger.Error("Error searching authors for '%s': %v", searchQuery, err)
			dbErr = err
		} else {
			suggestions = make([]string, 0, len(items))
			for _, item := range items {
				suggestions = append(suggestions, item.Name) // Assuming Name field
			}
		}
		// No default needed here as it's handled by the first switch
	}

	// Optional: Log if we are returning potentially incomplete results due to an error
	if dbErr != nil && len(suggestions) > 0 {
		h.logger.Warn("Returning %d suggestions for field '%s', query '%s' despite database error: %v", len(suggestions), fieldName, searchQuery, dbErr)
	} else if dbErr != nil { // Error occurred and no suggestions found
		h.logger.Error("Database error occurred and no suggestions found for field '%s', query '%s': %v", fieldName, searchQuery, dbErr)
		// Still render empty list - the error is logged, UI shows "no matches"
	}

	return web.Render(c, http.StatusOK, components.SuggestionList(idPrefix, fieldName, suggestions))
}

func (h *Handler) DatabaseSearchRegions(c echo.Context) error {
	return h.DatabaseFieldSearch(c, "region_names")
}

func (h *Handler) DatabaseSearchKeywords(c echo.Context) error {
	return h.DatabaseFieldSearch(c, "keyword_names")
}

func (h *Handler) DatabaseSearchAuthors(c echo.Context) error {
	return h.DatabaseFieldSearch(c, "author_names")
}
