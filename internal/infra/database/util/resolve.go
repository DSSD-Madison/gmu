package util

import (
	"context"
	"log"
	"strings"

	db "github.com/DSSD-Madison/gmu/internal/infra/database/sqlc/generated"
	"github.com/google/uuid"
)

// ResolverFunc defines the signature for resolving a name to a UUID.
// It's typically a getOrCreate function like GetOrCreateAuthor, etc.
type ResolverFunc func(ctx context.Context, q *db.Queries, name string) (uuid.UUID, error)

// ResolveIDs processes a list of raw form values and resolves them to UUIDs.
// - If the value starts with "new:", it creates the item via the resolver.
// - Otherwise, it expects a valid UUID string.
func ResolveIDs(ctx context.Context, q *db.Queries, values []string, resolver ResolverFunc) []uuid.UUID {
	var ids []uuid.UUID

	for _, raw := range values {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		if strings.HasPrefix(raw, "new:") {
			name := strings.TrimPrefix(raw, "new:")
			id, err := resolver(ctx, q, name)
			if err != nil {
				log.Printf("[ERROR] resolving %q: %v", name, err)
				continue
			}
			ids = append(ids, id)
		} else {
			u, err := uuid.Parse(raw)
			if err != nil {
				log.Printf("[WARN] Skipping invalid UUID: %s", raw)
				continue
			}
			ids = append(ids, u)
		}
	}

	return ids
}
