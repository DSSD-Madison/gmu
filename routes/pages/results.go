package pages

import (
	"net/http"
	"encoding/json"

	"github.com/labstack/echo/v4"
	"github.com/DSSD-Madison/gmu/models"
)

type PageData struct {
	Filters	[]models.FilterCategory
}

func Results(c echo.Context) error {
	// Fetch filter data from the API
	resp, err := http.Get("http://localhost:8080/api/filters")
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load filters")
	}
	defer resp.Body.Close()

	// Decode JSON response
	var filterResponse struct {
		Filters []models.FilterCategory `json:"Filters"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&filterResponse); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to decode filters")
	}

	// Prepare page data
	data := PageData{
		Filters: filterResponse.Filters,
	}

	return c.Render(http.StatusOK, "results", data)
}
