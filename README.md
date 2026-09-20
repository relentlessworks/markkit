# markkit

Agentic-first Markdown processing service. Convert Markdown to HTML, HTML to Markdown, extract plain text, generate table of contents, and get document statistics. Plain text API, agent-driven, single Go binary.

## Quick Start

```bash
# Build
make build

# Run
./markkit

# Convert Markdown to HTML
curl -X POST localhost:8470/render -d '# Hello World'

# Extract plain text
curl -X POST localhost:8470/extract -d '**bold** and `code`'

# Generate table of contents
curl -X POST localhost:8470/toc -d '# Title
## Section
### Sub'

# Get document statistics
curl -X POST localhost:8470/stats -d '# Hello

Some text here.'

# Convert HTML to Markdown
curl -X POST localhost:8470/html2md -d '<h1>Title</h1><p>Text</p>'

# Get help
curl localhost:8470/help
```

## API Reference

| Method | Path | Description |
|--------|------|-------------|
| POST | /render | Convert Markdown to HTML |
| POST | /extract | Extract plain text from Markdown |
| POST | /toc | Generate table of contents |
| POST | /stats | Get document statistics |
| POST | /html2md | Convert HTML to Markdown |
| POST | /mcp | MCP (Model Context Protocol) endpoint |
| GET | /help | Operating manual for agents |
| GET | /.well-known/agent.md | Same as /help |

## Response Formats

- **Plain text** by default (one labeled line per record)
- **JSON** via `Accept: application/json` header or `?format=json` query param
- **Errors**: `error: message | hint: what to do next`

## Configuration

| Env Var | Flag | Default | Description |
|---------|------|---------|-------------|
| MARKKIT_ADDR | -addr | :8470 | Listen address |
| MARKKIT_SECRET | -secret | auto | Token signing secret |

## Principles

- **Agent IS the interface** — No UI, no SDK. The API is the product.
- **Plain text by default** — Token-cheap, grepable, survives context truncation.
- **Instructive errors** — Every 4xx includes a hint for self-correction.
- **Self-documenting** — `GET /help` returns the operating manual.
- **Single static binary** — Go, CGO_ENABLED=0, zero external dependencies.
- **Zero config** — Runs out of the box with sensible defaults.
- **MCP connector** — Speaks Model Context Protocol at `/mcp`.

## License

MIT
