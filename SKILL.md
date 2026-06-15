---
name: context-distill
description: >
  Distills verbose output before sending it to the LLM, and acts as an MCP client
  that calls any other MCP tool and returns only the compact answer. Use before
  sending long command output, when comparing watch snapshots, before opening many
  files, and — most importantly — when you need to call an MCP tool whose output
  would be too large to pass directly (Jenkins logs, SQL result sets, API payloads).
---

Distill verbose output and call MCP tools without flooding context. Keep signal. Drop noise.

## Core flow

```
agent → distill_mcp_output (context-distill) → any MCP tool → compact answer
```

Context-distill is a **generic MCP client**. It can invoke any stdio MCP server,
call any tool on it, and return only the distilled answer — in one step.
You never see the raw payload. The LLM context stays clean.

## When to use

### Always distill before sending output to the LLM

Use `distill_batch` BEFORE sending any command output longer than 5–8 lines:

- test runs (`go test`, `pytest`, `npm test`)
- builds (`go build`, `npm run build`, `cargo build`)
- linters (`golangci-lint`, `eslint`, `ruff`)
- git commands (`git diff`, `git log`)
- docker / container logs
- any verbose CLI tool

### Call an MCP tool and distill in one step

Use `distill_mcp_output` when:

- you need data from another MCP (Jenkins, SQL, APIs) but the raw output is large
- you want to extract errors from Jenkins build logs without reading 800 lines
- you want a SQL result set summarised as JSON without seeing every row
- you already have a raw MCP payload and only need the compact answer

### Locate code before opening many files

Use `search_code` to find symbols, usages, config loading points, or entrypoints
before reading files.

### Compare two snapshots

Use `distill_watch` when you have two snapshots of the same source (watch mode,
periodic polling) and only need to know what changed.

## Skip distill only when

- output is ≤ 5–8 lines and already human-readable at a glance
- exact raw bytes are required (audit / compliance / binary integrity)
- interactive terminal debugging needs character-by-character flow

## MCP tool: `distill_mcp_output`

This is the key tool for the **agent → context-distill → MCP** pattern.

### Mode A — call an MCP tool and distill in one step

```
distill_mcp_output(
  server_command = "/path/to/mcp-server",
  server_args    = ["--transport", "stdio"],   # optional
  tool_name      = "<tool name>",
  tool_arguments = { ... },                    # optional
  question       = "<output contract>"
)
```

Context-distill launches the MCP server, calls the tool, and returns only the
distilled answer. The raw payload never reaches the LLM.

**Real example — extract errors from a Jenkins build log:**

```
distill_mcp_output(
  server_command = "gaz-mcp",
  server_args    = ["--transport", "stdio"],
  tool_name      = "jenkins_build_log",
  tool_arguments = {
    "environment":  "production",
    "job":          "My Pipeline",
    "build_number": 42,
    "start_line":   0
  },
  question = "Return JSON: {result, total_tests, failing_tests, root_cause, affected_specs}"
)
```

Returns:
```json
{"result":"FAIL","total_tests":9,"failing_tests":9,"root_cause":"ENOTFOUND adslzone.210.test","affected_specs":["metas.cy.js","templates.cy.js","posts/youtube.cy.js"]}
```

Instead of 800 lines of ANSI-coloured Cypress output.

### Mode B — distill a raw MCP payload you already have

```
distill_mcp_output(
  tool_name = "<mcp tool name>",   # optional, adds context to the prompt
  output    = "<raw mcp payload>",
  question  = "<output contract>"
)
```

## MCP tool: `distill_batch`

Distill any command output.

```
distill_batch(
  input    = "<raw output>",
  question = "<output contract>"
)
```

## MCP tool: `distill_watch`

Distill what changed between two snapshots.

```
distill_watch(
  previous_cycle = "<snapshot T-1>",
  current_cycle  = "<snapshot T>",
  question       = "<output contract>"
)
```

## MCP tool: `search_code`

Locate symbols, usages, config points, or paths before opening files.

```
search_code(
  query         = "<text|regex|symbol|path query>",
  mode          = "text | regex | symbol | path",
  question      = "<output contract>",
  scope         = ["optional/glob/**"],   # optional
  max_results   = 20,                     # optional
  context_lines = 2                       # optional
)
```

## Rules

1. Every call MUST include an output contract in `question`.
   Good formats: `PASS or FAIL` · `JSON {severity, file, message}` · `filenames, one per line`
2. One task per call.
3. Prefer machine-checkable formats (JSON, PASS/FAIL, one-item-per-line).

## Quick reference

| Source | `question` |
|---|---|
| `go test ./...` | `"PASS or FAIL. If FAIL, list failing test names, one per line."` |
| `git diff` | `"List only changed file paths, one per line."` |
| CI / build logs | `"JSON array: [{severity, file, message}]"` |
| Jenkins build log (via `distill_mcp_output`) | `"JSON: {result, failing_tests, root_cause, affected_specs}"` |
| SQL result set (via `distill_mcp_output`) | `"Return only the relevant rows as JSON array."` |
| raw MCP payload | `"Return only item names, one per line."` |
| `search_code` symbol lookup | `"Return definitions first as file:line, one per line."` |
