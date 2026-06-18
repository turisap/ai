package tools

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kirillshakirov/mcp-task-server/internal/mcp"
)

var toolCountUserTasks = mcp.Tool{
	Name:        "get_count_user_tasks",
	Description: "Count tasks assigned to a user",
	InputSchema: mcp.InputSchema{
		Type: "object",
		Properties: map[string]mcp.Property{
			"user_id": {Type: "string", Description: "UUID of the user"},
		},
		Required: []string{"user_id"},
	},
}

type getCountUserTasksArgs struct {
	UserID string `json:"user_id"`
}

func handleGetCountUserTasksWith(db *pgxpool.Pool) mcp.HandlerFunc {
	return func(ctx context.Context, raw json.RawMessage) (mcp.ToolCallResult, error) {
		var args getCountUserTasksArgs
		if err := json.Unmarshal(raw, &args); err != nil {
			return mcp.ErrorResult("invalid arguments"), nil
		}

		var count int
		err := db.QueryRow(ctx,
			`SELECT count(*) FROM tasks WHERE assignee_id = $1`,
			args.UserID,
		).Scan(&count)
		if err != nil {
			return mcp.ErrorResult("db error: " + err.Error()), nil
		}

		out, _ := json.Marshal(count)
		return mcp.TextResult(string(out)), nil
	}
}
