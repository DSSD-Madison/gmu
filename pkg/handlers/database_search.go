package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	db "github.com/DSSD-Madison/gmu/pkg/db/generated"
	"github.com/DSSD-Madison/gmu/pkg/logger"
	"github.com/DSSD-Madison/gmu/web"
	"github.com/DSSD-Madison/gmu/web/components"
	"github.com/labstack/echo/v4"
)

type DatabaseHandler struct {
	log logger.Logger
	db  *db.Queries
}

func NewDatabaseHandler(log logger.Logger, db *db.Queries) *DatabaseHandler {
	handlerLogger := log.With("handler", "Database")
	return &DatabaseHandler{
		log: handlerLogger,
		db:  db,
	}
}

func (dh *DatabaseHandler) DatabaseFieldSearch(c echo.Context, fieldName string) error {
	var idPrefix string

	switch fieldName {
	case "region_names":
		idPrefix = "regions"
	case "keyword_names":
		idPrefix = "keywords"
	case "author_names":
		idPrefix = "authors"
	case "category_names":
		idPrefix = "categories"
	default:
		dh.log.Error("DatabaseFieldSearch called with unsupported fieldName", "fieldName", fieldName)
		return c.String(http.StatusInternalServerError, "Internal server configuration error.")
	}

	searchQuery := c.QueryParam("name")
	searchQuery = strings.TrimSpace(searchQuery)

	if searchQuery == "" {
		return web.Render(c, http.StatusOK, components.SuggestionList(idPrefix, fieldName, []components.Pair{}))
	}

	ctx := c.Request().Context()

	var suggestions []components.Pair

	if len(searchQuery) > 3 {
		suggestions = append(suggestions, components.Pair{
			ID:   "NON",
			Name: searchQuery,
		})
	}
	var dbErr error

	dbQuery := sql.NullString{String: searchQuery, Valid: true}

	switch fieldName {
	case "region_names":
		items, err := dh.db.SearchRegionsByNamePrefix(ctx, dbQuery)
		if err != nil {
			dh.log.Error("Error searching regions for '%s': %v", searchQuery, err)
			dh.log.Error("Error searching regions", "query", searchQuery, "error", err)
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
		items, err := dh.db.SearchKeywordsByNamePrefix(ctx, dbQuery)
		if err != nil {
			dh.log.Error("Error searching keywords", "query", searchQuery, "error", err)
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
		items, err := dh.db.SearchAuthorsByNamePrefix(ctx, dbQuery)
		if err != nil {
			dh.log.Error("Error searching authors", "query", searchQuery, "error", err)
			dbErr = err
		} else {
			for _, item := range items {
				suggestions = append(suggestions, components.Pair{
					ID:   item.ID.String(),
					Name: item.Name,
				})
			}
		}
	case "category_names":
		items, err := dh.db.SearchCategoriesByNamePrefix(ctx, dbQuery)
		if err != nil {
			dh.log.Error("Error searching categories", "query", searchQuery, "error", err)
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
		dh.log.Warn("Issue returning suggestions", "field", fieldName, "count", len(suggestions), "query", searchQuery, "error", dbErr)
	} else if dbErr != nil {
		dh.log.Error("No suggestions for field", "fieldName", fieldName, "query", searchQuery, "error", dbErr)
	}

	return web.Render(c, http.StatusOK, components.SuggestionList(idPrefix, fieldName, suggestions))
}

func (dh *DatabaseHandler) DatabaseSearchRegions(c echo.Context) error {
	return dh.DatabaseFieldSearch(c, "region_names")
}

func (dh *DatabaseHandler) DatabaseSearchKeywords(c echo.Context) error {
	return dh.DatabaseFieldSearch(c, "keyword_names")
}

func (dh *DatabaseHandler) DatabaseSearchAuthors(c echo.Context) error {
	return dh.DatabaseFieldSearch(c, "author_names")
}

func (dh *DatabaseHandler) DatabaseSearchCategories(c echo.Context) error {
	return dh.DatabaseFieldSearch(c, "category_names")
}
