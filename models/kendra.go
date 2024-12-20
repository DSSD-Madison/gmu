package models

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/kendra"

	"github.com/DSSD-Madison/gmu/internal"
)

var opts = kendra.Options{
	Credentials: prov,
	Region:      region,
}

var client = kendra.New(opts)

type KendraResult struct {
	Title   string
	Excerpt string
	Link    string
}

type KendraResults struct {
	QueryOutput kendra.QueryOutput
	Results     []KendraResult
	Query       string
	Count       int
	Filters     []FilterCategory
}

func queryOutputToResults(out kendra.QueryOutput) KendraResults {
	results := KendraResults{
		Results: []KendraResult{},
	}

	for _, item := range out.ResultItems {
		res := KendraResult{
			Title:   internal.TrimExtension(*item.DocumentTitle.Text),
			Excerpt: *item.DocumentExcerpt.Text,
			Link:    *item.DocumentURI,
		}
		results.Results = append(results.Results, res)
		results.Count = int(*out.TotalNumberOfResults)
	}

	return results
}

func MakeQuery(query string) KendraResults {
	kendraQuery := kendra.QueryInput{
		IndexId:   &indexId,
		QueryText: &query,
	}
	out, err := client.Query(context.TODO(), &kendraQuery)

	// TODO: this needs to be fixed to a proper error
	if err != nil {
		log.Printf("Kendra Query Failed %+v", err)
	}

	results := queryOutputToResults(*out)
	results.Query = query
	results.Filters = Filters()
	return results
}
