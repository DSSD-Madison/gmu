package routes

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func Test(c echo.Context) error {
	time.Sleep(time.Second * 5)

	return c.Render(http.StatusOK, "test", nil)
}
