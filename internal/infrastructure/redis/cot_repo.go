package redis

import (
	"context"
	"encoding/json"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/trace/trace/internal/application/ports"
	"github.com/trace/trace/internal/domain"
)

const cotTTL = 30 * 24 * time.Hour

// CoTTraceRepository implements ports.CoTTraceRepository with a 30-day TTL.
type CoTTraceRepository struct {
	client *Client
}

func NewCoTTraceRepository(client *Client) *CoTTraceRepository {
	return &CoTTraceRepository{client: client}
}

func (r *CoTTraceRepository) Save(ctx context.Context, trace ports.CoTTrace) error {
	data, err := json.Marshal(trace)
	if err != nil {
		return err
	}
	return r.client.rdb.Set(ctx, CoTTraceKey(trace.TraceID), data, cotTTL).Err()
}

func (r *CoTTraceRepository) Get(ctx context.Context, traceID string) (*ports.CoTTrace, error) {
	data, err := r.client.rdb.Get(ctx, CoTTraceKey(traceID)).Bytes()
	if err != nil {
		if err == goredis.Nil {
			return nil, domain.ErrMCPToolNotFound
		}
		return nil, err
	}
	var t ports.CoTTrace
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// VetoEventRepository implements ports.VetoEventRepository.
type VetoEventRepository struct {
	client *Client
}

func NewVetoEventRepository(client *Client) *VetoEventRepository {
	return &VetoEventRepository{client: client}
}

func (r *VetoEventRepository) Save(ctx context.Context, event ports.VetoEvent) (string, error) {
	vetoID := "veto-" + event.UserID + "-" + event.SessionDate
	data, err := json.Marshal(event)
	if err != nil {
		return "", err
	}
	pipe := r.client.rdb.TxPipeline()
	pipe.Set(ctx, VetoEventKey(vetoID), data, 30*24*time.Hour)
	pipe.RPush(ctx, VetoIndexKey(event.UserID), vetoID)
	_, err = pipe.Exec(ctx)
	return vetoID, err
}

func (r *VetoEventRepository) GetByUser(ctx context.Context, userID string, limit int) ([]ports.VetoEvent, error) {
	ids, err := r.client.rdb.LRange(ctx, VetoIndexKey(userID), 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}
	events := make([]ports.VetoEvent, 0, len(ids))
	for _, id := range ids {
		data, err := r.client.rdb.Get(ctx, VetoEventKey(id)).Bytes()
		if err != nil {
			continue
		}
		var e ports.VetoEvent
		if json.Unmarshal(data, &e) == nil {
			events = append(events, e)
		}
	}
	return events, nil
}
