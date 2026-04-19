# context-distill

A Go tool that **distills command output and retrieves code context** before it reaches a paid LLM. Available as a **Skill** (recommended), a **standalone CLI**, and an **MCP server**. Inspired by the `distill` CLI and built with hexagonal architecture, dependency injection, and TDD.

## Overview

`context-distill` exposes three operations accessible in three ways:

| Mode | Best for | How it works |
|---|---|---|
| **Skill** ⭐ (recommended) | Any agent that reads markdown and runs shell commands | The agent reads a `SKILL.md` file from its skills directory and learns when and how to invoke the CLI. Zero config. |
| **CLI** | Local scripts, CI pipelines, shell-capable agents | Direct subcommands: `distill_batch`, `distill_watch`, `search_code`. |
| **MCP** | Agents with native MCP support (Claude Desktop, Cursor, Codex…) | MCP server over `stdio` transport. |

| Operation | Purpose |
|---|---|
| `distill_batch` | Compresses full command output to answer a single, explicit question. |
| `distill_watch` | Compares two consecutive snapshots and returns only the relevant delta. |
| `search_code` | Locates relevant repository code and returns compact matches. |

All three modes share the same use cases, validation rules, and output behavior — only the invocation method differs.

Additionally:

- LLM provider configuration via YAML and environment variables.
- Interactive terminal UI for first-time setup (`--config-ui`).
- Support for Ollama and any OpenAI-compatible provider.

### Why Skill Mode Is Recommended

| | Skill | CLI | MCP |
|---|---|---|---|
| Agent config required | **None** — drop `SKILL.md` in the agent's skills directory | Agent must know how to run shell commands | Register server in client config |
| Works across agents | ✅ Any agent that reads markdown | ✅ Any agent that runs shell | ⚠️ Only MCP-compatible clients |
| Setup complexity | Copy one file | Install binary | Install binary + register transport |
| Portability | Works in any repo | Works in any shell | Tied to MCP client config |

Skill mode works because modern coding agents (Codex, Claude Code, Cursor, Aider, OpenCode…) already read project documentation and execute shell commands. A `SKILL.md` file teaches the agent **when** and **how** to call the CLI — no protocol integration needed.

## Features

- **Triple interface** — Skill (zero-config) + CLI (shell) + MCP (protocol-native).
- **Three core operations** — `distill_batch`, `distill_watch`, `search_code`.
- **Hexagonal architecture** — `distill/domain`, `distill/application`, `platform/*`.
- **Dependency injection** via `sarulabs/di`.
- **Config management** with `viper` + `.env`.
- **Provider-specific validation** at config time.
- **Interactive setup UI** (`--config-ui`).
- **Unit, integration, and optional live tests.**

## Requirements

- Go **1.26+**
- Make (recommended)

