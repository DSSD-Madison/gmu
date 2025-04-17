package routes

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/pkg/logger"
	"github.com/DSSD-Madison/gmu/pkg/services"
	"github.com/DSSD-Madison/gmu/web"
	"github.com/DSSD-Madison/gmu/web/components"
)

type SuggestionsHandler struct {
	log       logger.Logger
	suggester services.Suggester
}

func NewSuggestionsHandler(log logger.Logger, suggester services.Suggester) *SuggestionsHandler {
	handlerLogger := log.With("Handler", "Suggestions")
	return &SuggestionsHandler{
		log:       handlerLogger,
		suggester: suggester,
	}
}

func (h *SuggestionsHandler) SearchSuggestions(c echo.Context) error {
	ctx := c.Request().Context()
	query := strings.TrimSpace(c.FormValue("query"))

	// Don't return suggestions for empty or very short queries?
	if len(query) < 2 { // Example minimum length
		h.log.DebugContext(ctx, "Query too short for suggestions", "query", query)
		// Return empty response or specific component indicating no suggestions
		return c.NoContent(http.StatusOK) // Or return an empty suggestions component
	}

	h.log.InfoContext(ctx, "Fetching suggestions", "query", query)

	// --- Call the Suggestion Service ---
	suggestions, err := h.suggester.GetSuggestions(ctx, query)
	if err != nil {
		h.log.ErrorContext(ctx, "Suggestion service failed", "query", query, "error", err)
		// Don't expose internal errors usually. Maybe return empty?
		// Consider specific error types from service if needed.
		return c.NoContent(http.StatusInternalServerError) // Or OK with empty component
	}
	// --- End Service Call ---

	// Render the suggestions component
	h.log.DebugContext(ctx, "Rendering suggestions", "count", len(suggestions.Suggestions))
	return web.Render(c, http.StatusOK, components.Suggestions(suggestions))
}

func (h *Handler) SearchSuggestions(c echo.Context) error {
	query := c.FormValue("query")

	if query == "" {
		return nil
	}

	suggestions, err := h.client.GetSuggestions(c.Request().Context(), "")
	// TODO: add error status code
	if err != nil {
		return nil
	}

	return web.Render(c, http.StatusOK, components.Suggestions(suggestions))
}
