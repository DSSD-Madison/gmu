package pages

import (
	"net/http"
	// "encoding/json"
	// "log"

	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/models"
)

// func Search(c echo.Context) error {
// 	query := c.FormValue("query")
// 	if len(query) < 3 {
// 		return c.String(http.StatusBadRequest, "Error 400, could not get query from request.")
// 	}
// 	results := models.MakeQuery(query)
// 	return c.Render(http.StatusOK, "search", results)
// }

type FiltersResponse struct {
	Filters []models.FilterCategory
}

// func Search(c echo.Context) error {
// 	// Fetch filter data from the API
// 	resp, err := http.Get("http://localhost:8080/api/filters")
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	// Decode JSON into the imported struct
// 	var filtersResponse FiltersResponse
// 	if err := json.NewDecoder(resp.Body).Decode(&filtersResponse); err != nil {
// 		log.Println("Error unmarshaling body:", err)
// 		return err
// 	}

// 	log.Println("Filters Response:", filtersResponse)

// 	return c.Render(http.StatusOK, "search", map[string]interface{}{
// 		"Filters": filtersResponse.Filters,
// 	})
// }

func Search(c echo.Context) error {
	return c.Render(http.StatusOK, "search", nil)
}