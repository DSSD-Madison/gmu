// File: db/util/getorcreate.go

package util

import (
	"context"
	"database/sql"
	"fmt"

	db "github.com/DSSD-Madison/gmu/internal/infra/database/sqlc/generated"
	"github.com/google/uuid"
)

func GetOrCreateAuthor(ctx context.Context, q *db.Queries, name string) (uuid.UUID, error) {
	author, err := q.FindAuthorByName(ctx, name)
	if err == nil {
		return author.ID, nil
	}
	if err == sql.ErrNoRows {
		newID := uuid.New()
		err := q.InsertAuthor(ctx, db.InsertAuthorParams{
			ID:   newID,
			Name: name,
		})
		if err != nil {
			return uuid.Nil, fmt.Errorf("insert failed for author %q: %w", name, err)
		}
		return newID, nil
	}
	return uuid.Nil, fmt.Errorf("find failed for author %q: %w", name, err)
}

func GetOrCreateKeyword(ctx context.Context, q *db.Queries, name string) (uuid.UUID, error) {
	keyword, err := q.FindKeywordByName(ctx, name)
	if err == nil {
		return keyword.ID, nil
	}
	if err == sql.ErrNoRows {
		newID := uuid.New()
		err := q.InsertKeyword(ctx, db.InsertKeywordParams{
			ID:   newID,
			Name: name,
		})
		if err != nil {
			return uuid.Nil, fmt.Errorf("insert failed for keyword %q: %w", name, err)
		}
		return newID, nil
	}
	return uuid.Nil, fmt.Errorf("find failed for keyword %q: %w", name, err)
}

func GetOrCreateRegion(ctx context.Context, q *db.Queries, name string) (uuid.UUID, error) {
	region, err := q.FindRegionByName(ctx, name)
	if err == nil {
		return region.ID, nil
	}
	if err == sql.ErrNoRows {
		newID := uuid.New()
		err := q.InsertRegion(ctx, db.InsertRegionParams{
			ID:   newID,
			Name: name,
		})
		if err != nil {
			return uuid.Nil, fmt.Errorf("insert failed for region %q: %w", name, err)
		}
		return newID, nil
	}
	return uuid.Nil, fmt.Errorf("find failed for region %q: %w", name, err)
}

func GetOrCreateCategory(ctx context.Context, q *db.Queries, name string) (uuid.UUID, error) {
	category, err := q.FindCategoryByName(ctx, name)
	if err == nil {
		return category.ID, nil
	}
	if err == sql.ErrNoRows {
		newID := uuid.New()
		err := q.InsertCategory(ctx, db.InsertCategoryParams{
			ID:   newID,
			Name: name,
		})
		if err != nil {
			return uuid.Nil, fmt.Errorf("insert failed for category %q: %w", name, err)
		}
		return newID, nil
	}
	return uuid.Nil, fmt.Errorf("find failed for category %q: %w", name, err)
}
