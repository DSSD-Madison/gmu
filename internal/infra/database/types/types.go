package db_types

import db "github.com/DSSD-Madison/gmu/internal/infra/database/sqlc/generated"

type PDFMetadata struct {
	FileID        string
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
