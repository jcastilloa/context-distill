# context-distill

A Go tool that **distills command output** before it reaches a paid LLM. Available as a **Skill** (recommended), a **standalone CLI**, and an **MCP server**. Inspired by the `distill` CLI and built with hexagonal architecture, dependency injection, and TDD.

## Overview

`context-distill` exposes two distillation operations accessible in **three ways**:

| Mode | Best for | How it works |
|---|---|---|
| **Skill** ⭐ (recommended) | Any agent that can read markdown and run shell commands | The agent reads a `SKILL.md` file and learns when and how to invoke the CLI. Zero config on the agent side. |
| **CLI** | Local scripts, CI pipelines, shell-capable agents | Direct subcommands: `context-distill distill_batch`, `context-distill distill_watch`. |
| **MCP** | Agents/clients with native MCP support (Claude Desktop, Cursor, Codex…) | Runs as an MCP server over `stdio` transport. |

| Operation | Purpose |
|---|---|
| `distill_batch` | Compresses full command output to answer a single, explicit question. |
| `distill_watch` | Compares two consecutive snapshots and returns only the relevant delta. |

All three modes share the same underlying use cases, validation rules, and output behavior — only the invocation method differs.

It also provides:

- LLM provider configuration via YAML and environment variables.
- An interactive terminal UI for first-time setup (`--config-ui`).
- Support for Ollama and any OpenAI-compatible provider.

## Why Skill Mode Is Recommended

| | Skill | CLI | MCP |
|---|---|---|---|
| Agent config required | **None** — just drop `SKILL.md` | Agent must know how to run shell commands | Register server in client config |
| Works across agents | ✅ Any agent that reads markdown | ✅ Any agent that runs shell | ⚠️ Only MCP-compatible clients |
| Setup complexity | Copy one file | Install binary | Install binary + register transport |
| Portability | Works in any repo | Works in any shell | Tied to MCP client config |

Skill mode works because modern coding agents (Codex, Claude Code, Cursor, Aider, OpenCode…) already know how to read project documentation and execute shell commands. A `SKILL.md` file teaches the agent **when** to distill and **how** to call the CLI — no protocol integration needed.

## Features

- **Triple interface** — Skill file for zero-config agent adoption + CLI for direct shell use + MCP tools for protocol-native clients.
- **Hexagonal architecture** — `distill/domain`, `distill/application`, `platform/*`.
- **Dependency injection** via `sarulabs/di`.
- **Config management** with `viper` + `.env`.
- **Provider-specific validation** at config time.
- **Interactive setup UI** (`--config-ui`).
- **Unit, integration, and optional live tests.**

## Requirements

- Go **1.26+**
- Make (recommended)

If you prefer not to compile, you can install a prebuilt binary from GitHub Releases (see below).

## Installation

### Option A: Build from source

```bash
make build
```

The binary is placed at `./bin/context-distill`.

To install it into your PATH:

```bash
make install
# installs to ~/.local/bin/context-distill
```

### Option B: Prebuilt binary (no build required)

**Linux / macOS:**

```bash
# Latest release
curl -fsSL https://raw.githubusercontent.com/jcastilloa/context-distill/master/scripts/install.sh | sh

# Specific version
curl -fsSL https://raw.githubusercontent.com/jcastilloa/context-distill/master/scripts/install.sh | VERSION=vX.Y.Z sh
```

**Windows (PowerShell):**

```powershell
# Latest release
iwr https://raw.githubusercontent.com/jcastilloa/context-distill/master/scripts/install.ps1 -UseBasicParsing | iex

# Specific version
$env:VERSION='vX.Y.Z'; iwr https://raw.githubusercontent.com/jcastilloa/context-distill/master/scripts/install.ps1 -UseBasicParsing | iex
```

**Installer environment variables:**

| Variable | Default |
|---|---|
| `REPO` | `jcastilloa/context-distill` |
| `SERVICE_NAME` | `context-distill` |
| `INSTALL_DIR` | `~/.local/bin` (Linux/macOS) · `%LOCALAPPDATA%\context-distill\bin` (Windows) |
| `VERSION` | Latest release tag |

## Makefile Targets

```bash
make help
```

| Target | Description |
|---|---|
| `make build` | Build the binary to `./bin/context-distill` |
| `make install` | Install binary to `~/.local/bin/context-distill` |
| `make clean` | Remove `./bin` |

