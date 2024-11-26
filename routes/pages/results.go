package pages

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type FilterOption struct {
	Label	string
	Count	int
}

type FilterCategory struct {
	Category	string
	Options	[]FilterOption
}

type PageData struct {
	Where	string
	Filters	[]FilterCategory
}

func Results(c echo.Context) error {
	filters := []FilterCategory{
		{
			Category: "Authors",
			Options: []FilterOption{
				{Label: "Search for Common Ground (SFCG)", Count: 35},
				{Label: "The United States Agency for International Development (USAID)", Count: 32},
				{Label: "Mercy Corps", Count: 8},
			},
		},
		{
			Category: "File Type",
			Options: []FilterOption{
				{Label: "PDF", Count: 391},
				{Label: "MS_WORD", Count: 71},
			},
		},
		{
			Category: "Region",
			Options: []FilterOption{
				{Label: "Global", Count: 391},
				{Label: "Nepal", Count: 71},
			},
		},
	}


	data := PageData{
		Where: "Over here",
		Filters: filters,
	}

	return c.Render(http.StatusOK, "results", data)
}
