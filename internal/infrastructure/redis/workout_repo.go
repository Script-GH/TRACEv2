package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/trace/trace/internal/domain"
	"github.com/trace/trace/internal/domain/workout"
)

const _ = time.Hour // keep import for plan/log TTLs

// WorkoutLogRepository implements ports.WorkoutLogRepository.
// Uses a sorted set (score = session timestamp) as an index for efficient
// "last N logs" queries without a full scan.
type WorkoutLogRepository struct {
	client *Client
}

func NewWorkoutLogRepository(client *Client) *WorkoutLogRepository {
	return &WorkoutLogRepository{client: client}
}

func (r *WorkoutLogRepository) Save(ctx context.Context, log *workout.Log) error {
	data, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("log marshal: %w", err)
	}
	pipe := r.client.rdb.TxPipeline()
	pipe.Set(ctx, WorkoutLogKey(log.UserID, log.LogID), data, 30*24*time.Hour)
	pipe.ZAdd(ctx, WorkoutLogIndexKey(log.UserID), goredis.Z{Score: float64(log.CreatedAt), Member: log.LogID})
	_, err = pipe.Exec(ctx)
	return err
}

func (r *WorkoutLogRepository) GetHistory(ctx context.Context, userID string, n int) ([]workout.Log, error) {
	ids, err := r.client.rdb.ZRevRange(ctx, WorkoutLogIndexKey(userID), 0, int64(n-1)).Result()
	if err != nil {
		return nil, err
	}
	logs := make([]workout.Log, 0, len(ids))
	for _, id := range ids {
		data, err := r.client.rdb.Get(ctx, WorkoutLogKey(userID, id)).Bytes()
		if err != nil {
			continue
		}
		var l workout.Log
		if json.Unmarshal(data, &l) == nil {
			logs = append(logs, l)
		}
	}
	return logs, nil
}

func (r *WorkoutLogRepository) GetByID(ctx context.Context, logID string) (*workout.Log, error) {
	// logID format: userID:logID — for simplicity, scan isn't needed if we store by full key
	// Callers pass the full logID stored during Save.
	data, err := r.client.rdb.Get(ctx, logID).Bytes()
	if err != nil {
		return nil, domain.ErrWorkoutLogNotFound
	}
	var l workout.Log
	if err := json.Unmarshal(data, &l); err != nil {
		return nil, err
	}
	return &l, nil
}

// WorkoutPlanRepository implements ports.WorkoutPlanRepository.
// Plans are versioned and never overwritten — each date can have multiple plan versions.
type WorkoutPlanRepository struct {
	client *Client
}

func NewWorkoutPlanRepository(client *Client) *WorkoutPlanRepository {
	return &WorkoutPlanRepository{client: client}
}

func (r *WorkoutPlanRepository) Save(ctx context.Context, plan *workout.Plan) error {
	data, err := json.Marshal(plan)
	if err != nil {
		return fmt.Errorf("plan marshal: %w", err)
	}
	dateStr := plan.ForDate
	pipe := r.client.rdb.TxPipeline()
	pipe.Set(ctx, WorkoutPlanKey(plan.UserID, plan.PlanID), data, 7*24*time.Hour)
	pipe.Set(ctx, WorkoutPlanByDateKey(plan.UserID, dateStr), plan.PlanID, 7*24*time.Hour)
	// latest pointer — overwritten on each new plan
	pipe.Set(ctx, fmt.Sprintf("trace:%s:plan:latest", plan.UserID), plan.PlanID, 7*24*time.Hour)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *WorkoutPlanRepository) GetForDate(ctx context.Context, userID string, date string) (*workout.Plan, error) {
	planID, err := r.client.rdb.Get(ctx, WorkoutPlanByDateKey(userID, date)).Result()
	if err != nil {
		if err == goredis.Nil {
			return nil, domain.ErrWorkoutPlanNotFound
		}
		return nil, err
	}
	return r.getPlanByID(ctx, userID, planID)
}

func (r *WorkoutPlanRepository) GetLatest(ctx context.Context, userID string) (*workout.Plan, error) {
	planID, err := r.client.rdb.Get(ctx, fmt.Sprintf("trace:%s:plan:latest", userID)).Result()
	if err != nil {
		if err == goredis.Nil {
			return nil, domain.ErrWorkoutPlanNotFound
		}
		return nil, err
	}
	return r.getPlanByID(ctx, userID, planID)
}

func (r *WorkoutPlanRepository) getPlanByID(ctx context.Context, userID, planID string) (*workout.Plan, error) {
	data, err := r.client.rdb.Get(ctx, WorkoutPlanKey(userID, planID)).Bytes()
	if err != nil {
		return nil, domain.ErrWorkoutPlanNotFound
	}
	var p workout.Plan
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// ProgressionHistoryRepository implements ports.ProgressionHistoryRepository.
type ProgressionHistoryRepository struct {
	client *Client
}

func NewProgressionHistoryRepository(client *Client) *ProgressionHistoryRepository {
	return &ProgressionHistoryRepository{client: client}
}

func (r *ProgressionHistoryRepository) GetHistory(ctx context.Context, userID string, exercise string) ([]workout.ProgressionEntry, error) {
	key := ProgressionKey(userID, exercise)
	raw, err := r.client.rdb.LRange(ctx, key, 0, 49).Result()
	if err != nil || len(raw) == 0 {
		return nil, nil
	}
	entries := make([]workout.ProgressionEntry, 0, len(raw))
	for _, s := range raw {
		var e workout.ProgressionEntry
		if json.Unmarshal([]byte(s), &e) == nil {
			entries = append(entries, e)
		}
	}
	return entries, nil
}

func (r *ProgressionHistoryRepository) Save(ctx context.Context, userID string, entry workout.ProgressionEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	key := ProgressionKey(userID, entry.LogID)
	return r.client.rdb.LPush(ctx, key, string(data)).Err()
}
