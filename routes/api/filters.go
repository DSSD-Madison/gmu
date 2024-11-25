package api

import (
	"net/http"
	"github.com/labstack/echo/v4"
	"github.com/DSSD-Madison/gmu/models"
)

func RegisterFiltersRoutes(group *echo.Group) {
	group.GET("/filters", GetFilters)
}

func GetFilters(c echo.Context) error {
	// Example filter data
	filters := map[string]interface{}{
		"Filters": []models.FilterCategory{
			{
				Category: "Authors",
				Options: []models.FilterOption{
					{Label: "Search for Common Ground (SFCG)", Count: 35},
					{Label: "The United States Agency for International Development (USAID)", Count: 32},
					{Label: "Mercy Corps", Count: 8},
				},
			},
			{
				Category: "File Type",
				Options: []models.FilterOption{
					{Label: "PDF", Count: 391},
					{Label: "MS_WORD", Count: 71},
				},
			},
			{
				Category: "Region",
				Options: []models.FilterOption{
					{Label: "Global", Count: 391},
					{Label: "Nepal", Count: 71},
				},
			},
		},
	}

	return c.JSON(http.StatusOK, filters)
}