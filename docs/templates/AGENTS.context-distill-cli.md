## context-distill CLI Usage

Default behavior: distill command output before sending it to the LLM.
Use `search_code` before opening many files to locate symbols, usages, config paths, and entrypoints.

Use the local binary:
- `context-distill distill_batch`
- `context-distill distill_mcp_output`
- `context-distill distill_watch`
- `context-distill search_code`

### Rules

1. Every invocation MUST include an output contract in `--question`.
2. One task per call.
3. Prefer machine-checkable formats (PASS/FAIL, JSON, one-item-per-line).

### CLI patterns

#### Batch output

```bash
go test ./... 2>&1 | context-distill distill_batch \
  --question "Did all tests pass? Return only PASS or FAIL. If FAIL, list failing tests one per line."
```

#### MCP payload you already have

```bash
context-distill distill_mcp_output \
  --tool-name "list_items" \
  --question "Return only item names, one per line." \
  --output '<mcp payload>'
```

#### MCP call + distillation in one step

```bash
context-distill distill_mcp_output \
  --server-command "/absolute/path/to/mcp-server" \
  --server-arg "--transport" \
  --server-arg "stdio" \
  --tool-name "get_status" \
  --tool-arguments '{"scope":"production"}' \
  --question "Return only status and version as JSON."
```

#### Snapshot delta

```bash
context-distill distill_watch \
  --question "Return only newly failing services, one per line." \
  --previous-cycle "$(cat /tmp/status.prev)" \
  --current-cycle "$(cat /tmp/status.curr)"
```

#### Code retrieval

```bash
context-distill search_code \
  --query "provider_name" \
  --mode text \
  --question "Return only file:line, one per line."
```

### When to skip distill

- Output is <= 5-8 lines and readable at a glance.
- Exact raw bytes are required.
- Interactive terminal debugging needs exact character flow.
