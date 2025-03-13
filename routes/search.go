package routes

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/models"
)

const MinQueryLength = 3

func SearchSuggestions(c echo.Context) error {
	query := c.FormValue("query")

	if len(query) == 0 {
		return nil
	}
	suggestions, err := models.GetSuggestions(query)
	// TODO: add error status code
	if err != nil {
		return nil
	}
	return c.Render(http.StatusOK, "suggestions", suggestions)
}

type Job struct {
	ID       int
	Response chan models.KendraResults
}

const maxWorkers = 2

var (
	activeWorkers int32
	semaphore     = make(chan struct{}, maxWorkers)
)

func worker(job Job, query string) {
	defer func() {
		<-semaphore
		atomic.AddInt32(&activeWorkers, -1)
	}()

	// fmt.Printf("worker processing job: %d", job.ID)
	// time.Sleep(2 * time.Second)
	// result := fmt.Sprintf("Job %d completed!", job.ID)
	// fmt.Printf("worker finished job %d\n", job.ID)

	result := models.MakeQuery(query, nil)

	job.Response <- result
}

func Search(c echo.Context) error {

	query := c.FormValue("query")
	context := c.FormValue("context")

	if len(query) < MinQueryLength {
		return echo.NewHTTPError(http.StatusBadRequest, "Query too short")
	}
	if context == "home" {
		return c.Render(http.StatusOK, "search", query)
	}

	jobID := time.Now().UnixNano()
	respChan := make(chan models.KendraResults, 1)
	job := Job{
		ID:       int(jobID),
		Response: respChan,
	}

	semaphore <- struct{}{}
	atomic.AddInt32(&activeWorkers, 1)

	go worker(job, query)

	results := <-respChan

	// results := models.MakeQuery(query, nil)

	return c.Render(http.StatusOK, "results", results)
	// return c.String(http.StatusOK, result)
}
