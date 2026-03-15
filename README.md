# context-distill

A Go MCP server that **distills command output** before it is sent to another LLM, inspired by the `distill` CLI and implemented with hexagonal architecture, DI, and TDD.

## Overview

`context-distill` exposes two MCP tools:
- `distill_batch`: compresses full command output to answer a specific question.
- `distill_watch`: compares two consecutive cycles and reports only relevant changes.

It also includes:
- LLM provider configuration via YAML/env.
- Interactive terminal configuration UI with `tview`.
- Support for `ollama` and OpenAI-compatible providers.

## Features

- Hexagonal architecture (`distill/domain`, `distill/application`, `platform/*`).
- Dependency injection with `sarulabs/di`.
- Config management with `viper` + `.env`.
- Provider-specific config validation.
- Interactive setup UI (`--config-ui`).
- Unit/integration tests + optional live test.

## Requirements

- Go `1.26+`
- Make (recommended)

If you prefer, you can install from GitHub Releases without compiling (see next section).

## Install Without Compiling

Install latest release binary:

- Linux/macOS:

```bash
curl -fsSL https://raw.githubusercontent.com/jcastilloa/context-distill/master/scripts/install.sh | sh
```

- Windows (PowerShell):

```powershell
iwr https://raw.githubusercontent.com/jcastilloa/context-distill/master/scripts/install.ps1 -UseBasicParsing | iex
```

Install a specific version:

- Linux/macOS:

```bash
curl -fsSL https://raw.githubusercontent.com/jcastilloa/context-distill/master/scripts/install.sh | VERSION=v0.1.0 sh
```

- Windows (PowerShell):

```powershell
$env:VERSION='v0.1.0'; iwr https://raw.githubusercontent.com/jcastilloa/context-distill/master/scripts/install.ps1 -UseBasicParsing | iex
```

Optional environment variables:
- `REPO` (default: `jcastilloa/context-distill`)
- `SERVICE_NAME` (default: `context-distill`)
- `INSTALL_DIR` (default Linux/macOS: `~/.local/bin`, Windows: `%LOCALAPPDATA%\context-distill\bin`)
- `VERSION` (default: latest release tag)

## Makefile

This repository already ships with a `Makefile`:

```bash
make help
```

### Available targets

| Target | Description |
|---|---|
| `make build` | Build the binary to `./bin/context-distill` |
| `make install` | Install binary to `~/.local/bin/context-distill` |
| `make clean` | Remove `./bin` |

## Quick Start

```bash
make build
./bin/context-distill --config-ui
./bin/context-distill --transport stdio
```

That is enough to build, configure, and start the MCP server.

## MCP Installation

### 1. Build

```bash
make build
```

Output:
- `./bin/context-distill`

### 1b. Or install a prebuilt binary (no build)

- Linux/macOS:

```bash
curl -fsSL https://raw.githubusercontent.com/jcastilloa/context-distill/master/scripts/install.sh | sh
```

- Windows (PowerShell):

```powershell
iwr https://raw.githubusercontent.com/jcastilloa/context-distill/master/scripts/install.ps1 -UseBasicParsing | iex
```

### 2. (Optional) Install to local PATH

```bash
make install
```

Output:
- `~/.local/bin/context-distill`

### 3. Register in your MCP client

#### Option A: JSON `mcpServers` clients

If your client uses a JSON file with `mcpServers`, add:

```json
{
  "mcpServers": {
    "context-distill": {
      "command": "/absolute/path/to/context-distill/bin/context-distill",
      "args": ["--transport", "stdio"]
    }
  }
}
```

#### Option B: Codex (`~/.codex/config.toml`) - recommended

Add this block to `~/.codex/config.toml`:

```toml
[mcp_servers.context-distill]
command = "/absolute/path/to/context-distill/bin/context-distill"
args = ["--transport", "stdio"]
startup_timeout_sec = 20.0
```

If you used `make install`, you can use:

```toml
[mcp_servers.context-distill]
command = "/home/<your-user>/.local/bin/context-distill"
args = ["--transport", "stdio"]
startup_timeout_sec = 20.0
```

#### Option C: Codex CLI registration (no manual TOML edit)

```bash
codex mcp add context-distill -- /absolute/path/to/context-distill/bin/context-distill --transport stdio
```

Verify in Codex:

```bash
codex mcp list
codex mcp get context-distill
```

Then restart your Codex session so it reloads MCP servers.

### MCP registration notes

- Use an absolute binary path.
- Keep `stdio` transport.
- If the server does not appear, run `codex mcp list --json` to inspect final config.

## MCP Configuration via Terminal UI

The UI is designed for fast local setup and persistence.

### Launch UI

```bash
./bin/context-distill --config-ui
# or without building:
go run ./cmd/server --config-ui
```

### Editable fields

- `provider_name` (dropdown list)
- `base_url`
- `api_key` (masked input)

