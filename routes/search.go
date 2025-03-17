package routes

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/db"
	"github.com/DSSD-Madison/gmu/handlers"
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

func Search(c echo.Context, queries *db.Queries) error {
	query := c.FormValue("query")
	fmt.Printf("query: %s\n", query)
	fmt.Printf("resp: %+v\n", c.Request().Header)

	if len(query) == 0 {
		return Home(c)
	}

	if len(query) < MinQueryLength {
		return echo.NewHTTPError(http.StatusBadRequest, "Query too short")
	}
	// Check if the request is coming from HTMX
	target := c.Request().Header.Get("HX-Target")

	if target == "root" || target == "" {
		return c.Render(http.StatusOK, "search-standalone", query)
	} else if target == "results-container" {
		fmt.Println("results doing")
		results := models.MakeQuery(query, nil)
		uris, err := ConvertToS3URIs(results)
		if err != nil {
			return err
		}
		documentMap, err := handlers.GetDocuments(c, queries, uris)
		if err != nil {
			fmt.Println("what is the eror")
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		fmt.Println(fmt.Sprintf("documentMap:", documentMap))
		fmt.Printf("Number of results before processing: %d\n", len(results.Results))

		for key, kendraResult := range results.Results {
			// Step 3: Convert the link to S3 URI
			fmt.Println("in the loop")
			s3URI, err := ConvertToS3URI(kendraResult.Link)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("invalid URL for result %s: %s", key, err.Error())})
			}
			fmt.Println("before step 4")
			// Step 4: Check if the document exists in the documentMap
			document, found := documentMap[s3URI]
			if !found {
				fmt.Println("not found")
				fmt.Println("tried with s3URIof")
				fmt.Println(s3URI)
				return c.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("document not found for S3 URI: %s", s3URI)})
			}
	
			// Step 5: Add S3FilePreview to KendraResult
			fmt.Println("before step 5")
			if document.S3FilePreview.Valid {
				image, err := ConvertS3URIToURL(document.S3FilePreview.String)
				if err != nil {
					return err;
				}
				kendraResult.Image = image; // Assign the valid value
			} else {
				kendraResult.Image = "" // Assign empty string or handle null accordingly
			}
			results.Results[key] = kendraResult // Update the KendraResult with the new field
			fmt.Println("after the last step")
		}
		fmt.Println("start printing")
		for _, kendraResult := range results.Results {
			fmt.Println(kendraResult.Image);
		}
		fmt.Println("end printing")
		
		return c.Render(http.StatusOK, "results", results)
	} else {
		return c.Render(http.StatusOK, "search", query)
	}

}
