package routes

import (
	"fmt"
	"net/http"
	"time"

	"github.com/DSSD-Madison/gmu/db"
	"github.com/DSSD-Madison/gmu/models"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	db *db.Queries
	q models.QueryExecutor
}

func NewHandler(db *db.Queries, q models.QueryExecutor) Handler {
	return Handler{
		db: db,
		q: q,
	}
}

// InitRoutes registers all the application routes
func InitRoutes(e *echo.Echo, h Handler) {
	e.GET("/", Home)

	e.GET("/search", func(c echo.Context) error {
		return Search(c, h)
	})
	e.POST("/search/suggestions", SearchSuggestions)
	// test route for testing the queue, other testing
	// e.GET("/test", test)
}

var queue = func() *models.KendraQueue[string, interface{}] {
	q := models.NewKendraQueue[string, interface{}](2, 4)
	return q
}()

func test(c echo.Context) error {
	resultChan := make(chan interface{})
	param := c.QueryParam("query")

	job := models.Job[string, interface{}]{
		Payload: param,
		Callback: func(payload string) {
			fmt.Println("Job result: ", payload)
			time.Sleep(time.Second * 2)
			resultChan <- payload
		},
	}
	queue.Enqueue(job)

	result := <- resultChan

	return c.String(http.StatusOK, fmt.Sprintf("Result: %+v", result))
}