### UI validation rules

- If provider requires API key (`openai`, `openrouter`, `jan`), save is blocked without key.
- If provider is OpenAI-compatible, `base_url` is required.
- Provider aliases are normalized (`OpenAI Compatible` -> `openai-compatible`, `dmr` -> `docker-model-runner`).

### Persisted config path

- `~/.config/context-distill/config.yaml`

Save behavior:
- Preserves existing YAML sections (`service`, `openai`, etc.).
- Updates only relevant `distill` fields.

## Manual Configuration (YAML)

You can also edit config manually.

### Config lookup order

1. `~/.config/<service>/config.yaml`
2. `./config.yaml`

### Example `config.yaml`

```yaml
service:
  transport: stdio
  version: 0.1.0

openai:
  provider_name: openai
  api_key: sk-xxxx
  base_url: https://api.openai.com/v1
  model: gpt-4o-mini
  timeout: 30s
  max_retries: 3
  supports_system_role: true
  supports_json_mode: true

distill:
  provider_name: ollama
  base_url: http://127.0.0.1:11434
  model: qwen3.5:2b
  timeout: 90s
  max_retries: 0
  thinking: false
```

## Provider Matrix

| Provider | Transport | API Key Required | Default Base URL |
|---|---|---|---|
| `ollama` | native ollama | No | `http://127.0.0.1:11434` |
| `openai` | openai-compatible | Yes | `https://api.openai.com/v1` |
| `openrouter` | openai-compatible | Yes | `https://openrouter.ai/api/v1` |
| `openai-compatible` | openai-compatible | No (backend-dependent) | no forced default |
| `lmstudio` | openai-compatible | No | `http://127.0.0.1:1234/v1` |
| `jan` | openai-compatible | Yes | `http://127.0.0.1:1337/v1` |
| `localai` | openai-compatible | No | `http://127.0.0.1:8080/v1` |
| `vllm` | openai-compatible | No | `http://127.0.0.1:8000/v1` |
| `sglang` | openai-compatible | No | no forced default |
| `llama.cpp` | openai-compatible | No | no forced default |
| `mlx-lm` | openai-compatible | No | no forced default |
| `docker-model-runner` | openai-compatible | No | `http://127.0.0.1:12434/engines/v1` |

## Running the MCP Server

### Run locally

```bash
./bin/context-distill --transport stdio
# or without building:
go run ./cmd/server --transport stdio
```

### Version

```bash
go run ./cmd/server version
```

### CLI flags

| Flag | Description | Default |
|---|---|---|
| `--transport` | MCP transport (`stdio`) | `service.transport` |
| `--config-ui` | Open setup UI and exit | `false` |

## MCP Tools

### Tool: `distill_batch`

Input:
- `question` (string, required)
- `input` (string, required)

Output:
- Short distilled answer focused on `question`.

### Tool: `distill_watch`

Input:
- `question` (string, required)
- `previous_cycle` (string, required)
- `current_cycle` (string, required)

Output:
- Short summary of relevant changes.
- Returns no-change message when nothing relevant changed.

## Concrete Usage Guidance (Important)

To get high-quality distillation, the `question` must be explicit and constrained.

### Good vs bad `question`

Bad:
- `"What happened?"`
- `"Summarize this"`

Good:
- `"Did tests pass? Return only PASS or FAIL, then failing test names if FAIL."`
- `"List only files that changed. One file path per line."`
- `"Return valid JSON only with keys: severity, file, message."`

### Practical examples for `distill_batch`

- Input source: `go test ./...` output.
  Question: `"Did tests pass? Return only PASS or FAIL."`
- Input source: `git diff` output.
  Question: `"List changed files and one short reason per file. Max 10 lines."`
- Input source: CI logs.
  Question: `"Return only blocking errors with file and line if available."`

### Practical examples for `distill_watch`

- Previous cycle: test watcher output at T-1.
- Current cycle: test watcher output at T.
  Question: `"What changed in failures count? Return one short sentence."`

- Previous cycle: deployment status snapshot.
- Current cycle: deployment status snapshot.
  Question: `"Return only newly failing services. One per line."`

## AGENTS.md Template For Projects Using This MCP

Add a section like this in the consumer project’s `AGENTS.md`.
The goal is to make usage consistent: explicit questions, explicit output contracts, and clear rules for when to distill vs when to keep raw output.

