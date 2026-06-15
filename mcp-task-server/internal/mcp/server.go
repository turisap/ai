package mcp

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
)

const protocolVersion = "2024-11-05"

// Server is the MCP HTTP server.
// One instance serves all clients; each client connection gets its own Session.
type Server struct {
	name     string
	version  string
	registry *Registry

	sessions sync.Map // session ID → *Session
}

func NewServer(name, version string, registry *Registry) *Server {
	return &Server{name: name, version: version, registry: registry}
}

// Handler returns an http.Handler to mount at your chosen path (e.g. /mcp).
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", s.handleMCP)
	return mux
}

// handleMCP dispatches GET (SSE stream) and POST (JSON-RPC message).
func (s *Server) handleMCP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleSSE(w, r)
	case http.MethodPost:
		s.handlePost(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleSSE opens a long-lived SSE stream for the client.
// The client sends its session ID as ?sessionId=... on subsequent POSTs.
func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	session, err := NewSession(w)
	if err != nil {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	s.sessions.Store(session.ID, session)
	defer func() {
		s.sessions.Delete(session.ID)
		session.Close()
	}()

	slog.Info("client connected", "session_id", session.ID)

	// Send the endpoint event so the client knows where to POST messages.
	endpointData, _ := json.Marshal(map[string]string{
		"uri": "/mcp?sessionId=" + session.ID,
	})
	if err := session.SendEvent("endpoint", endpointData); err != nil {
		return
	}

	// Block until the client disconnects.
	<-r.Context().Done()
	slog.Info("client disconnected", "session_id", session.ID)
}

// handlePost receives a JSON-RPC request, dispatches it, and responds.
// For the Streamable HTTP transport the response goes back in the HTTP body
// (not over SSE), which keeps request/response simple for tool calls.
func (s *Server) handlePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, ErrorResponse(nil, ErrParse, "parse error"))
		return
	}

	slog.Debug("received", "method", req.Method, "id", req.ID)

	resp := s.dispatch(r.Context(), req)
	writeJSON(w, resp)
}

// dispatch routes a JSON-RPC request to the correct handler.
func (s *Server) dispatch(ctx context.Context, req Request) Response {
	switch req.Method {

	case "initialize":
		return s.handleInitialize(req)

	case "initialized":
		// Notification — no response needed, but we must not return an error.
		return Response{} // caller skips writing empty responses

	case "tools/list":
		return Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  ToolsListResult{Tools: s.registry.List()},
		}

	case "tools/call":
		var params ToolCallParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return ErrorResponse(req.ID, ErrInvalidParams, "invalid params")
		}
		result, err := s.registry.Dispatch(ctx, params.Name, params.Arguments)
		if err != nil {
			return ErrorResponse(req.ID, ErrInternal, err.Error())
		}
		return Response{JSONRPC: "2.0", ID: req.ID, Result: result}

	case "ping":
		return Response{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{}}

	default:
		return ErrorResponse(req.ID, ErrMethodNotFound, "method not found: "+req.Method)
	}
}

func (s *Server) handleInitialize(req Request) Response {
	// We accept any protocol version the client sends; real servers
	// should negotiate down to the highest mutually supported version.
	return Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: InitializeResult{
			ProtocolVersion: protocolVersion,
			ServerInfo: ImplementationInfo{
				Name:    s.name,
				Version: s.version,
			},
			Capabilities: ServerCapabilities{
				Tools: &ToolsCapability{ListChanged: false},
			},
		},
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	// Skip empty responses (notifications like "initialized").
	if v == (Response{}) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to write response", "err", err)
	}
}
