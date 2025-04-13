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
	searchQuery = strings.TrimSpace(searchQuery)

	if searchQuery == "" {
		return web.Render(c, http.StatusOK, components.SuggestionList(idPrefix, fieldName, []components.Pair{}))
	}

	ctx := c.Request().Context()
	var suggestions []components.Pair
	var dbErr error

	dbQuery := sql.NullString{String: searchQuery, Valid: true}

	switch fieldName {
	case "region_names":
		items, err := h.db.SearchRegionsByNamePrefix(ctx, dbQuery)
		if err != nil {
			h.logger.Error("Error searching regions for '%s': %v", searchQuery, err)
			dbErr = err
		} else {
			for _, item := range items {
				suggestions = append(suggestions, components.Pair{
					ID:   item.ID.String(),
					Name: item.Name,
				})
			}
		}
	case "keyword_names":
		items, err := h.db.SearchKeywordsByNamePrefix(ctx, dbQuery)
		if err != nil {
			h.logger.Error("Error searching keywords for '%s': %v", searchQuery, err)
			dbErr = err
		} else {
			for _, item := range items {
				suggestions = append(suggestions, components.Pair{
					ID:   item.ID.String(),
					Name: item.Name,
				})
			}
		}
	case "author_names":
		items, err := h.db.SearchAuthorsByNamePrefix(ctx, dbQuery)
		if err != nil {
			h.logger.Error("Error searching authors for '%s': %v", searchQuery, err)
			dbErr = err
		} else {
			for _, item := range items {
				suggestions = append(suggestions, components.Pair{
					ID:   item.ID.String(),
					Name: item.Name,
				})
			}
		}
	}

	if dbErr != nil && len(suggestions) > 0 {
		h.logger.Warn("Returning %d suggestions for field '%s', query '%s' despite error: %v", len(suggestions), fieldName, searchQuery, dbErr)
	} else if dbErr != nil {
		h.logger.Error("No suggestions for field '%s', query '%s': %v", fieldName, searchQuery, dbErr)
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
