package models

type Region struct {
	id   int
	name string
}

type ImageLink struct {
	pdf string
	s3  string
}

type Author struct {
	id   int
	name string
}

type WrittenBy struct {
	id        int
	articleId int
	authorId  int
}

type Keyword struct {
	id      int
	keyword string
}

type KeywordRef struct {
	id        int
	articleId int
	keywordId int
}

type Document struct {
	id          int
	title       string
	abstract    string
	regionId    int
	category    string
	publishDate string // should be Date
	source      string
	imageLinks  ImageLink
	modified    string // should be Date
	created     string // should be Date
}

func newDocument() Document {
	return Document{
		id:          0,
		title:       "title",
		abstract:    "abstract",
		regionId:    0,
		category:    "category",
		publishDate: "Date",
		source:      "source",
		imageLinks:  ImageLink{pdf: "pdf_link", s3: "s3_link"},
		modified:    "last modified",
		created:     "created",
	}
}
