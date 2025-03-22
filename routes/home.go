package routes

import (
	"net/http"

	"github.com/DSSD-Madison/gmu/components"
	"github.com/DSSD-Madison/gmu/models"
	"github.com/labstack/echo/v4"
)

func (h *Handler) Home(c echo.Context) error {
	return models.Render(c, http.StatusOK, components.Home())
}