Prebuilt binaries are also available from GitHub Releases (see [Installation](#installation)).

## Installation

### Option A: Build from source

```bash
make build          # → ./bin/context-distill
make install        # → ~/.local/bin/context-distill
```

### Option B: Prebuilt binary

**Linux / macOS:**

```bash
curl -fsSL https://raw.githubusercontent.com/jcastilloa/context-distill/master/scripts/install.sh | sh
# Specific version:
curl -fsSL https://raw.githubusercontent.com/jcastilloa/context-distill/master/scripts/install.sh | VERSION=vX.Y.Z sh
```

**Windows (PowerShell):**

```powershell
iwr https://raw.githubusercontent.com/jcastilloa/context-distill/master/scripts/install.ps1 -UseBasicParsing | iex
# Specific version:
$env:VERSION='vX.Y.Z'; iwr https://raw.githubusercontent.com/jcastilloa/context-distill/master/scripts/install.ps1 -UseBasicParsing | iex
```

**Installer environment variables:**

| Variable | Default |
|---|---|
| `REPO` | `jcastilloa/context-distill` |
| `SERVICE_NAME` | `context-distill` |
| `INSTALL_DIR` | `~/.local/bin` (Linux/macOS) · `%LOCALAPPDATA%\context-distill\bin` (Windows) |
| `VERSION` | Latest release tag |

### Makefile Targets

| Target | Description |
|---|---|
| `make build` | Build binary to `./bin/context-distill` |
| `make install` | Install to `~/.local/bin/context-distill` |
| `make clean` | Remove `./bin` |

## Quick Start

```bash
# 1. Configure provider
context-distill --config-ui

# 2. Verify
echo "PASS: TestA, PASS: TestB, FAIL: TestC - expected 4 got 5" \
  | context-distill distill_batch --question "Did tests pass? PASS or FAIL. If FAIL, list failing test names."

context-distill distill_watch \
  --question "What changed? One short sentence." \
  --previous-cycle "services: api=OK, db=OK, cache=OK" \
  --current-cycle "services: api=OK, db=FAIL, cache=OK"

context-distill search_code --query "provider_name" --mode text \
  --question "Return only file:line, one per line."
```

If commands return expected compact answers, setup is ready. Then choose your mode:

- **Skill** ⭐ — Copy `SKILL.md` into the agent's skills directory (see [Skill Setup](#skill-setup-recommended)).
- **CLI** — Use subcommands directly (see [CLI Reference](#cli-reference)).
- **MCP** — Run `context-distill --transport stdio` and register the server (see [MCP Server Mode](#mcp-server-mode)).

---

## Skill Setup (Recommended)

A skill is a `SKILL.md` file placed inside an agent's skills directory. The agent discovers it, reads the instructions, and learns **when** to activate and **what commands** to run. No protocol config, no server registration.

### Where to install

| Agent | Project-level path | Global path |
|---|---|---|
| **Claude Code** | `.claude/skills/context-distill/SKILL.md` | `~/.claude/skills/context-distill/SKILL.md` |
| **Codex** | `.codex/skills/context-distill/SKILL.md` | `~/.codex/skills/context-distill/SKILL.md` |
| **OpenCode** | `.opencode/skills/context-distill/SKILL.md` | `~/.opencode/skills/context-distill/SKILL.md` |
| **Cursor** | `.cursor/skills/context-distill/SKILL.md` | `~/.cursor/skills/context-distill/SKILL.md` |

**Project-level** (teams): every agent working on the repo picks it up.
**Global** (personal): available in every project without per-repo setup.

```bash
# Single agent, project-level (e.g. Claude Code)
mkdir -p .claude/skills/context-distill
cp SKILL.md .claude/skills/context-distill/SKILL.md

# All agents, global
for agent in .claude .codex .opencode .cursor; do
  mkdir -p ~/"$agent"/skills/context-distill
  cp SKILL.md ~/"$agent"/skills/context-distill/SKILL.md
done
```

### SKILL.md contents

> The full `SKILL.md` is maintained as a standalone file in the repository root. Below is the reference copy.

````markdown
---
name: context-distill
description: >
  Distills verbose command output and retrieves compact code context before sending
  payloads to an LLM. Saves tokens, reduces noise, and keeps context windows clean.
  Use before sending command output longer than 5–8 lines, after tests/builds/linters/git
  commands/docker logs, when comparing watch-mode snapshots, and before opening many files
  to locate symbols/usages/config paths.
---

Distill verbose CLI output and retrieve compact code context before passing to LLM. Keep signal. Drop noise.

## Activation

Use BEFORE sending any command output longer than 5–8 lines to LLM.

Use AFTER:
- tests
- builds
- linters
- git commands
- docker logs
- any verbose CLI tool

Use when:
- comparing two snapshots of same source in watch mode
- unsure whether to distill
- locating symbols/usages/config loading/entrypoints before opening many files

Default rule: **always distill**. Unnecessary distill cost ≈ 0. Flooding context expensive.

## Skip

Do not use when:
- output is ≤ 5–8 lines and already human-readable
- exact raw bytes required (audit / compliance / binary integrity)
- interactive terminal debugging needs character-by-character flow

## Commands

### Distill full output

```bash
# Pipe — preferred
<command> | context-distill distill_batch --question "<question with output contract>"

# Explicit flag
context-distill distill_batch --question "<question with output contract>" --input "<raw output>"

# Explicit stdin marker
<command> | context-distill distill_batch --question "<question with output contract>" --input -
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

## Rules

1. **Every call MUST include an output contract in `--question`.**
   Say exact return format:
   - `PASS or FAIL`
   - `valid JSON {severity, file, message}`
   - `filenames, one per line`

2. **One task per call.**
   Do not mix unrelated questions.

3. **Prefer machine-checkable formats.**
   Use:
   - PASS/FAIL
   - JSON
   - one-item-per-line

## Examples

| Source command    | Question                                                                              |
|-------------------|---------------------------------------------------------------------------------------|
| `go test ./...`   | `"Did all tests pass? PASS or FAIL. If FAIL, list failing test names, one per line."` |
| `git diff`        | `"List only changed file paths, one per line."`                                       |
| CI / build logs   | `"Return JSON array: [{severity, file, message}]."`                                   |
| `docker logs`     | `"Summarise errors only. One bullet per distinct error."`                             |
| `find` / `ls -lR` | `"Return only *.go paths, one per line."`                                             |

### Watch examples

| Question                                               | previous_cycle | current_cycle |
|--------------------------------------------------------|----------------|---------------|
| `"What changed in failure count? One short sentence."` | snapshot T-1   | snapshot T    |
| `"Return only newly failing services, one per line."`  | status at T-1  | status at T   |

### Search examples

| Query type | Example contract |
|---|---|
| `symbol` | `"Return likely definitions first as file:line, one per line."` |
| `text` | `"Return top 10 matches as JSON array [{file,line,snippet}]."` |
| `path` | `"Return matching file paths only, one per line."` |

## Binary location

If installed via `make install`: `~/.local/bin/context-distill`

If binary not in PATH, use absolute path.
````

---

## CLI Reference

### Input methods

```bash
echo "data" | context-distill distill_batch --question "..."   # Pipe (preferred)
context-distill distill_batch --question "..." --input "data"   # Explicit flag
echo "data" | context-distill distill_batch --question "..." --input -  # Explicit stdin marker
```

### `distill_batch`

Distills one raw output payload using an explicit question contract.

| Flag | Required | Description |
|---|---|---|
| `--question` | yes | Exact question to answer from the command output. |
| `--input` | no | Raw command output. If omitted, reads from stdin. |

```bash
go test ./... 2>&1 | context-distill distill_batch --question "Did all tests pass? Return only PASS or FAIL."
```

### `distill_watch`

Returns only the relevant delta between two snapshots.

| Flag | Required | Description |
|---|---|---|
| `--question` | yes | Exact question to answer from cycle changes. |
| `--previous-cycle` | yes | Previous watch cycle output snapshot. |
| `--current-cycle` | yes | Current watch cycle output snapshot. |

```bash
context-distill distill_watch \
  --question "Return only newly failing services, one per line." \
  --previous-cycle "$(cat /tmp/health.prev)" \
  --current-cycle "$(cat /tmp/health.curr)"
```

### `search_code`

Searches repository code locally and distills compact matches.

| Flag | Required | Description |
|---|---|---|
| `--query` | yes | Search query (text, regex, symbol name, or path fragment). |
| `--mode` | yes | Search mode: `text`, `regex`, `symbol`, or `path`. |
| `--question` | yes | Output contract for the result. |
| `--scope` | no | Glob filters (repeat flag or comma-separated). |
| `--max-results` | no | Hard limit for returned candidates (default `20`). |
| `--context-lines` | no | Context lines around each match (default `2`). |

```bash
context-distill search_code \
  --query "distill_watch" --mode symbol \
  --question "Return definitions first, then usages, as file:line."
```

### CLI Notes

- Invalid or missing inputs return a non-zero exit code.
- Output goes to stdout exactly as produced by the use case.

---

## MCP Server Mode

### Running the server

```bash
context-distill --transport stdio
# or without building:
go run ./cmd/server --transport stdio
```

### Server Flags

| Flag | Description | Default |
|---|---|---|
| `--transport` | MCP transport mode (`stdio`) | `service.transport` |
| `--config-ui` | Open setup UI and exit | `false` |

### MCP Tools

The MCP tools expose the same operations as the CLI over `stdio` transport.

#### `distill_batch`

| Parameter | Type | Required | Description |
|---|---|---|---|
| `question` | string | yes | What to extract. Must include an output contract. |
| `input` | string | yes | Raw command output to distill. |

#### `distill_watch`

| Parameter | Type | Required | Description |
|---|---|---|---|
| `question` | string | yes | What delta to report. Must include an output contract. |
| `previous_cycle` | string | yes | Snapshot from the previous cycle. |
| `current_cycle` | string | yes | Snapshot from the current cycle. |

#### `search_code`

| Parameter | Type | Required | Description |
|---|---|---|---|
| `query` | string | yes | Search query (text, regex, symbol, or path). |
| `mode` | string | yes | `text`, `regex`, `symbol`, `path`. |
| `question` | string | yes | Output contract for the response. |
| `scope` | array[string] | no | Glob scope filters. |
| `max_results` | number | no | Hard match limit (default `20`). |
| `context_lines` | number | no | Context lines per match (default `2`). |

### MCP Client Registration

#### JSON-based clients (Claude Desktop, Cursor, etc.)

```json
{
  "mcpServers": {
    "context-distill": {
      "command": "/absolute/path/to/context-distill",
      "args": ["--transport", "stdio"]
    }
  }
}
```

#### Codex — TOML config (recommended)

```toml
# ~/.codex/config.toml
[mcp_servers.context-distill]
command = "/absolute/path/to/context-distill"
args = ["--transport", "stdio"]
startup_timeout_sec = 20.0
```

Or via CLI:

```bash
codex mcp add context-distill -- /absolute/path/to/context-distill --transport stdio
codex mcp list        # verify
```

Restart the Codex session after registration.

#### OpenCode — interactive CLI

```bash
opencode mcp add
```

Follow the prompts: **Location** → Current project or Global · **Name** → `context-distill` · **Type** → `local` · **Command** → `/absolute/path/to/context-distill --transport stdio`.

Or edit `opencode.json` directly:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "context-distill": {
      "type": "local",
      "command": ["/absolute/path/to/context-distill", "--transport", "stdio"],
      "enabled": true
    }
  }
}
```

#### Registration notes

- Always use an **absolute** binary path.
- Always use `stdio` transport.
- If the server does not appear, inspect with `codex mcp list --json` or restart the agent session.

---

## Configuration

### Terminal UI (recommended for first-time setup)

```bash
context-distill --config-ui
```

| Field | Notes |
|---|---|
| `provider_name` | Dropdown of supported providers. |
| `base_url` | Required for OpenAI-compatible providers. |
| `api_key` | Masked input. Required for `openai`, `openrouter`, `jan`. |

**Validation:** providers requiring an API key block save until one is entered. OpenAI-compatible providers require `base_url`. Provider aliases are normalized (e.g. `OpenAI Compatible` → `openai-compatible`, `dmr` → `docker-model-runner`).

**Config path:** `~/.config/context-distill/config.yaml`. Save preserves existing YAML sections and updates only the relevant `distill` fields.

### Manual YAML configuration

**Lookup order:** `~/.config/<service>/config.yaml` → `./config.yaml`

```yaml
service:
  transport: stdio

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

> `service.version` is injected at build time and does not need to be set manually.

### Provider Matrix

| Provider | Transport | API Key | Default Base URL |
|---|---|---|---|
| `ollama` | native ollama | No | `http://127.0.0.1:11434` |
| `openai` | openai-compatible | Yes | `https://api.openai.com/v1` |
| `openrouter` | openai-compatible | Yes | `https://openrouter.ai/api/v1` |
| `openai-compatible` | openai-compatible | No (backend-dependent) | — |
| `lmstudio` | openai-compatible | No | `http://127.0.0.1:1234/v1` |
| `jan` | openai-compatible | Yes | `http://127.0.0.1:1337/v1` |
| `localai` | openai-compatible | No | `http://127.0.0.1:8080/v1` |
| `vllm` | openai-compatible | No | `http://127.0.0.1:8000/v1` |
| `sglang` | openai-compatible | No | — |
| `llama.cpp` | openai-compatible | No | — |
| `mlx-lm` | openai-compatible | No | — |
| `docker-model-runner` | openai-compatible | No | `http://127.0.0.1:12434/engines/v1` |

---

## Writing Good Questions

The quality of output depends on the `question` contract. Be explicit about **what** you want and **in what format**.

| ❌ Bad | ✅ Good |
|---|---|
| `"What happened?"` | `"Did tests pass? Return only PASS or FAIL. If FAIL, list failing test names, one per line."` |
| `"Summarize this"` | `"List only changed file paths, one per line."` |
| | `"Return valid JSON only with keys: severity, file, message."` |

---

## AGENTS.md Templates

If you prefer to embed instructions in your project's `AGENTS.md` instead of using `SKILL.md`, use one of the templates below. All share a common base; only the invocation style and strictness differ.

### Common rules (all modes)

1. **Every call MUST include an output contract in `--question` / `question`** — exact return format (`PASS or FAIL`, `valid JSON {severity, file, message}`, `filenames, one per line`, etc.).
2. **One task per call.** No mixing unrelated questions.
3. **Prefer machine-checkable formats** (PASS/FAIL, JSON, one-item-per-line).

**When to skip (exceptions only):**
- Output is ≤ 5–8 lines and already human-readable.
- Exact raw bytes required (audit / compliance / binary integrity).
- Interactive terminal debugging where character-by-character flow matters.

### Template: CLI Mode

````md
## context-distill CLI Usage

Default behavior: distill command output before sending it to the LLM.
Use `search_code` before opening many files to locate symbols/usages/config paths.

### CLI patterns

```bash
# Batch
go test ./... 2>&1 | context-distill distill_batch \
  --question "Did all tests pass? Return only PASS or FAIL. If FAIL, list failing tests one per line."

# Snapshot delta
context-distill distill_watch \
  --question "Return only newly failing services, one per line." \
  --previous-cycle "$(cat /tmp/status.prev)" \
  --current-cycle "$(cat /tmp/status.curr)"

# Code retrieval
context-distill search_code \
  --query "provider_name" --mode text \
  --question "Return only file:line, one per line."
```
````

### Template: MCP Mode

````md
## context-distill MCP Usage

**Default: ALWAYS distill.** Use `distill_batch` for ANY command output before
sending it to the LLM. Skip ONLY if output ≤ 5–8 lines and readable at a glance.
When unsure: **distill** — unnecessary calls cost ≈ 0; flooding context is expensive.

Use `search_code` before opening many files for symbol/config/entrypoint localization.

### `distill_batch` examples

| Source command    | `question`                                                                          |
|-------------------|-------------------------------------------------------------------------------------|
| `go test ./...`   | "Did all tests pass? PASS or FAIL. If FAIL, list failing test names, one per line." |
| `git diff`        | "List only changed file paths, one per line."                                       |
| CI / build logs   | "Return JSON array: `[{severity, file, message}]`."                                |
| `docker logs`     | "Summarise errors only. One bullet per distinct error."                             |
| `find` / `ls -lR` | "Return only `*.go` paths, one per line."                                           |

### `distill_watch` examples

| `question`                                           | `previous_cycle` | `current_cycle` |
|------------------------------------------------------|------------------|-----------------|
| "What changed in failure count? One short sentence." | snapshot T-1     | snapshot T      |
| "Return only newly failing services, one per line."  | status at T-1    | status at T     |

### `search_code` example

`query="LoadDistillConfig", mode="symbol", question="Return likely definitions first as file:line, one per line."`
````

### Template: Strict CI Mode

````md
## context-distill MCP Usage (CI Mode)

CRITICAL: For any command output consumed by automation, call `distill_batch` first.
CRITICAL: For repository localization in automation steps, call `search_code` first.
CRITICAL: Every `question` must define an explicit, machine-parseable output contract.

### Standard contracts

- **Test status:** `"Return valid JSON only with keys: status, failing_tests. status must be PASS or FAIL."`
- **Lint status:** `"Return valid JSON only with keys: status, issues. issues must be an array of {file, line, message}."`
- **Diff summary:** `"Return valid JSON only with key files_changed as an array of file paths."`

### `distill_watch` in CI

Use only for periodic snapshots with strict delta output.
Example: `"Return valid JSON only with keys: changed, added, removed. Each must be an array of strings."`

### `search_code` in CI

Example: `context-distill search_code --query "provider_name" --mode text --question "Return valid JSON only as array [{file,line,snippet}]."`

### Failure handling

- If output does not match the requested schema, treat as invalid and re-run with a stricter question.
- If exact raw output is needed for audit/compliance, bypass distillation.
````

### Policy one-liner

```md
Default policy: use `context-distill` (via Skill, CLI, or MCP) to distill command output and run `search_code` for repository localization before sending data to an LLM, unless raw output is explicitly required.
```

---

## Project Structure

```text
context-distill/
├── cmd/server/              # main.go, bootstrap.go, openai_distill_config.go
├── distill/
│   ├── application/distillation/
│   └── domain/
├── mcp/
│   ├── application/
│   └── domain/
├── platform/
│   ├── config/
│   ├── configui/
│   ├── di/
│   ├── mcp/                 # commands/, server/, tools/
│   ├── ollama/
│   └── openai/
├── shared/
│   ├── ai/domain/
│   └── config/domain/
├── SKILL.md
├── config.sample.yaml
├── config.yaml
├── Makefile
└── AGENTS.md
```

## Architecture

```text
platform  →  shared + distill/application + distill/domain
distill/application  →  distill/domain
cmd  →  platform + shared
```

**Constraint:** `shared` and `distill/domain` must never import `platform`.

## Development

```bash
go test ./...       # unit + integration tests
go vet ./...        # static checks
```

### Suggested local workflow

1. `go test ./...`
2. `context-distill --config-ui`
3. Verify CLI:
   ```bash
   echo "hello world" | context-distill distill_batch --question "Return the input verbatim."
   context-distill search_code --query "provider_name" --mode text --question "Return only file:line, one per line."
   ```
4. (Optional) Start MCP server: `context-distill --transport stdio`
5. Validate from your MCP client or via CLI.

### Optional live test

```bash
DISTILL_LIVE_TEST=1 OPENAI_BASE_URL=https://openrouter.ai/api/v1 \
go test -tags=live ./platform/di -run TestLiveDistillBatchWithOpenAICompatibleProvider -v
```

## Troubleshooting

| Problem | Fix |
|---|---|
| `provider unauthorized` | Verify `distill.api_key` (or fallback `openai.api_key`). |
| `requires base_url` | Set `distill.base_url` — fastest via `--config-ui`. |
| MCP client does not detect server | Confirm absolute binary path, execute permissions, `stdio` transport. |
| Config validation fails on start | Run `--config-ui` first. |
| CLI returns non-zero exit code | Check required flags: `distill_batch` → `--question`; `distill_watch` → `--question`, `--previous-cycle`, `--current-cycle`; `search_code` → `--query`, `--mode`, `--question`. |
| Agent ignores `SKILL.md` | Ensure correct path (see [Skill Setup](#skill-setup-recommended)). |

## Security

- **Never** commit real API keys to public repositories.
- Prefer environment-based secrets in shared or CI environments.

## License

Copyright © 2026 jcastilloa. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice,
   this list of conditions and the following disclaimer.
2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.
3. Neither the name of the copyright holder nor the names of its contributors
   may be used to endorse or promote products derived from this software
   without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.
