package routes

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/models"
)

func Book(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid ID")
	}

	book := models.NewBook()
	book.Id = id
	return c.Render(http.StatusOK, "book-page", book)
}
