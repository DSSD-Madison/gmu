package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/pkg/logger"
	"github.com/DSSD-Madison/gmu/pkg/middleware"
	"github.com/DSSD-Madison/gmu/web"
	"github.com/DSSD-Madison/gmu/web/components"
)

type HomeHandler struct {
	log logger.Logger
}

func NewHomeHandler(log logger.Logger) *HomeHandler {
	handlerLogger := log.With("handler", "Home")
	return &HomeHandler{log: handlerLogger}
}

func (h *HomeHandler) Home(c echo.Context) error {
	h.log.InfoContext(c.Request().Context(), "Rendering home page")
	isAuthorized, isMaster := middleware.GetSessionFlags(c)
	return web.Render(c, http.StatusOK, components.Home(isAuthorized, isMaster))
}
