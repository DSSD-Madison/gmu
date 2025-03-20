package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) Home(c echo.Context) error {
	return c.Render(http.StatusOK, "index", nil)
}
