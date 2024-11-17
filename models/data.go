package models

var Ipsum = "Lorem ipsum odor amet, consectetuer adipiscing elit. Lorem felis mi ex senectus..."
var Data = NewData()
var id = 0

type Result struct {
	ImgPath     string
	Description string
	Title       string
	Id          int
}

func NewResult(title, desc, path string) Result {
	id++
	return Result{Title: title, Description: desc, ImgPath: path, Id: id}
}

type Filter struct {
	Name        string
	ID          int
	Description string
	Active      bool
}

func NewFilter(id int, name, description string) Filter {
	return Filter{Name: name, ID: id, Description: description, Active: false}
}

type Results = []Result
type Filters = []Filter

type DataStruct struct {
	Results      Results
	Filters      Filters
	Query        string
	ResultsCount int
}

func NewData() DataStruct {
	return DataStruct{
		Results: []Result{
			NewResult("MyTitle2", Ipsum, "images/cptt.jpg"),
			NewResult("MyTitle3", Ipsum, "images/crafting_interpreters.jpg"),
		},
		Filters: []Filter{NewFilter(0, "Filter1", "Description1")},
		Query:   "",
	}
}

type Book struct {
	Title       string
	Author      string
	Description string
	Img         string
	Id          int
}

func NewBook() Book {
	return Book{
		Title:       "Title",
		Author:      "Author",
		Description: Ipsum,
		Img:         "images/cptt.jpg",
	}
}
