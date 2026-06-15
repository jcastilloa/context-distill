## context-distill MCP Usage (CI Mode)

CRITICAL: For command output consumed by automation, call `distill_batch` first.
CRITICAL: For raw MCP payloads or MCP tool calls in automation, use `distill_mcp_output`.
CRITICAL: For repository localization in automation, call `search_code` first.

Every `question` must define an explicit machine-parseable output contract.
- Prefer JSON objects or arrays only.
- No markdown.
- No prose outside the requested schema.

### Standard contracts

- Test status:
  `"Return valid JSON only with keys: status, failing_tests. status must be PASS or FAIL."`
- Lint status:
  `"Return valid JSON only with keys: status, issues. issues must be an array of {file, line, message}."`
- Diff summary:
  `"Return valid JSON only with key files_changed as an array of file paths."`
- MCP distilled info:
  `"Return valid JSON only with keys: status, version."`

### Examples

```bash
context-distill distill_batch \
  --question "Return valid JSON only with keys: status, failing_tests. status must be PASS or FAIL." \
  --input "<raw test output>"
```

```bash
context-distill distill_mcp_output \
  --server-command "/absolute/path/to/mcp-server" \
  --server-arg "--transport" \
  --server-arg "stdio" \
  --tool-name "get_status" \
  --tool-arguments '{"scope":"production"}' \
  --question "Return valid JSON only with keys: status, version."
```

```bash
context-distill search_code \
  --query "provider_name" \
  --mode text \
  --question "Return valid JSON only as array [{file,line,snippet}]."
```

### Failure handling

- If the output does not match the requested schema, treat it as invalid and re-run with a stricter question.
- If exact raw output is required for audit or compliance, bypass distillation.
