# Deterministic Workflow Hooks for Game Commands

## Problem Statement

The game workflows (Ulysses Protocol, Feature Implementation Game, etc.) have explicit stage gates but rely on agent inference to determine if criteria are met. This creates non-determinism where we want determinism.

## Solution: Enforced Stage Gates via Hooks

Use Claude Code's hook system to **programmatically validate** gate criteria before allowing workflow progression.

## Architecture

```
Game Workflow State (.game-state/)
  ↓
Pre/Post Tool Hooks
  ↓
Gate Validation Scripts
  ↓
PASS → Allow progression
FAIL → Block + Provide actionable feedback
```

## Hook Configuration

Add to your Claude settings (`~/.config/claude-code/settings.json` or project-local):

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": ".claude/commands/games/hooks/validate-gate.sh"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": ".claude/commands/games/hooks/record-action.sh"
          }
        ]
      }
    ],
    "UserPromptSubmit": [
      {
        "matcher": ".*",
        "hooks": [
          {
            "type": "command",
            "command": ".claude/commands/games/hooks/inject-game-state.sh"
          }
        ]
      }
    ]
  }
}
```

## Hook Types

### 1. Gate Validation Hooks (PreToolUse)

**Purpose**: Block actions that violate current stage gate requirements

**Example**: In Ulysses Protocol Phase 2, block implementation tools until planning gate passes

```bash
# .claude/commands/games/hooks/validate-gate.sh
#!/bin/bash
# Reads tool call from stdin, checks current game state, blocks if gate not passed

STATE_FILE=".ulysses-protocol/state.json"

if [ ! -f "$STATE_FILE" ]; then
  # No active game, allow
  exit 0
fi

# Read tool call from stdin
TOOL_INPUT=$(cat)
TOOL_NAME=$(echo "$TOOL_INPUT" | jq -r '.toolName')
CURRENT_PHASE=$(jq -r '.phase' "$STATE_FILE")

# Enforce phase gates
case "$CURRENT_PHASE" in
  1) # Reconnaissance phase
     if [[ "$TOOL_NAME" == "Write" ]] || [[ "$TOOL_NAME" == "Edit" ]]; then
       GATE_PASSED=$(jq -r '.gates.reconnaissance.passed' "$STATE_FILE")
       if [ "$GATE_PASSED" != "true" ]; then
         # Block and provide feedback
         cat << EOF
{
  "block": true,
  "message": "🚫 Gate Violation: Cannot proceed to implementation. Reconnaissance gate not passed.\n\nMissing requirements:\n$(jq -r '.gates.reconnaissance.missing[]' "$STATE_FILE")\n\nRun: /ulysses-gate-check to see status"
}
EOF
         exit 1
       fi
     fi
     ;;
esac

exit 0  # Allow
```

### 2. Action Recording Hooks (PostToolUse)

**Purpose**: Track actions and automatically update game state

```bash
# .claude/commands/games/hooks/record-action.sh
#!/bin/bash
# Records completed actions and updates gate criteria

TOOL_RESULT=$(cat)
TOOL_NAME=$(echo "$TOOL_RESULT" | jq -r '.toolName')

STATE_FILE=".ulysses-protocol/state.json"
[ ! -f "$STATE_FILE" ] && exit 0

# Update gate criteria based on actions
if [[ "$TOOL_NAME" == "Grep" ]] || [[ "$TOOL_NAME" == "Read" ]]; then
  # Mark reconnaissance actions
  jq '.gates.reconnaissance.actions_completed += 1' "$STATE_FILE" > tmp && mv tmp "$STATE_FILE"

  # Auto-check if criteria met
  .claude/commands/games/hooks/auto-gate-check.sh reconnaissance
fi

exit 0
```

### 3. State Injection Hooks (UserPromptSubmit)

**Purpose**: Inject current game state into conversation context

```bash
# .claude/commands/games/hooks/inject-game-state.sh
#!/bin/bash
# Adds game state context to user prompts

STATE_FILE=".ulysses-protocol/state.json"

if [ -f "$STATE_FILE" ]; then
  CURRENT_PHASE=$(jq -r '.phase' "$STATE_FILE")
  GATE_STATUS=$(jq -r '.gates' "$STATE_FILE" | jq -c)

  cat << EOF
{
  "append": "\n\n---\n**Active Game State:**\n- Protocol: Ulysses\n- Phase: $CURRENT_PHASE\n- Gates: $GATE_STATUS\n---"
}
EOF
fi

