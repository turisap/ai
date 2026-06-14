package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

// HandlerFunc is the signature every tool handler must implement.
type HandlerFunc func(ctx context.Context, args json.RawMessage) (ToolCallResult, error)

// Registry holds all registered tools and dispatches tools/call.
type Registry struct {
	tools    []Tool
	handlers map[string]HandlerFunc
}

func NewRegistry() *Registry {
	return &Registry{handlers: make(map[string]HandlerFunc)}
}

// Register adds a tool definition + its handler to the registry.
func (r *Registry) Register(tool Tool, handler HandlerFunc) {
	r.tools = append(r.tools, tool)
	r.handlers[tool.Name] = handler
}

// List returns all registered tool definitions (for tools/list).
func (r *Registry) List() []Tool {
	return r.tools
}

// Dispatch executes the named tool (for tools/call).
func (r *Registry) Dispatch(ctx context.Context, name string, args json.RawMessage) (ToolCallResult, error) {
	handler, ok := r.handlers[name]
	if !ok {
		return ToolCallResult{}, fmt.Errorf("unknown tool: %s", name)
	}
	return handler(ctx, args)
}
