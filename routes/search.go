package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/models"
)

func LoadSearch(c echo.Context) error {
	query := c.FormValue("query")
	if len(query) < 3 {
		return c.String(http.StatusBadRequest, "Error 400, could not get query from request.")
	}
	return c.Render(http.StatusOK, "results-init", query)
}

func Search(c echo.Context) error {
	query := c.FormValue("query")
	if len(query) < 3 {
		return c.String(http.StatusBadRequest, "Error 400, could not get query from request.")
	}
	results := models.MakeQuery(query)
	return c.Render(http.StatusOK, "results", results)
}
