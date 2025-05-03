package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/internal/application"
	"github.com/DSSD-Madison/gmu/pkg/logger"
	"github.com/DSSD-Madison/gmu/web"
	"github.com/DSSD-Madison/gmu/web/components"
)

type HomeHandler struct {
	log            logger.Logger
	sessionManager application.SessionManager
}

func NewHomeHandler(log logger.Logger, sessionManager application.SessionManager) *HomeHandler {
	handlerLogger := log.With("handler", "Home")
	return &HomeHandler{
		log:            handlerLogger,
		sessionManager: sessionManager,
	}
}

func (h *HomeHandler) Home(c echo.Context) error {
	h.log.InfoContext(c.Request().Context(), "Rendering home page")
	isAuthorized := h.sessionManager.IsAuthenticated(c)
	isMaster := h.sessionManager.IsMaster(c)
	return web.Render(c, http.StatusOK, components.Home(isAuthorized, isMaster))
}