exit 0
```

## Gate Definition Format

Each game maintains a `.game-state/gates.json` defining **deterministic criteria**:

```json
{
  "protocol": "ulysses",
  "phase": 1,
  "gates": {
    "reconnaissance": {
      "criteria": [
        {
          "type": "file_exists",
          "path": ".ulysses-protocol/problem-statement.md",
          "description": "Problem statement documented"
        },
        {
          "type": "file_contains",
          "path": ".ulysses-protocol/root-cause.md",
          "pattern": "Root Cause:|Hypothesis:",
          "description": "Root cause analysis documented"
        },
        {
          "type": "command_success",
          "command": "grep -r 'TODO\\|FIXME' . | wc -l",
          "validator": "[ $OUTPUT -lt 50 ]",
          "description": "Technical debt catalogued"
        },
        {
          "type": "manual_approval",
          "approver_role": "tech_lead",
          "description": "Technical lead review required"
        }
      ],
      "passed": false,
      "missing": ["problem-statement.md", "root-cause.md", "tech_lead_approval"]
    },
    "planning": {
      "criteria": [
        {
          "type": "file_exists",
          "path": ".ulysses-protocol/approaches.md",
          "description": "3 solution approaches documented"
        },
        {
          "type": "json_valid",
          "path": ".ulysses-protocol/decision-matrix.json",
          "schema": "decision_matrix_schema.json",
          "description": "Decision matrix complete"
        }
      ],
      "passed": false
    }
  }
}
```

## Criterion Types

### 1. File-Based Criteria (Deterministic)

```json
{
  "type": "file_exists",
  "path": ".game-state/artifact.md"
}

{
  "type": "file_contains",
  "path": ".game-state/plan.md",
  "pattern": "Success Criteria:",
  "min_occurrences": 3
}

{
  "type": "file_line_count",
  "path": ".game-state/test-results.txt",
  "min_lines": 10
}
```

### 2. Command-Based Criteria (Deterministic)

```json
{
  "type": "command_success",
  "command": "make test",
  "description": "All tests pass"
}

{
  "type": "command_output",
  "command": "git diff --name-only",
  "validator": "[ $(echo $OUTPUT | wc -w) -eq 0 ]",
  "description": "No uncommitted changes"
}

{
  "type": "metric_threshold",
  "command": "go test -bench . | grep 'ns/op'",
  "threshold": "< 1000000",
  "description": "Performance within bounds"
}
```

### 3. State-Based Criteria (Deterministic)

```json
{
  "type": "iteration_count",
  "max": 3,
  "description": "Iteration limit not exceeded"
}

{
  "type": "time_elapsed",
  "budget_seconds": 7200,
  "description": "Within time budget"
}

{
  "type": "dependency_met",
  "depends_on": ["reconnaissance.passed"],
  "description": "Previous gate passed"
}
```

### 4. Manual Approval (Semi-Deterministic)

```json
{
  "type": "manual_approval",
  "approver_role": "tech_lead",
  "approval_file": ".game-state/approvals/tech-lead.sig",
  "description": "Tech lead sign-off required"
}
```

**Approval mechanism:**

```bash
# Human approver runs:
echo "APPROVED_BY: jane.doe@company.com $(date -Iseconds)" > .game-state/approvals/tech-lead.sig

# Hook validates signature exists
```

## Utility Scripts

### Gate Checker

```bash
#!/bin/bash
# .claude/commands/games/hooks/check-gate.sh GATE_NAME

GATE_NAME="$1"
STATE_FILE=".ulysses-protocol/state.json"

# Load criteria
CRITERIA=$(jq -r ".gates.$GATE_NAME.criteria[]" "$STATE_FILE")

ALL_PASSED=true

echo "🔍 Checking gate: $GATE_NAME"
echo ""

# Check each criterion
while IFS= read -r criterion; do
  TYPE=$(echo "$criterion" | jq -r '.type')
  DESC=$(echo "$criterion" | jq -r '.description')

  case "$TYPE" in
    file_exists)
      PATH=$(echo "$criterion" | jq -r '.path')
      if [ -f "$PATH" ]; then
        echo "✅ $DESC"
      else
        echo "❌ $DESC (missing: $PATH)"
        ALL_PASSED=false
      fi
      ;;
    command_success)
      CMD=$(echo "$criterion" | jq -r '.command')
      if eval "$CMD" &>/dev/null; then
        echo "✅ $DESC"
      else
        echo "❌ $DESC (command failed)"
        ALL_PASSED=false
      fi
      ;;
  esac
