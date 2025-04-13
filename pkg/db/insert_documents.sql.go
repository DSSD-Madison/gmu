// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0
// source: insert_documents.sql

package db

import (
	"context"

	"github.com/google/uuid"
)

const insertDocAuthor = `-- name: InsertDocAuthor :exec
INSERT INTO doc_authors (id, doc_id, author_id)
VALUES ($1, $2, $3)
`

type InsertDocAuthorParams struct {
	ID       uuid.UUID
	DocID    uuid.NullUUID
	AuthorID uuid.NullUUID
}

func (q *Queries) InsertDocAuthor(ctx context.Context, arg InsertDocAuthorParams) error {
	_, err := q.db.ExecContext(ctx, insertDocAuthor, arg.ID, arg.DocID, arg.AuthorID)
	return err
}

const insertDocCategory = `-- name: InsertDocCategory :exec
INSERT INTO doc_categories (id, doc_id, category_id)
VALUES ($1, $2, $3)
`

type InsertDocCategoryParams struct {
	ID         uuid.UUID
	DocID      uuid.NullUUID
	CategoryID uuid.NullUUID
}

func (q *Queries) InsertDocCategory(ctx context.Context, arg InsertDocCategoryParams) error {
	_, err := q.db.ExecContext(ctx, insertDocCategory, arg.ID, arg.DocID, arg.CategoryID)
	return err
}

const insertDocKeyword = `-- name: InsertDocKeyword :exec
INSERT INTO doc_keywords (id, doc_id, keyword_id)
VALUES ($1, $2, $3)
`

type InsertDocKeywordParams struct {
	ID        uuid.UUID
	DocID     uuid.NullUUID
	KeywordID uuid.NullUUID
}

func (q *Queries) InsertDocKeyword(ctx context.Context, arg InsertDocKeywordParams) error {
	_, err := q.db.ExecContext(ctx, insertDocKeyword, arg.ID, arg.DocID, arg.KeywordID)
	return err
}

const insertDocRegion = `-- name: InsertDocRegion :exec
INSERT INTO doc_regions (id, doc_id, region_id)
VALUES ($1, $2, $3)
`

type InsertDocRegionParams struct {
	ID       uuid.UUID
	DocID    uuid.NullUUID
	RegionID uuid.NullUUID
}

func (q *Queries) InsertDocRegion(ctx context.Context, arg InsertDocRegionParams) error {
	_, err := q.db.ExecContext(ctx, insertDocRegion, arg.ID, arg.DocID, arg.RegionID)
	return err
}

const insertUploadedDocument = `-- name: InsertUploadedDocument :exec
INSERT INTO documents (
  id,
  s3_file,
  file_name,
  title,
  created_at,
  has_duplicate
) VALUES ($1, $2, $3, $4, NOW(), false)
`

type InsertUploadedDocumentParams struct {
	ID       uuid.UUID
	S3File   string
	FileName string
	Title    string
}

func (q *Queries) InsertUploadedDocument(ctx context.Context, arg InsertUploadedDocumentParams) error {
	_, err := q.db.ExecContext(ctx, insertUploadedDocument,
		arg.ID,
		arg.S3File,
		arg.FileName,
		arg.Title,
	)
	return err
}
