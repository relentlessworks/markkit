package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/relentlessworks/markkit/internal/markdown"
)

// Handler holds shared state for HTTP handlers.
type Handler struct {
	Secret string
}

// New creates a new API handler.
func New(secret string) *Handler {
	return &Handler{Secret: secret}
}

// Routes returns the HTTP mux with all routes registered.
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/help", h.help)
	mux.HandleFunc("/.well-known/agent.md", h.help)
	mux.HandleFunc("/render", h.render)
	mux.HandleFunc("/extract", h.extract)
	mux.HandleFunc("/toc", h.toc)
	mux.HandleFunc("/stats", h.stats)
	mux.HandleFunc("/html2md", h.html2md)
	mux.HandleFunc("/mcp", h.mcp)
	mux.HandleFunc("/", h.root)
	return mux
}

// wantsJSON checks if the client wants JSON responses.
func wantsJSON(r *http.Request) bool {
	if r.URL.Query().Get("format") == "json" {
		return true
	}
	return strings.Contains(r.Header.Get("Accept"), "application/json")
}

// writeError writes an instructive error response.
func writeError(w http.ResponseWriter, status int, msg, hint string) {
	w.WriteHeader(status)
	fmt.Fprintf(w, "error: %s | hint: %s\n", msg, hint)
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

// readBody reads the request body as a string.
func readBody(r *http.Request) (string, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (h *Handler) root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	fmt.Fprintln(w, "markkit — agentic-first Markdown processing service")
	fmt.Fprintln(w, "GET /help for the operating manual")
}

func (h *Handler) help(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, helpText)
}

const helpText = `# markkit — Agentic-First Markdown Processing Service

Convert Markdown to HTML, HTML to Markdown, extract plain text, generate
table of contents, and get document statistics. Plain text API, agent-driven.

## Endpoints

POST /render
  Body: Markdown text
  Returns: HTML output (plain text) or JSON {html: "..."}
  Example: curl -X POST localhost:8470/render -d '# Hello'
  Response: <h1 id="hello">Hello</h1>

POST /extract
  Body: Markdown text
  Returns: Plain text with formatting removed
  Example: curl -X POST localhost:8470/extract -d '**bold** and ` + "`code`" + `'
  Response: bold and code

POST /toc
  Body: Markdown text
  Returns: Table of contents (one heading per line)
  Format: level=1 text="Hello" slug="hello"
  JSON: [{level:1, text:"Hello", slug:"hello"}]

POST /stats
  Body: Markdown text
  Returns: Document statistics
  Format: words=42 chars=200 lines=10 paragraphs=3 headings=2 ...
  JSON: {words:42, chars:200, lines:10, ...}

POST /html2md
  Body: HTML text
  Returns: Markdown output
  Example: curl -X POST localhost:8470/html2md -d '<h1>Title</h1>'
  Response: # Title

POST /mcp
  Body: JSON-RPC 2.0 request
  Tools: render, extract, toc, stats, html2md

GET /help or /.well-known/agent.md
  Returns: This operating manual

## Response Formats
- Plain text by default (one labeled line per record)
- JSON via Accept: application/json header or ?format=json query param
- Errors: error: message | hint: what to do next

## Configuration
- MARKKIT_ADDR (default :8470) — listen address
- MARKKIT_SECRET — token signing secret (auto-generated if empty)

## Notes
- No database needed — pure stateless computation
- No auth required for basic usage
- Single Go binary, CGO_ENABLED=0, zero external dependencies
`

func (h *Handler) render(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, 405, "method not allowed", "use POST to submit Markdown text")
		return
	}
	body, err := readBody(r)
	if err != nil {
		writeError(w, 400, "failed to read body", "ensure the request body contains Markdown text")
		return
	}
	if strings.TrimSpace(body) == "" {
		writeError(w, 400, "empty body", "provide Markdown text in the request body")
		return
	}
	html := markdown.Render(body)
	if wantsJSON(r) {
		writeJSON(w, map[string]string{"html": html})
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

func (h *Handler) extract(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, 405, "method not allowed", "use POST to submit Markdown text")
		return
	}
	body, err := readBody(r)
	if err != nil {
		writeError(w, 400, "failed to read body", "ensure the request body contains Markdown text")
		return
	}
	if strings.TrimSpace(body) == "" {
		writeError(w, 400, "empty body", "provide Markdown text in the request body")
		return
	}
	text := markdown.ExtractText(body)
	if wantsJSON(r) {
		writeJSON(w, map[string]string{"text": text})
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, text)
}

func (h *Handler) toc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, 405, "method not allowed", "use POST to submit Markdown text")
		return
	}
	body, err := readBody(r)
	if err != nil {
		writeError(w, 400, "failed to read body", "ensure the request body contains Markdown text")
		return
	}
	if strings.TrimSpace(body) == "" {
		writeError(w, 400, "empty body", "provide Markdown text in the request body")
		return
	}
	headings := markdown.TOC(body)
	if wantsJSON(r) {
		writeJSON(w, headings)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	for _, h2 := range headings {
		indent := strings.Repeat("  ", h2.Level-1)
		fmt.Fprintf(w, "%slevel=%d text=%q slug=%s\n", indent, h2.Level, h2.Text, h2.Slug)
	}
}

func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, 405, "method not allowed", "use POST to submit Markdown text")
		return
	}
	body, err := readBody(r)
	if err != nil {
		writeError(w, 400, "failed to read body", "ensure the request body contains Markdown text")
		return
	}
	if strings.TrimSpace(body) == "" {
		writeError(w, 400, "empty body", "provide Markdown text in the request body")
		return
	}
	s := markdown.ComputeStats(body)
	if wantsJSON(r) {
		writeJSON(w, s)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "words=%d chars=%d runes=%d lines=%d paragraphs=%d headings=%d code_blocks=%d links=%d images=%d list_items=%d blockquotes=%d reading_time_min=%d\n",
		s.Words, s.Chars, s.Runes, s.Lines, s.Paragraphs, s.Headings,
		s.CodeBlocks, s.Links, s.Images, s.ListItems, s.Blockquotes, s.ReadingTime)
}

func (h *Handler) html2md(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, 405, "method not allowed", "use POST to submit HTML text")
		return
	}
	body, err := readBody(r)
	if err != nil {
		writeError(w, 400, "failed to read body", "ensure the request body contains HTML text")
		return
	}
	if strings.TrimSpace(body) == "" {
		writeError(w, 400, "empty body", "provide HTML text in the request body")
		return
	}
	md := markdown.HTMLToMarkdown(body)
	if wantsJSON(r) {
		writeJSON(w, map[string]string{"markdown": md})
		return
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	fmt.Fprint(w, md)
}
