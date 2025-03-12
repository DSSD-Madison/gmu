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
	requests    = 1
	concurrency = 1
)

func worker(id int, wg *sync.WaitGroup, results chan<- time.Duration, t *testing.T) {
	defer wg.Done()
	start := time.Now()
	data := url.Values{
		"query": {"women mediators"},
	}
	resp, err := http.PostForm(URL, data)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Worker %d: Error %v\n", id, err)
		return
	}
	t.Errorf("Worker %d\n", id)
	resp.Body.Close()
	results <- duration
}

func TestKendra(t *testing.T) {
	var wg sync.WaitGroup
	results := make(chan time.Duration, requests)

	startTime := time.Now()
	t.Errorf("start")

	for i := 0; i < requests; i++ {
		wg.Add(i)
		go func() {
			t.Run(fmt.Sprintf("test: %d", i), func(t *testing.T) { worker(i, &wg, results, t) })
		}()
	}

	wg.Wait()
	close(results)

	var totalDuration time.Duration
	var count int
	var maxDuration time.Duration
	var minDuration time.Duration = time.Hour

	for duration := range results {
		count++
		totalDuration += duration
		if duration > maxDuration {
			maxDuration = duration
		}
		if duration < minDuration {
			minDuration = duration
		}
	}

	fmt.Errorf("\nTotal Requests: %d\n", count)
	fmt.Errorf("Total Time: %v\n", time.Since(startTime))
	fmt.Errorf("Max Response Time: %v\n", maxDuration)
	fmt.Errorf("Min Response Time: %v\n", minDuration)

}
