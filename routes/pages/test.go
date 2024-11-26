package pages

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func Test(c echo.Context) error {
	return c.Render(http.StatusOK, "test", nil)
}
