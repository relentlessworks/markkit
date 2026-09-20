package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/relentlessworks/markkit/internal/markdown"
)

// JSONRPCRequest is a JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

// JSONRPCResponse is a JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError is a JSON-RPC error.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (h *Handler) mcp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, 405, "method not allowed", "use POST with a JSON-RPC 2.0 request body")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONRPCError(w, nil, -32700, "parse error")
		return
	}

	var req JSONRPCRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSONRPCError(w, nil, -32700, "parse error")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch req.Method {
	case "initialize":
		writeJSONRPCResult(w, req.ID, map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":   map[string]interface{}{},
			"serverInfo": map[string]interface{}{
				"name":    "markkit",
				"version": "0.1.0",
			},
		})

	case "tools/list":
		writeJSONRPCResult(w, req.ID, map[string]interface{}{
			"tools": mcpTools(),
		})

	case "tools/call":
		h.handleMCPToolCall(w, &req)

	default:
		writeJSONRPCError(w, req.ID, -32601, "method not found")
	}
}

func mcpTools() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "render",
			"description": "Convert Markdown to HTML",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{
						"type":        "string",
						"description": "Markdown text to convert",
					},
				},
				"required": []string{"text"},
			},
		},
		{
			"name":        "extract",
			"description": "Extract plain text from Markdown, removing all formatting",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{
						"type":        "string",
						"description": "Markdown text to extract from",
					},
				},
				"required": []string{"text"},
			},
		},
		{
			"name":        "toc",
			"description": "Generate a table of contents from Markdown headings",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{
						"type":        "string",
						"description": "Markdown text to generate TOC from",
					},
				},
				"required": []string{"text"},
			},
		},
		{
			"name":        "stats",
			"description": "Get document statistics (word count, char count, reading time, etc.)",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{
						"type":        "string",
						"description": "Markdown text to analyze",
					},
				},
				"required": []string{"text"},
			},
		},
		{
			"name":        "html2md",
			"description": "Convert HTML to Markdown",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{
						"type":        "string",
						"description": "HTML text to convert",
					},
				},
				"required": []string{"text"},
			},
		},
	}
}

func (h *Handler) handleMCPToolCall(w http.ResponseWriter, req *JSONRPCRequest) {
	var params struct {
		Name      string            `json:"name"`
		Arguments map[string]string `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		writeJSONRPCError(w, req.ID, -32602, "invalid params")
		return
	}

	text := params.Arguments["text"]
	if text == "" {
		writeJSONRPCError(w, req.ID, -32602, "missing required argument: text")
		return
	}

	var result string
	switch params.Name {
	case "render":
		result = markdown.Render(text)
	case "extract":
		result = markdown.ExtractText(text)
	case "toc":
		headings := markdown.TOC(text)
		var sb strings.Builder
		for _, h2 := range headings {
			indent := strings.Repeat("  ", h2.Level-1)
			fmt.Fprintf(&sb, "%slevel=%d text=%q slug=%s\n", indent, h2.Level, h2.Text, h2.Slug)
		}
		result = sb.String()
	case "stats":
		s := markdown.ComputeStats(text)
		result = fmt.Sprintf("words=%d chars=%d runes=%d lines=%d paragraphs=%d headings=%d code_blocks=%d links=%d images=%d list_items=%d blockquotes=%d reading_time_min=%d",
			s.Words, s.Chars, s.Runes, s.Lines, s.Paragraphs, s.Headings,
			s.CodeBlocks, s.Links, s.Images, s.ListItems, s.Blockquotes, s.ReadingTime)
	case "html2md":
		result = markdown.HTMLToMarkdown(text)
	default:
		writeJSONRPCError(w, req.ID, -32602, fmt.Sprintf("unknown tool: %s", params.Name))
		return
	}

	writeJSONRPCResult(w, req.ID, map[string]interface{}{
		"content": []map[string]string{
			{"type": "text", "text": result},
		},
	})
}

func writeJSONRPCResult(w http.ResponseWriter, id interface{}, result interface{}) {
	json.NewEncoder(w).Encode(JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	})
}

func writeJSONRPCError(w http.ResponseWriter, id interface{}, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &RPCError{Code: code, Message: msg},
	})
}
