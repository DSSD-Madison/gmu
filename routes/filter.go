package routes

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/models"
)

var active = false

func Filter(c echo.Context) error {
	newData := models.Data
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid ID")
	}

	for i, filter := range newData.Filters {
		if filter.ID == id {
			newData.Filters[i].Active = !filter.Active
		}
	}

	var res []models.Result
	active = !active
	if active {
		for _, result := range newData.Results {
			if result.Id%2 == 0 {
				res = append(res, result)
			}
		}
		newData.Results = res
	}

	newData.ResultsCount = len(newData.Results)
	newData.Query = c.FormValue("query")
	err = c.Render(http.StatusOK, "sidecolumn", newData)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "results", newData)
}
