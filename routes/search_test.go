package routes

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/models"
	_ "github.com/DSSD-Madison/gmu/testing_init"
)

func TestSearch(t *testing.T) {

	tests := []struct {
		paramName      string
		paramValue     string
		expectedStatus int
	}{
		{paramName: "query", paramValue: "f", expectedStatus: http.StatusBadRequest},
		{paramName: "query", paramValue: "", expectedStatus: http.StatusOK},
		{paramName: "query", paramValue: "women", expectedStatus: http.StatusOK},
	}

	for _, test := range tests {
		e := echo.New()
		e.Renderer = models.NewTemplate()
		q := make(url.Values)
		q.Set(test.paramName, test.paramValue)
		req := httptest.NewRequest(http.MethodGet, "/search/?"+q.Encode(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := Search(c)
		if err != nil {
			if e, ok := err.(*echo.HTTPError); ok {
				rec.Code = e.Code
			}
		}
		if rec.Code != test.expectedStatus {
			t.Errorf("query=%s - received status incorrect. expected=%d, got=%d", test.paramValue, test.expectedStatus, rec.Code)
		}
	}
}
