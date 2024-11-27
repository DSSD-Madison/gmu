package pages

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/DSSD-Madison/gmu/models"
)

type FiltersResponse struct {
	Filters []models.FilterCategory
}

func Search(c echo.Context) error {
	return c.Render(http.StatusOK, "search", nil)
}