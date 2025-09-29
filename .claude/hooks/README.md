# Immediate Feedback Hooks

Provides instant feedback to agents on code quality, test failures, and syntax errors after every file change.

## Philosophy

**Agents should get feedback as fast as humans do in their IDEs** - linting errors, test failures, and syntax issues should be surfaced immediately, not after committing or pushing.

## Hooks Implemented

### 1. Post-Edit Linting (`post-edit-lint.sh`)

**Trigger**: After every `Edit` tool use

**Behavior**:

- Runs appropriate linter based on file type:
  - **Go**: `golangci-lint run --fix`
  - **TypeScript/JS**: `eslint --fix`
  - **Python**: `ruff check --fix`
  - **Shell**: `shellcheck`
- Auto-fixes where possible
- Reports issues immediately in conversation

**Example Output**:

```
⚠️ **Linting Issues Detected in pkg/mcp/server.go**

Run `golangci-lint run --fix pkg/mcp/server.go` to see details.

Common issues:
- Unused imports
- Unused variables
- Missing error checks
- Formatting issues
```

### 2. Post-Edit Testing (`post-edit-test.sh`)

**Trigger**: After every `Edit` tool use on code files

**Behavior**:

- Finds and runs relevant tests:
  - **Go**: `go test ./package`
  - **TypeScript**: `jest file.test.ts`
  - **Python**: `pytest file_test.py`
- Shows last 20 lines of output
- Surfaces failures immediately

**Example Output**:

```
❌ **Tests Failed for pkg/mcp**

Run `go test -v ./pkg/mcp` to see full output.

**Action Required**: Fix failing tests before proceeding.
```

### 3. Post-Write Validation (`post-write-validate.sh`)

**Trigger**: After every `Write` tool use (new file creation)

**Behavior**:

- Validates file syntax:
  - **Go**: Checks if code compiles, runs `gofmt`
  - **TypeScript**: Type-checks with `tsc --noEmit`
  - **Python**: Syntax check with `py_compile`
  - **JSON**: Validates with `jq`
  - **YAML**: Validates with `yamllint`
  - **Shell**: Syntax check, makes executable
- Reports all issues in one feedback block

**Example Output**:

```
📝 **File Validation Results for cmd/new-tool/main.go**

- ❌ Go syntax errors
- ⚠️ Formatted with gofmt
- ✅ Made executable
```

## Configuration

Hooks are configured in `.claude/hooks.json`:

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit",
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/post-edit-lint.sh"
          },
          {
            "type": "command",
            "command": ".claude/hooks/post-edit-test.sh"
          }
        ]
      },
      {
        "matcher": "Write",
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/post-write-validate.sh"
          }
        ]
      }
    ]
  }
}
```

## Benefits

### 1. **Immediate Error Detection**

- Agent knows instantly if code has issues
- No need to wait for CI/CD pipeline
- Prevents cascading errors

### 2. **Faster Iteration**

- Fix issues right away
- Reduce back-and-forth
- Maintain flow state

### 3. **Learning Loop**

- Agent sees patterns in linting errors
- Learns project-specific conventions
- Improves code quality over time

### 4. **Reduced Context Switching**

- Everything in one conversation
- No need to check external tools
- Self-contained feedback

## Supported Languages

| Language | Linter | Test Runner | Syntax Check |
|----------|--------|-------------|--------------|
| Go | golangci-lint | go test | go build |
| TypeScript | eslint | jest | tsc |
| JavaScript | eslint | jest | - |
| Python | ruff | pytest | py_compile |
| Shell | shellcheck | - | bash -n |
| JSON | - | - | jq |
| YAML | yamllint | - | yamllint |

## Hook Execution Flow

```
Agent Edits File
      ↓
Edit Tool Completes
      ↓
Post-Edit Hooks Triggered (parallel)
      ↓
┌─────────────┬─────────────┐
│ Linter Hook │  Test Hook  │
└──────┬──────┴──────┬──────┘
       ↓              ↓
  Run Linter    Run Tests
       ↓              ↓
  Format JSON   Format JSON
       └──────┬───────┘
              ↓
      Inject into Conversation
              ↓
   Agent Sees Feedback Immediately
```

## Installation

Hooks are automatically active when:

1. `.claude/hooks.json` exists (✅ already created)
2. Hook scripts are executable (✅ already set)
3. Claude Code reads project configuration (restart may be needed)

To verify hooks are active:

```bash
# Check hook configuration
cat .claude/hooks.json

# Verify scripts are executable
ls -la .claude/hooks/*.sh

# Test a hook manually
echo '{"arguments":{"file_path":"cmd/simple-mcp/main.go"}}' | .claude/hooks/post-edit-lint.sh
```

## Customization

### Adding More Linters

Edit `.claude/hooks/post-edit-lint.sh`:

```bash
# Add Rust support
if [[ "$FILE_PATH" =~ \.rs$ ]]; then
  if command -v rustfmt &> /dev/null; then
    rustfmt "$FILE_PATH"
  fi
  if command -v clippy &> /dev/null; then
    cargo clippy --fix "$FILE_PATH"
  fi
fi
```

### Adjusting Test Scope

Edit `.claude/hooks/post-edit-test.sh`:

```bash
# Run full test suite instead of package tests
if [[ "$FILE_PATH" =~ \.go$ ]]; then
  go test ./... -short  # Run all tests in short mode
fi
```

### Disabling Specific Hooks

Remove from `.claude/hooks.json`:

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit",
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/post-edit-lint.sh"
          }
          // Remove test hook to disable automatic testing
        ]
      }
    ]
  }
}
```

## Debugging Hooks

If hooks aren't working:

```bash
# Run Claude Code with debug flag
claude --debug

# Check hook execution logs
# Hooks output to stderr, visible in debug mode

# Test hook manually with sample input
echo '{"toolName":"Edit","arguments":{"file_path":"test.go"}}' | .claude/hooks/post-edit-lint.sh
```

## Performance Considerations

- **Linting**: Fast (< 1s for most files)
- **Testing**: Can be slow for large test suites
  - Consider adding file size/complexity gates
  - Use test parallelization
  - Run only affected tests

## Future Enhancements

- [ ] **Git pre-commit integration** - Run hooks before commit
- [ ] **Performance budgets** - Fail if file changes degrade performance
- [ ] **Security scanning** - Check for vulnerabilities on dependency changes
- [ ] **Documentation validation** - Ensure code changes have corresponding docs
- [ ] **Type coverage tracking** - Monitor TypeScript type coverage
- [ ] **Benchmark comparison** - Compare performance before/after changes

## Related

- [Game Workflow Hooks](./games/hooks/README.md) - Stage-gate validation for game protocols
- [MCP SDK Migration Tool](./../commands/mcp-sdk-migrate.sh) - Automated migration utility

---

**Result**: Agents get IDE-like feedback instantly, leading to higher code quality and faster iteration cycles.
