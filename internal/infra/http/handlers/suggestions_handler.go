package handlers

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/internal/application"
	"github.com/DSSD-Madison/gmu/pkg/logger"
	"github.com/DSSD-Madison/gmu/web"
	"github.com/DSSD-Madison/gmu/web/components"
)

type SuggestionsHandler struct {
	log       logger.Logger
	suggester application.Suggester
}

func NewSuggestionsHandler(log logger.Logger, suggester application.Suggester) *SuggestionsHandler {
	handlerLogger := log.With("Handler", "Suggestions")
	return &SuggestionsHandler{
		log:       handlerLogger,
		suggester: suggester,
	}
}

func (h *SuggestionsHandler) SearchSuggestions(c echo.Context) error {
	ctx := c.Request().Context()
	query := strings.TrimSpace(c.FormValue("query"))

	h.log.InfoContext(ctx, "Fetching suggestions", "query", query)

	suggestions, err := h.suggester.GetSuggestions(ctx, query)
	if err != nil {
		h.log.ErrorContext(ctx, "Suggestion service failed", "query", query, "error", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	h.log.DebugContext(ctx, "Rendering suggestions", "count", len(suggestions.Suggestions))
	return web.Render(c, http.StatusOK, components.Suggestions(suggestions))
}