done <<< "$CRITERIA"

if [ "$ALL_PASSED" = true ]; then
  echo ""
  echo "🎉 Gate passed! Updating state..."
  jq ".gates.$GATE_NAME.passed = true | .gates.$GATE_NAME.missing = []" "$STATE_FILE" > tmp && mv tmp "$STATE_FILE"
  exit 0
else
  echo ""
  echo "⚠️  Gate not passed. Complete missing items."
  exit 1
fi
```

### Gate Status Command

```bash
#!/bin/bash
# /ulysses-gate-status - Show current gate status

STATE_FILE=".ulysses-protocol/state.json"
[ ! -f "$STATE_FILE" ] && echo "No active Ulysses Protocol session" && exit 0

CURRENT_PHASE=$(jq -r '.phase' "$STATE_FILE")
PHASE_NAME=$(jq -r ".phase_names[$CURRENT_PHASE]" "$STATE_FILE")

echo "📊 Ulysses Protocol Status"
echo "=========================="
echo "Phase: $CURRENT_PHASE - $PHASE_NAME"
echo ""

# Show all gates
jq -r '.gates | to_entries[] | "Gate: \(.key)\nPassed: \(.value.passed)\nMissing: \(.value.missing | join(", "))\n"' "$STATE_FILE"
```

## Integration with Game Workflows

### Example: Ulysses Protocol with Hooks

**Phase 1 (Reconnaissance) enforced gates:**

1. ✅ Problem statement file exists
2. ✅ Root cause documented
3. ✅ System dependencies mapped
4. ✅ Risk assessment complete
5. ⚠️ Manual approval from tech lead

**Agent cannot proceed to Phase 2 until all ✅**

**Hook blocks Write/Edit tools:**

```
🚫 Gate Violation: Cannot proceed to implementation.

Missing requirements:
- Manual approval from tech lead

To approve: /ulysses-approve reconnaissance tech_lead
To check status: /ulysses-gate-status
```

### Example: Feature Implementation Game

**Phase 2 (Implementation) enforced gates:**

1. ✅ Feature flag created
2. ✅ Tests written (min 80% coverage)
3. ✅ Performance benchmarks run
4. ✅ Docs updated
5. ❌ No linting errors

**Hook blocks Git commit:**

```
🚫 Cannot commit: Implementation gate not passed

Failing criteria:
❌ Linting errors present (12 errors)

Run: make lint-fix
Then: /feature-gate-check implementation
```

## Benefits

### 1. **Determinism**

- Gates are programmatically verified
- No agent discretion on "is this good enough?"
- Explicit, traceable criteria

### 2. **Early Failure Detection**

- Blocks bad actions before they happen
- Prevents wasted iterations
- Forces adherence to process

### 3. **Auditability**

- Every gate check is logged
- State transitions recorded
- Approval trail maintained

### 4. **Reduced Cognitive Load**

- Agent doesn't guess if it should proceed
- Clear feedback on what's missing
- Structured problem-solving

## Implementation Checklist

- [ ] Create `.claude/commands/games/hooks/` directory
- [ ] Implement `validate-gate.sh` (PreToolUse hook)
- [ ] Implement `record-action.sh` (PostToolUse hook)
- [ ] Implement `inject-game-state.sh` (UserPromptSubmit hook)
- [ ] Create `check-gate.sh` utility
- [ ] Define gate schemas for each game (ulysses, feature-impl, sandbox-test)
- [ ] Update game workflows to initialize gates
- [ ] Add `/game-gate-status` slash command
- [ ] Add `/game-approve` slash command for manual gates
- [ ] Document gate criteria for each phase
- [ ] Test with actual workflow execution

## Next Steps

Want me to implement:

1. **The core hook scripts** (validate-gate.sh, record-action.sh, etc.)
2. **Gate definition schemas** for Ulysses Protocol, Feature Implementation Game, etc.
3. **Integration into existing game workflows**
4. **Testing harness** to validate the hook system works

Which would you like first?
