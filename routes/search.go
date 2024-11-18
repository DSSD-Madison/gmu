package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/models"
)

func Search(c echo.Context) error {
	models.Data.ResultsCount = len(models.Data.Results)
	models.Data.Query = c.FormValue("query")
	return c.Render(http.StatusOK, "results-page", models.Data)
}
