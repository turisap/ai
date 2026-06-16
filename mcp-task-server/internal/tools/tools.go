package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/kirillshakirov/mcp-task-server/internal/mcp"
)

// RegisterAll wires tool definitions to their real handlers.
// db and rdb are injected here; each handler closes over what it needs.
func RegisterAll(registry *mcp.Registry, db *pgxpool.Pool, rdb *redis.Client) {
	registry.Register(toolGetTasks, handleGetTasksWith(db))
	registry.Register(toolGetTaskCounters, handleGetTaskCountersWith(rdb))
	registry.Register(toolCreateTask, handleCreateTaskWith(db))
}

// ── get_tasks ────────────────────────────────────────────────────────────────

var toolGetTasks = mcp.Tool{
	Name:        "get_tasks",
	Description: "List tasks for a store. Optionally filter by status.",
	InputSchema: mcp.InputSchema{
		Type: "object",
		Properties: map[string]mcp.Property{
			"store_id": {Type: "string", Description: "UUID of the store"},
			"status":   {Type: "string", Description: "Filter by status: open, in_progress, done (optional)"},
			"limit":    {Type: "integer", Description: "Max results, default 20"},
		},
		Required: []string{"store_id"},
	},
}

type getTasksArgs struct {
	StoreID string `json:"store_id"`
	Status  string `json:"status"`
	Limit   int    `json:"limit"`
}

type taskRow struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Status     string `json:"status"`
	AssigneeID string `json:"assignee_id,omitempty"`
}

func handleGetTasksWith(db *pgxpool.Pool) mcp.HandlerFunc {
	return func(ctx context.Context, raw json.RawMessage) (mcp.ToolCallResult, error) {
		var args getTasksArgs
		if err := json.Unmarshal(raw, &args); err != nil {
			return mcp.ErrorResult("invalid arguments"), nil
		}
		if args.Limit == 0 {
			args.Limit = 20
		}

		// $2 = '' short-circuits the status filter when not provided.
		rows, err := db.Query(ctx,
			`SELECT id, title, status, COALESCE(assignee_id::text, '')
			 FROM tasks
			 WHERE store_id = $1 AND ($2 = '' OR status = $2::task_status)
			 ORDER BY created_at DESC
			 LIMIT $3`,
			args.StoreID, args.Status, args.Limit,
		)
		if err != nil {
			return mcp.ErrorResult("db error: " + err.Error()), nil
		}
		defer rows.Close()

		var results []taskRow
		for rows.Next() {
			var r taskRow
			if err := rows.Scan(&r.ID, &r.Title, &r.Status, &r.AssigneeID); err != nil {
				return mcp.ErrorResult("scan error: " + err.Error()), nil
			}
			results = append(results, r)
		}
		if err := rows.Err(); err != nil {
			return mcp.ErrorResult("rows error: " + err.Error()), nil
		}

		out, _ := json.Marshal(results)
		return mcp.TextResult(string(out)), nil
	}
}

// ── get_task_counters ────────────────────────────────────────────────────────

var toolGetTaskCounters = mcp.Tool{
	Name:        "get_task_counters",
	Description: "Get task counters for a store from Redis (open, in_progress, done totals).",
	InputSchema: mcp.InputSchema{
		Type: "object",
		Properties: map[string]mcp.Property{
			"store_id": {Type: "string", Description: "UUID of the store"},
		},
		Required: []string{"store_id"},
	},
}

type getCountersArgs struct {
	StoreID string `json:"store_id"`
}

func handleGetTaskCountersWith(rdb *redis.Client) mcp.HandlerFunc {
	return func(ctx context.Context, raw json.RawMessage) (mcp.ToolCallResult, error) {
		var args getCountersArgs
		if err := json.Unmarshal(raw, &args); err != nil {
			return mcp.ErrorResult("invalid arguments"), nil
		}

		// Your CDC projection stores counters as a hash:
		//   key:   counters:store:<store_id>
		//   fields: open, in_progress, done
		key := "counters:store:" + args.StoreID
		vals, err := rdb.HGetAll(ctx, key).Result()
		if err != nil {
			return mcp.ErrorResult("redis error: " + err.Error()), nil
		}
		if len(vals) == 0 {
			return mcp.TextResult(fmt.Sprintf("no counters found for store %s", args.StoreID)), nil
		}

		out, _ := json.Marshal(vals)
		return mcp.TextResult(string(out)), nil
	}
}

// ── create_task ──────────────────────────────────────────────────────────────

var toolCreateTask = mcp.Tool{
	Name:        "create_task",
	Description: "Create a new task for a store. Mutation goes through the outbox for reliability.",
	InputSchema: mcp.InputSchema{
		Type: "object",
		Properties: map[string]mcp.Property{
			"store_id":    {Type: "string", Description: "UUID of the store"},
			"title":       {Type: "string", Description: "Task title"},
			"assignee_id": {Type: "string", Description: "UUID of the assignee (optional)"},
		},
		Required: []string{"store_id", "title"},
	},
}

type createTaskArgs struct {
	StoreID    string `json:"store_id"`
	Title      string `json:"title"`
	AssigneeID string `json:"assignee_id"`
}

type outboxPayload struct {
	TaskID     string `json:"task_id"`
	StoreID    string `json:"store_id"`
	Title      string `json:"title"`
	AssigneeID string `json:"assignee_id,omitempty"`
}

func handleCreateTaskWith(db *pgxpool.Pool) mcp.HandlerFunc {
	return func(ctx context.Context, raw json.RawMessage) (mcp.ToolCallResult, error) {
		var args createTaskArgs
		if err := json.Unmarshal(raw, &args); err != nil {
			return mcp.ErrorResult("invalid arguments"), nil
		}

		taskID := uuid.New()

		payload, _ := json.Marshal(outboxPayload{
			TaskID:     taskID.String(),
			StoreID:    args.StoreID,
			Title:      args.Title,
			AssigneeID: args.AssigneeID,
		})

		// Single transaction: insert task + outbox row atomically.
		// If Kafka relay is down, the outbox worker will retry — task creation
		// is never lost.
		tx, err := db.Begin(ctx)
		if err != nil {
			return mcp.ErrorResult("tx begin: " + err.Error()), nil
		}
		defer tx.Rollback(ctx)

		_, err = tx.Exec(ctx,
			`INSERT INTO tasks (id, store_id, title, assignee_id, status, created_at)
			 VALUES ($1, $2, $3, NULLIF($4, '')::uuid, 'open', now())`,
			taskID, args.StoreID, args.Title, args.AssigneeID,
		)
		if err != nil {
			return mcp.ErrorResult("insert task: " + err.Error()), nil
		}

		_, err = tx.Exec(ctx,
			`INSERT INTO outbox (aggregate_id, event_type, payload, created_at)
			 VALUES ($1, 'task.created', $2, now())`,
			taskID, payload,
		)
		if err != nil {
			return mcp.ErrorResult("insert outbox: " + err.Error()), nil
		}

		if err := tx.Commit(ctx); err != nil {
			return mcp.ErrorResult("tx commit: " + err.Error()), nil
		}

		return mcp.TextResult(fmt.Sprintf(`{"task_id":%q,"status":"created"}`, taskID)), nil
	}
}
