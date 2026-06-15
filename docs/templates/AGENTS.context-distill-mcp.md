## context-distill MCP Usage

Default: ALWAYS distill.

Use:
- `distill_batch` for verbose command output
- `distill_mcp_output` for raw MCP payloads or one-step MCP call + distillation
- `distill_watch` for snapshot deltas
- `search_code` before opening many files

### Rules

1. Every call MUST include an output contract in `question`.
2. One task per call.
3. Prefer machine-checkable formats (PASS/FAIL, JSON, one-per-line).

### `distill_batch` examples

| Source | `question` |
|---|---|
| `go test ./...` | "Did all tests pass? PASS or FAIL. If FAIL, list failing test names, one per line." |
| `git diff` | "List only changed file paths, one per line." |
| CI / build logs | "Return valid JSON array: [{severity, file, message}]." |

### `distill_mcp_output` examples

| Case | Parameters |
|---|---|
| Raw MCP payload already available | `tool_name="list_items", output="<payload>", question="Return only item names, one per line."` |
| Call MCP tool and distill in one step | `server_command="/absolute/path/to/mcp-server", tool_name="get_status", tool_arguments={"scope":"production"}, question="Return only status and version as JSON."` |

### `distill_watch`

Use when you have two snapshots of the same source and only care about the delta.

### `search_code`

Use before opening many files when locating symbols, config loading, entrypoints, or provider wiring.

Example:
`query="LoadDistillConfig", mode="symbol", question="Return likely definitions first as file:line, one per line."`

### When to skip distill

- Output <= 5-8 lines and already readable.
- Exact raw bytes are required.
- Interactive terminal debugging needs exact character flow.