```md
## context-distill MCP Usage

CRITICAL: When command output is large and would otherwise be sent to a paid LLM, call `distill_batch` first.

CRITICAL: The `question` must be explicit about output format.
Examples:
- "Return only PASS or FAIL."
- "Return valid JSON only."
- "Return only filenames, one per line."

CRITICAL: For recurring/watch-like command output, compare cycles with `distill_watch`:
- `previous_cycle`: prior snapshot
- `current_cycle`: latest snapshot
- ask only for relevant deltas

Do not use raw output unless exact uncompressed output is required.

### Mandatory patterns

- Always include an output contract in `question`.
- Keep the question scoped to one task.
- Prefer machine-checkable formats when possible (PASS/FAIL, JSON, one-item-per-line).

### `distill_batch` examples

Input source: `go test ./...` output
Tool call:
- question: "Did tests pass? Return only PASS or FAIL. If FAIL, append failing test names."
- input: "<full go test output>"

Input source: `git diff` output
Tool call:
- question: "List only changed file paths, one per line."
- input: "<full git diff output>"

Input source: CI logs
Tool call:
- question: "Return valid JSON only with keys: severity, file, message."
- input: "<full CI log output>"

### `distill_watch` examples

Use when output is periodic or watch-like.

Tool call:
- question: "What changed in failure count? One short sentence."
- previous_cycle: "<cycle T-1>"
- current_cycle: "<cycle T>"

Tool call:
- question: "Return only newly failing services, one per line."
- previous_cycle: "<deployment status at T-1>"
- current_cycle: "<deployment status at T>"

### When NOT to distill

- You need exact raw output for compliance/audit.
- You need full stack traces without compression.
- You are debugging an interactive prompt exchange where exact terminal flow matters.
```

### Suggested policy line for engineering teams

Use this one-liner policy in project docs:

```md
Default policy: distill command output with `context-distill` before sending logs/traces/diffs to an LLM, unless raw output is explicitly required.
```

## AGENTS.md Template (Strict CI Mode)

Use this variant when the consumer project runs automated pipelines and requires deterministic, machine-parseable output.

```md
## context-distill MCP Usage (CI Mode)

CRITICAL: For large command output consumed by automation, call `distill_batch` first.

CRITICAL: Every `question` must define an explicit output contract and MUST be machine-parseable.
- Prefer JSON objects or arrays only.
- No markdown.
- No prose outside the requested schema.

CRITICAL: If JSON is requested, enforce:
- "Return valid JSON only."
- fixed keys and fixed value shapes.

### Standard contracts

- Test status:
  Question: "Return valid JSON only with keys: status, failing_tests. status must be PASS or FAIL."
- Lint status:
  Question: "Return valid JSON only with keys: status, issues. issues must be an array of {file,line,message}."
- Diff summary:
  Question: "Return valid JSON only with key files_changed as an array of file paths."

### `distill_watch` in CI

Use `distill_watch` only for periodic snapshots with strict delta output.

Question example:
- "Return valid JSON only with keys: changed, added, removed. Each must be an array of strings."

### Failure handling

- If the distillation output does not match the requested schema, treat it as invalid and re-run with a stricter question.
- If exact raw output is needed for audit/compliance, bypass distillation.
```

## Project Structure

```text
context-distill/
├── cmd/
│   └── server/
│       ├── main.go
│       ├── bootstrap.go
│       └── openai_distill_config.go
├── distill/
│   ├── application/
│   │   └── distillation/
│   └── domain/
├── mcp/
│   ├── application/
│   └── domain/
├── platform/
│   ├── config/
│   ├── configui/
│   ├── di/
│   ├── mcp/
│   │   ├── commands/
│   │   ├── server/
│   │   └── tools/
│   ├── ollama/
│   └── openai/
├── shared/
│   ├── ai/domain/
│   └── config/domain/
├── config.sample.yaml
├── config.yaml
├── Makefile
└── AGENTS.md
```

## Architecture

Dependency rule:

```text
platform -> shared + distill/application + distill/domain
distill/application -> distill/domain
cmd -> platform + shared
```

Constraint:
- `shared` and `distill/domain` must not import `platform`.

## Development and Quality

### Tests and static checks

```bash
go test ./...
go vet ./...
```

### Suggested local workflow

1. `go test ./...`
2. `./bin/context-distill --config-ui`
3. `./bin/context-distill --transport stdio`
4. Validate behavior from your MCP client

## Optional Live Test (real provider)

```bash
DISTILL_LIVE_TEST=1 OPENAI_BASE_URL=https://openrouter.ai/api/v1 \
go test -tags=live ./platform/di -run TestLiveDistillBatchWithOpenAICompatibleProvider -v
```

## Troubleshooting

### `provider unauthorized`

Check:
- `distill.api_key`
- or fallback `openai.api_key` depending on provider.

### `requires base_url`

Set `distill.base_url` (fastest path is `--config-ui`).

### MCP client does not detect server

Check:
- absolute binary path;
- executable permissions;
- `stdio` transport in client config.

### Server fails due to config validation

Run:
- `--config-ui` for initial setup;
- then run normally (`./bin/context-distill --transport stdio`).

## Security

- Do not commit real API keys to public repositories.
- Prefer environment-based secrets in shared environments.

## License

Internal/private project (adjust to your final license model).
