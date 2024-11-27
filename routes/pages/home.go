package pages

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func Home(c echo.Context) error {
	return c.Render(http.StatusOK, "home", nil)
}
