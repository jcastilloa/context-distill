---
name: context-distill
description: >
  Distills verbose command output and retrieves compact code context before sending
  payloads to an LLM. Use before sending long command output, when comparing watch
  snapshots, before opening many files, and when you need to distill MCP tool output
  or invoke an MCP tool and distill its result in one step.
---

Distill verbose CLI output and retrieve compact code context before passing to LLM. Keep signal. Drop noise.

## Activation

Use BEFORE sending any command output longer than 5-8 lines to LLM.

Use AFTER:
- tests
- builds
- linters
- git commands
- docker logs
- any verbose CLI tool

Use when:
- comparing two snapshots of the same source in watch mode
- unsure whether to distill
- locating symbols, usages, config loading, or entrypoints before opening many files
- you already have raw MCP tool output and want to distill it
- you want to call an MCP tool and distill the result in one step

Default rule: **always distill**. Unnecessary distill cost is low. Flooding context is expensive.

## Skip

Do not use when:
- output is <= 5-8 lines and already human-readable
- exact raw bytes are required (audit / compliance / binary integrity)
- interactive terminal debugging needs character-by-character flow

## Commands

### Distill full output

```bash
# Pipe - preferred
<command> | context-distill distill_batch --question "<question with output contract>"

# Explicit flag
context-distill distill_batch --question "<question with output contract>" --input "<raw output>"

# Explicit stdin marker
<command> | context-distill distill_batch --question "<question with output contract>" --input -
```

### Distill raw MCP payload you already have

```bash
context-distill distill_mcp_output \
  --tool-name "<mcp tool name>" \
  --question "<question with output contract>" \
  --output "<raw mcp payload>"
```

Use this when another tool already returned an MCP payload and you only need the compact answer.

### Call an MCP tool and distill in one step

```bash
context-distill distill_mcp_output \
  --server-command "/absolute/path/to/server" \
  --server-arg "--transport" \
  --server-arg "stdio" \
  --tool-name "<mcp tool name>" \
  --tool-arguments '{"key":"value"}' \
  --question "<question with output contract>"
```

Use this when you want one command to do both steps:
1. call the MCP tool
2. distill the result

Example:

```bash
context-distill distill_mcp_output \
  --server-command "/absolute/path/to/mcp-server" \
  --server-arg "--transport" \
  --server-arg "stdio" \
  --tool-name "get_status" \
  --tool-arguments '{"scope":"production"}' \
  --question "Return only status and version as JSON."
```

### Distill delta between two snapshots

```bash
context-distill distill_watch \
  --question "<question with output contract>" \
  --previous-cycle "<snapshot T-1>" \
  --current-cycle "<snapshot T>"
```

### Locate code before opening many files

```bash
context-distill search_code \
  --query "<text|regex|symbol|path query>" \
  --mode "<text|regex|symbol|path>" \
  --question "<output contract>" \
  --scope "<optional glob list>" \
  --max-results 20 \
  --context-lines 2
```

Use CLI flags only. Do not use shell args like `search_code mode=text query=...`.

## Rules

1. Every call MUST include an output contract in `--question`.
   Good formats:
   - `PASS or FAIL`
   - `valid JSON {severity, file, message}`
   - `filenames, one per line`
2. One task per call.
3. Prefer machine-checkable formats.

## Examples

| Source | Question |
|---|---|
| `go test ./...` | `"Did all tests pass? Return only PASS or FAIL. If FAIL, list failing test names, one per line."` |
| `git diff` | `"List only changed file paths, one per line."` |
| CI / build logs | `"Return valid JSON array: [{severity, file, message}]."` |
| raw MCP list payload | `"Return only item names, one per line."` |
| direct MCP call to `get_status` | `"Return only status and version as JSON."` |
| `search_code` symbol lookup | `"Return likely definitions first as file:line, one per line."` |
