package models

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	_ "github.com/DSSD-Madison/gmu/testing_init"
)

const (
	URL         = "https://gmustaging.savepointsoftware.com"
	localURL    = "http://localhost:8080"
	requests    = 1
	concurrency = 1
)

func worker(id int, wg *sync.WaitGroup, results chan<- time.Duration, t *testing.T) {
	defer wg.Done()
	start := time.Now()
	t.Log("start")
	data := url.Values{
		"query": {"women mediators"},
	}
	resp, err := http.PostForm(URL, data)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("Worker %d: Error %v\n", id, err)
		return
	}
	resp.Body.Close()
	results <- duration
}

func workerTest(id int, wg *sync.WaitGroup, results chan<- time.Duration, t *testing.T) {
	defer wg.Done()
	start := time.Now()
	time.Sleep(1 * time.Second)
	duration := time.Since(start)

	results <- duration
}

func TestKendra(t *testing.T) {
	t.Parallel()
}
