package db_types

import db "github.com/DSSD-Madison/gmu/pkg/db/generated"

type PDFMetadata struct {
	FileId        string
	Document      db.FindDocumentByIDRow
	AllAuthors    []db.Author
	AllKeywords   []db.Keyword
	AllRegions    []db.Region
	AllCategories []db.Category
	AuthorNames   []string
	KeywordNames  []string
	RegionNames   []string
	CategoryNames []string
}
