package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/trace/trace/internal/domain"
	"github.com/trace/trace/internal/domain/graph"
)

// GraphRepository implements ports.CausalGraphRepository using Redis.
// Graph updates are append-only: every Save also writes an immutable snapshot.
type GraphRepository struct {
	client *Client
}

func NewGraphRepository(client *Client) *GraphRepository {
	return &GraphRepository{client: client}
}

func (r *GraphRepository) Get(ctx context.Context, userID string) (*graph.CausalGraph, error) {
	key := GraphKey(userID)
	data, err := r.client.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil, domain.ErrGraphNotFound
	}
	var g graph.CausalGraph
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, fmt.Errorf("graph unmarshal: %w", err)
	}
	return &g, nil
}

// Save writes the graph and an immutable snapshot for this version.
// It uses a pipeline so both writes are atomic from the caller's perspective.
func (r *GraphRepository) Save(ctx context.Context, g *graph.CausalGraph) error {
	g.LastUpdated = time.Now().UnixNano()

	data, err := json.Marshal(g)
	if err != nil {
		return fmt.Errorf("graph marshal: %w", err)
	}

	pipe := r.client.rdb.TxPipeline()
	pipe.Set(ctx, GraphKey(g.UserID), data, 0)
	pipe.Set(ctx, GraphVersionKey(g.UserID), g.Version, 0)
	pipe.Set(ctx, GraphSnapshotKey(g.UserID, g.Version), data, ttlForSnapshot)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *GraphRepository) GetVersion(ctx context.Context, userID string) (int, error) {
	v, err := r.client.rdb.Get(ctx, GraphVersionKey(userID)).Int()
	if err != nil {
		return 0, domain.ErrGraphNotFound
	}
	return v, nil
}

// ttlForSnapshot is how long immutable graph snapshots are retained.
const ttlForSnapshot = 90 * 24 * time.Hour
