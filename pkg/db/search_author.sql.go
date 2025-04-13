// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0
// source: search_author.sql

package db

import (
	"context"
	"database/sql"
)

const searchAuthorsByNamePrefix = `-- name: SearchAuthorsByNamePrefix :many
SELECT
    id,
    name
FROM
    authors
WHERE
    name ILIKE $1 || '%'  -- Case-insensitive prefix search
ORDER BY
    name -- Optional: order results alphabetically
    LIMIT 10
`

func (q *Queries) SearchAuthorsByNamePrefix(ctx context.Context, dollar_1 sql.NullString) ([]Author, error) {
	rows, err := q.db.QueryContext(ctx, searchAuthorsByNamePrefix, dollar_1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Author
	for rows.Next() {
		var i Author
		if err := rows.Scan(&i.ID, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