## Quick Start

### 1. Configure the provider (all modes)

```bash
context-distill --config-ui
```

### 2. Verify it works

```bash
echo "PASS: TestA, PASS: TestB, FAIL: TestC - expected 4 got 5" | context-distill distill_batch --question "Did tests pass? Return only PASS or FAIL. If FAIL, list failing test names."
```

```bash
context-distill distill_watch --question "What changed? Return one short sentence." --previous-cycle "services: api=OK, db=OK, cache=OK" --current-cycle "services: api=OK, db=FAIL, cache=OK"
```

If both return a distilled answer, you are ready.

### 3. Choose your mode

#### ⭐ Skill mode (recommended)

Copy `SKILL.md` (see [Skill Setup](#skill-setup-recommended)) into your project root or `~/.skills/`. Your agent will read it automatically and start distilling.

#### CLI mode

Use the subcommands directly in scripts or agent shell calls:

```bash
# Pipe (preferred)
echo "data" | context-distill distill_batch --question "..."

# Explicit flag
context-distill distill_batch --question "..." --input "data"

# Explicit stdin marker
echo "data" | context-distill distill_batch --question "..." --input -
```

#### MCP mode

```bash
context-distill --transport stdio
```

Then register the server in your MCP client (see [MCP Client Registration](#mcp-client-registration)).

---

## Skill Setup (Recommended)

### What is a Skill?

A skill is a `SKILL.md` file that an agent reads on demand to learn a new capability. The agent discovers the file, reads the instructions, and knows **when** to activate and **what commands** to run. No protocol config, no server registration.

### Install the Skill

**Option A — Per-project (recommended for teams):**

Copy the `SKILL.md` template below into your project root. Every agent working on this repo will pick it up.

**Option B — Global (recommended for personal use):**

```bash
mkdir -p ~/.skills
# Copy the SKILL.md content into ~/.skills/context-distill.skill.md
```

### SKILL.md

Create this file in your project root (or `~/.skills/context-distill.skill.md` for global use):

````markdown
# Skill: context-distill

## Description

Distills verbose command output (logs, test results, diffs, status checks) into
short, structured answers before sending them to an LLM. Saves tokens, reduces
noise, and keeps context windows clean.

## When to activate

- BEFORE sending any command output longer than 5–8 lines to the LLM.
- AFTER running tests, builds, linters, git commands, docker logs, or any
  verbose CLI tool.
- When comparing two snapshots of the same source (watch mode).
- When unsure whether to distill: **always distill** — unnecessary calls
  cost ≈ 0; flooding context is expensive.

## When to skip

- Output is ≤ 5–8 lines and already human-readable.
- Exact raw bytes are required (audit / compliance / binary integrity).
- Interactive terminal debugging where character-by-character flow matters.

## Commands

### Distill full output (batch)

```bash
# Pipe — preferred
<command> | context-distill distill_batch --question "<question with output contract>"

# Explicit flag
context-distill distill_batch --question "<question with output contract>" --input "<raw output>"

# Explicit stdin marker
<command> | context-distill distill_batch --question "<question with output contract>" --input -
```

### Distill delta between two snapshots (watch)

```bash
context-distill distill_watch \
  --question "<question with output contract>" \
  --previous-cycle "<snapshot T-1>" \
  --current-cycle "<snapshot T>"
```

## Rules

1. **Every call MUST include an output contract in `--question`** — tell the
   distiller the exact return format: `"PASS or FAIL"`, `"valid JSON {severity, file, message}"`,
   `"filenames, one per line"`, etc.
2. **One task per call.** Never mix unrelated questions.
3. **Prefer machine-checkable formats** (PASS/FAIL, JSON, one-item-per-line).

## Examples

| Source command      | Question                                                                              |
|---------------------|---------------------------------------------------------------------------------------|
| `go test ./...`     | `"Did all tests pass? PASS or FAIL. If FAIL, list failing test names, one per line."` |
| `git diff`          | `"List only changed file paths, one per line."`                                       |
| CI / build logs     | `"Return JSON array: [{severity, file, message}]."`                                  |
| `docker logs`       | `"Summarise errors only. One bullet per distinct error."`                             |
| `find` / `ls -lR`   | `"Return only *.go paths, one per line."`                                             |

### Watch examples

| Question                                               | previous_cycle | current_cycle |
|--------------------------------------------------------|----------------|---------------|
| `"What changed in failure count? One short sentence."` | snapshot T-1   | snapshot T    |
| `"Return only newly failing services, one per line."`  | status at T-1  | status at T   |

## Binary location

If installed via `make install`: `~/.local/bin/context-distill`

If the binary is not in PATH, use the absolute path.
````

---

## CLI Commands Reference

The CLI commands provide the same distillation capabilities as the MCP tools but are invoked directly from the shell. Use them in local scripts, CI pipelines, or with agent runtimes that execute shell commands instead of MCP tools.

### Input methods

```bash
# 1. Pipe (preferred)
echo "data" | context-distill distill_batch --question "..."

# 2. Explicit flag
context-distill distill_batch --question "..." --input "data"

# 3. Explicit stdin marker
echo "data" | context-distill distill_batch --question "..." --input -
```

### `distill_batch`

Distills one raw output payload using an explicit question contract.

```bash
go test ./... 2>&1 | context-distill distill_batch --question "Did all tests pass? Return only PASS or FAIL."
```

Flags:

| Flag | Required | Description |
|---|---|---|
| `--question` | yes | Exact question to answer from the command output. |
| `--input` | no | Raw command output to distill. If omitted, reads from stdin. |

### `distill_watch`

Distills only the relevant delta between two snapshots.

```bash
context-distill distill_watch \
  --question "Return only newly failing services, one per line." \
  --previous-cycle "$(cat /tmp/health.prev)" \
  --current-cycle "$(cat /tmp/health.curr)"
```

Flags:

| Flag | Required | Description |
|---|---|---|
| `--question` | yes | Exact question to answer from cycle changes. |
| `--previous-cycle` | yes | Previous watch cycle output snapshot. |
| `--current-cycle` | yes | Current watch cycle output snapshot. |

### CLI Notes

- CLI commands and MCP tools share the same underlying use cases and validation rules.
- Invalid/missing inputs return a non-zero exit code.
- Output is written to standard output exactly as produced by the distill use case.

---

## MCP Server Mode

### Running the server

```bash
context-distill --transport stdio
# or without building:
go run ./cmd/server --transport stdio
```

### Version

```bash
go run ./cmd/server version
```

### Server Flags

| Flag | Description | Default |
|---|---|---|
| `--transport` | MCP transport mode (`stdio`) | `service.transport` |
| `--config-ui` | Open setup UI and exit | `false` |

### MCP Client Registration

After building or installing the binary, register it in your MCP client to use the tool interface.

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

#### Codex — manual TOML (recommended)

Add to `~/.codex/config.toml`:

```toml
[mcp_servers.context-distill]
command = "/absolute/path/to/context-distill"
args = ["--transport", "stdio"]
startup_timeout_sec = 20.0
```

If you used `make install`:

```toml
[mcp_servers.context-distill]
command = "/home/<your-user>/.local/bin/context-distill"
args = ["--transport", "stdio"]
startup_timeout_sec = 20.0
```

#### Codex — CLI registration

```bash
codex mcp add context-distill -- /absolute/path/to/context-distill --transport stdio
```

Verify:

```bash
codex mcp list
codex mcp get context-distill
```

Restart your Codex session so it picks up the new server.

#### OpenCode — interactive CLI

```bash
opencode mcp add
```

Follow the prompts:

1. **Location** → `Current project` or `Global`.
2. **Name** → `context-distill`.
3. **Type** → `local`.
4. **Command** → `/absolute/path/to/context-distill --transport stdio`.

Verify:

```bash
opencode mcp list
```

If the server is not connected yet, restart your OpenCode session.

#### OpenCode — manual config (`opencode.json`)

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
- If the server does not appear, run `codex mcp list --json` to inspect the resolved config.

### MCP Tools Reference

The MCP tools expose the same distillation capabilities as the CLI commands, but are consumed by MCP-compatible clients over the `stdio` transport.

#### `distill_batch`

| Parameter | Type | Required | Description |
|---|---|---|---|
| `question` | string | yes | What to extract from the input. Must include an output contract. |
| `input` | string | yes | Raw command output to distill. |

Returns a short, focused answer to `question`.

#### `distill_watch`

| Parameter | Type | Required | Description |
|---|---|---|---|
| `question` | string | yes | What delta to report. Must include an output contract. |
| `previous_cycle` | string | yes | Snapshot from the previous cycle. |
| `current_cycle` | string | yes | Snapshot from the current cycle. |

Returns a short summary of relevant changes, or a no-change message when nothing meaningful differs.

---

## Configuration

### Terminal UI (recommended for first-time setup)

```bash
context-distill --config-ui
# or without building:
go run ./cmd/server --config-ui
```

**Editable fields:**

| Field | Notes |
|---|---|
| `provider_name` | Dropdown list of supported providers. |
| `base_url` | Required for OpenAI-compatible providers. |
| `api_key` | Masked input. Required for `openai`, `openrouter`, `jan`. |

**Validation rules:**

- Providers that require an API key block save until one is entered.
- OpenAI-compatible providers require a `base_url`.
- Provider aliases are normalized automatically (e.g. `OpenAI Compatible` → `openai-compatible`, `dmr` → `docker-model-runner`).

**Persisted config path:** `~/.config/context-distill/config.yaml`

Save preserves existing YAML sections (`service`, `openai`, etc.) and updates only the relevant `distill` fields.

### Manual YAML configuration

You can also edit the config file directly.

**Lookup order:**

1. `~/.config/<service>/config.yaml`
2. `./config.yaml`

**Example `config.yaml`:**

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

> **Note:** `service.version` is injected at build time from binary metadata and does not need to be set manually.

## Provider Matrix

| Provider | Transport | API Key Required | Default Base URL |
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

## Writing Good Questions

The quality of the distillation depends entirely on the `question` — whether invoked via Skill, CLI, or MCP. Be explicit about **what** you want and **in what format**.

### Bad questions

- `"What happened?"`
- `"Summarize this"`

### Good questions

- `"Did tests pass? Return only PASS or FAIL. If FAIL, list failing test names, one per line."`
- `"List only changed file paths, one per line."`
- `"Return valid JSON only with keys: severity, file, message."`

### `distill_batch` examples

| Source | Question |
|---|---|
| `go test ./...` | `"Did tests pass? Return only PASS or FAIL."` |
| `git diff` | `"List changed files and one short reason per file. Max 10 lines."` |
| CI logs | `"Return only blocking errors with file and line if available."` |

### `distill_watch` examples

| Snapshots | Question |
|---|---|
| Test watcher output at T−1 / T | `"What changed in failure count? Return one short sentence."` |
| Deployment status at T−1 / T | `"Return only newly failing services, one per line."` |

---

## AGENTS.md Templates

If you prefer to embed instructions directly in your project's `AGENTS.md` instead of using a standalone `SKILL.md`, use one of the templates below.

### AGENTS.md Template (CLI Mode)

````md
## context-distill CLI Usage

Default behavior: distill command output before sending it to the LLM.

Use the local binary:
- `context-distill distill_batch`
- `context-distill distill_watch`

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

#### Snapshot delta

```bash
context-distill distill_watch \
  --question "Return only newly failing services, one per line." \
  --previous-cycle "$(cat /tmp/status.prev)" \
  --current-cycle "$(cat /tmp/status.curr)"
```

### When to skip distill (exceptions only)

- Output is ≤ 5–8 lines and readable at a glance.
- Exact raw bytes are required (audit/compliance/binary integrity).
- Interactive terminal debugging where exact character flow matters.
````

### AGENTS.md Template (MCP Mode)

````md
## context-distill MCP Usage

**Default: ALWAYS distill.** Use `distill_batch` for ANY command output before
sending it to the LLM. Skip ONLY if the output is ≤ 5–8 lines and readable at
a glance. When unsure: **distill** — unnecessary calls cost ≈ 0; flooding
context is expensive.

### Rules

1. **Every call MUST include an output contract in `question`** — tell the
   distiller the exact return format: `"PASS or FAIL"`, `"valid JSON {severity, file, message}"`,
   `"filenames, one per line"`, etc.
2. **One task per call.** No mixing unrelated questions.
3. **Prefer machine-checkable formats** (PASS/FAIL, JSON, one-item-per-line).

### `distill_batch` examples

| Source command    | `question`                                                                          |
|-------------------|-------------------------------------------------------------------------------------|
| `go test ./...`   | "Did all tests pass? PASS or FAIL. If FAIL, list failing test names, one per line." |
| `git diff`        | "List only changed file paths, one per line."                                       |
| CI / build logs   | "Return JSON array: `[{severity, file, message}]`."                                |
| `docker logs`     | "Summarise errors only. One bullet per distinct error."                             |
| `find` / `ls -lR` | "Return only `*.go` paths, one per line."                                           |

### `distill_watch` — diff between snapshots

Use when you have two snapshots of the same source to extract only what changed.

| `question`                                           | `previous_cycle` | `current_cycle` |
|------------------------------------------------------|------------------|-----------------|
| "What changed in failure count? One short sentence." | snapshot T-1     | snapshot T      |
| "Return only newly failing services, one per line."  | status at T-1    | status at T     |

### When to skip distill (exceptions only)

- Output **≤ 5–8 lines**, already human-readable.
- You need **exact raw bytes** (compliance / audit / binary integrity).
- Debugging an **interactive terminal** where character-by-character flow matters.
````

### AGENTS.md Template (Strict CI Mode)

````md
## context-distill MCP Usage (CI Mode)

CRITICAL: For any command output consumed by automation, call `distill_batch` first.

CRITICAL: Every `question` must define an explicit output contract and MUST be machine-parseable.
- Prefer JSON objects or arrays only.
- No markdown.
- No prose outside the requested schema.

CRITICAL: If JSON is requested, enforce:
- "Return valid JSON only."
- Fixed keys and fixed value shapes.

### Standard contracts

- **Test status:**
  `"Return valid JSON only with keys: status, failing_tests. status must be PASS or FAIL."`
- **Lint status:**
  `"Return valid JSON only with keys: status, issues. issues must be an array of {file, line, message}."`
- **Diff summary:**
  `"Return valid JSON only with key files_changed as an array of file paths."`

### `distill_watch` in CI

Use `distill_watch` only for periodic snapshots with strict delta output.

Example question:
`"Return valid JSON only with keys: changed, added, removed. Each must be an array of strings."`

### Failure handling

- If the distillation output does not match the requested schema, treat it as invalid and re-run with a stricter question.
- If exact raw output is needed for audit or compliance, bypass distillation.
````

### Suggested policy one-liner

Drop this into your project docs for a quick reference:

```md
Default policy: distill command output with `context-distill` (via Skill, CLI, or MCP) before sending logs/traces/diffs to an LLM, unless raw output is explicitly required.
```

---

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
├── SKILL.md
├── config.sample.yaml
├── config.yaml
├── Makefile
└── AGENTS.md
```

## Architecture

**Dependency rule:**

```text
platform  →  shared + distill/application + distill/domain
distill/application  →  distill/domain
cmd  →  platform + shared
```

**Constraint:** `shared` and `distill/domain` must never import `platform`.

## Development

### Tests and static checks

```bash
go test ./...
go vet ./...
```

### Suggested local workflow

1. `go test ./...`
2. `context-distill --config-ui`
3. Verify CLI works:
   ```bash
   echo "hello world" | context-distill distill_batch --question "Return the input verbatim."
   ```
4. (Optional) Start MCP server: `context-distill --transport stdio`
5. Validate behavior from your MCP client or via CLI commands.

### Optional live test (real provider)

```bash
DISTILL_LIVE_TEST=1 OPENAI_BASE_URL=https://openrouter.ai/api/v1 \
go test -tags=live ./platform/di -run TestLiveDistillBatchWithOpenAICompatibleProvider -v
```

## Troubleshooting

| Problem | Fix |
|---|---|
| `provider unauthorized` | Verify `distill.api_key` (or the fallback `openai.api_key`, depending on the provider). |
| `requires base_url` | Set `distill.base_url`. The fastest path is `--config-ui`. |
| MCP client does not detect the server | Confirm the binary path is absolute, has execute permissions, and transport is `stdio`. |
| Server fails on config validation | Run `--config-ui` for initial setup, then start normally. |
| CLI command returns non-zero exit code | Check that all required flags (`--question`, `--input` or `--previous-cycle`/`--current-cycle`) are provided and non-empty. |
| Agent ignores `SKILL.md` | Ensure the file is in the project root or in `~/.skills/`. Some agents require explicit instructions to read skill files. |

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
