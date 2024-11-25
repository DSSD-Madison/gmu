package pages

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func Results(c echo.Context) error {
	return c.Render(http.StatusOK, "results", nil)
}
