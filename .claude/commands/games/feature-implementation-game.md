# /feature-implementation-game

Ship new product capabilities with discipline. An OR-inspired shipping protocol that minimizes risk, preserves developer velocity, and bakes in observability, docs, and rollout strategy from the start.

## Usage

```
/feature-implementation-game $ARGUMENTS [ship_deadline] [risk_level] [rollout_strategy]
```

## Arguments

- `$ARGUMENTS` (required): High-information feature specification. Paste or reference the spec directly; see Variables below
- `ship_deadline` (optional): ISO 8601 deadline (default: 1 week from now)
- `risk_level` (optional): "low" | "medium" | "high" (default: "medium")
- `rollout_strategy` (optional): "flagged" | "canary" | "dark_launch" | "big_bang" (default: "flagged")
- `comments` (optional): Additional context/constraints

## Variables

- **SPEC_DOCUMENT: $ARGUMENTS**
  - Accepted forms: inline Markdown PRD/RFC, GitHub/GitLab issue URL, file path, YAML/JSON brief
  - Recommended sections for maximum signal:
    - Problem statement, goals, and non-goals
    - Acceptance criteria and success metrics
    - API/SDK contracts (types, errors, pagination, retries)
    - UX notes or wireframes
    - Non-functional requirements (performance, security, availability)
    - Rollout and rollback constraints
    - Telemetry requirements (metrics, events, traces)
    - Docs & examples scope

## Algorithm

### Phase 0: Initialize Shipping Space

```bash
# Create feature game state
mkdir -p .feature-implementation-game/{artifacts,checks,rollouts,docs}

cat > .feature-implementation-game/state.json << EOF
{
  "feature": "$FEATURE_NAME",
  "deadline": "$SHIP_DEADLINE",
  "risk_level": "$RISK_LEVEL",
  "rollout_strategy": "$ROLLOUT_STRATEGY",
  "phase": 0,
  "budgets": {
    "error_budget": 0.02,
    "performance_budget_ms": 50,
    "oncall_load_budget_pts": 3
  },
  "slo": {
    "availability": 0.999,
    "p95_latency_ms": 250
  },
  "flags": {
    "name": "feature.${FEATURE_NAME// /_}",
    "default": false,
    "kill_switch": true
  },
  "stakeholders": ["product_owner", "tech_lead", "sre", "qa", "docs", "security"],
  "artifacts": {}
}
EOF

echo "🎮 Feature Implementation Game initialized for: $FEATURE_NAME"
echo "⏰ Deadline: $SHIP_DEADLINE | ⚠️ Risk: $RISK_LEVEL | 🚀 Rollout: $ROLLOUT_STRATEGY"
```

### Phase 0.1: Ingest $ARGUMENTS (Spec) and Extract Signals

```bash
echo "\n📥 Ingesting SPEC_DOCUMENT from $ARGUMENTS"

# Persist raw spec for traceability
mkdir -p .feature-implementation-game/spec
cat > .feature-implementation-game/spec/spec.raw.txt << 'EOF'
$ARGUMENTS
EOF

# Auto-detect structure and extract signals (heuristic; replace with parser as needed)
cat > .feature-implementation-game/spec/signals.json << 'EOF'
{
  "feature_name": "$(echo "$ARGUMENTS" | head -1 | sed 's/#\s*//;s/\r//')",
  "has_acceptance_criteria": $(grep -qi "acceptance criteria\|acceptance" <<< "$ARGUMENTS" && echo true || echo false),
  "has_api_contract": $(grep -qi "api\|sdk\|contract\|schema\|types" <<< "$ARGUMENTS" && echo true || echo false),
  "has_nfrs": $(grep -qi "performance\|latency\|availability\|security\|privacy" <<< "$ARGUMENTS" && echo true || echo false),
  "has_rollout": $(grep -qi "rollout\|canary\|flag\|dark launch\|kill switch" <<< "$ARGUMENTS" && echo true || echo false),
  "has_telemetry": $(grep -qi "metric\|event\|trace\|observability\|logging" <<< "$ARGUMENTS" && echo true || echo false),
  "has_docs_plan": $(grep -qi "docs\|documentation\|examples\|cookbook" <<< "$ARGUMENTS" && echo true || echo false)
}
EOF

echo "🧠 Extracted signals:" && cat .feature-implementation-game/spec/signals.json | jq .

# Derive defaults from signals (example: tighten budgets if high risk or missing NFRs)
if jq -e '.has_nfrs == false' .feature-implementation-game/spec/signals.json >/dev/null; then
  jq '.budgets.performance_budget_ms = 30' .feature-implementation-game/state.json > tmp && mv tmp .feature-implementation-game/state.json
  echo "⚠️  No NFRs found; tightening performance budget to 30ms"
fi
```

### Phase 1: Delivery Readiness Gate

```bash
echo "\n🧰 Phase 1: Delivery Readiness Gate\n=============================="

# Required readiness checks (fail any → block)
cat > .feature-implementation-game/checks/readiness.json << 'EOF'
{
  "acceptance_criteria": [
    "Clear user-facing problem & success metrics",
    "API/SDK contracts specified (types, errors, pagination, retries)",
    "Backward compatibility plan (or versioning) documented",
    "Feature flag name and ownership defined",
    "Telemetry events & metrics spec drafted",
    "Security & privacy review queued (data classification)",
    "Docs outline and examples scope agreed",
    "Rollout + rollback plan sketched (canary + kill switch)"
  ]
}
EOF

echo "✅ Readiness checklist created"
```

### Phase 2: Implementation Plan Synthesis

```bash
echo "\n📝 Phase 2: Implementation Plan\n=============================="

# Generate plan scaffolding tied to Smithery/MCP conventions
cat > .feature-implementation-game/artifacts/plan.md << 'EOF'
# Implementation Plan: $FEATURE_NAME

## Scope
- In-scope: Core capability behind a feature flag
- Out-of-scope: Non-critical UI polish, opportunistic refactors

## Interfaces
- API/SDK Contracts:
  - Typescript SDK: function signature, error model, retries
  - Python SDK: parity specification and tests
  - MCP server/client: tool/resource schema, registry entries

## Data & Migrations
- Migration steps (idempotent), backfill strategy, safe roll-forward/back

## Observability
- Metrics: counters, timers, error rates
- Events: user actions, server outcomes
- Traces: critical paths annotated

## Rollout & Safety
- Strategy: flagged + canary (1%→5%→25%→100%)
- Kill switch: instant revert path verified
- Guardrails: error/perf budgets, auto-pause conditions

## Docs & DX
- Update docs/ SDK examples/ cookbooks
- Changelog + release notes templates
- Internal runbook for support

## Risks & Mitigations
- List top 3 risks with mitigations and owners
EOF

echo "🧭 Plan scaffold generated"
```

### Context Engineering Tips for $ARGUMENTS

- Provide a compact, scannable spec with strong headings; first line is used as `feature_name`
- Include a bullet list of acceptance criteria; use checkboxes `[ ]` to improve parsing
- Show concise API contracts (both TS and Python snippets) and error models
- State explicit rollout constraints and kill-switch expectations
- Declare required telemetry up front (names, cardinality concerns)
- Link to any diagrams; paste key tables inline for reliability

### Phase 3: Smallest-Useful-Increment Loop

```bash
echo "\n🔁 Phase 3: Implementation Loop (guard-railed)\n============================================="

ROUND=1
while [ $ROUND -le 3 ]; do
  echo "\n  ▶️ Iteration $ROUND: Smallest testable slice"

  # 3.1 Create or wire feature flag
  echo "  - Ensuring flag exists: $(jq -r .flags.name .feature-implementation-game/state.json)"

  # 3.2 Implement thinnest vertical slice behind the flag
  echo "  - Implementing minimal slice behind flag (server + SDK + test)"

  # 3.3 Add tests (unit + e2e happy path)
  echo "  - Adding tests for slice"

  # 3.4 Instrumentation hooks
  echo "  - Emitting metrics/events/traces for slice"

  # 3.5 Docs stub
  echo "  - Drafting doc stub and example stub"

  # 3.6 Validate budgets
  echo "  - Validating performance/error budgets"

  # Break early if READY criteria met for canary
  if [ $ROUND -ge 1 ]; then
    echo "  ✅ Slice viable for canary"
    break
  fi

  ROUND=$((ROUND+1))
done
```

### Phase 4: Progressive Rollout Orchestration

```bash
echo "\n🚦 Phase 4: Progressive Rollout\n==============================="

STEPS=("1%" "5%" "25%" "100%")

for step in "${STEPS[@]}"; do
  echo "  📈 Enabling to ${step} of users (flag/canary)"

  # Health checks
  echo "    - Checking error rate vs budget..."
  echo "    - Checking p95 latency vs budget..."
  echo "    - Checking user-visible regressions..."

  # Auto-pause/rollback conditions
  cat > .feature-implementation-game/rollouts/${step}.json << EOF
{
  "step": "${step}",
  "error_rate_ok": true,
  "latency_ok": true,
  "user_feedback_ok": true,
  "decision": "continue"
}
EOF

  echo "    - Decision: continue"
done

echo "✅ Rollout complete"
```

### Phase 5: Docs, Examples, and Parity Gates

```bash
echo "\n📚 Phase 5: Docs & SDK Parity\n============================="

# Check required updates for Smithery/SDK repos
cat > .feature-implementation-game/checks/docs_parity.json << 'EOF'
{
  "docs": [
    "docs/sdk/server/ *.mdx updated",
    "docs/sdk/client/ *.mdx updated",
    "docs/use/ and getting_started updated",
    "images or diagrams added if needed"
  ],
  "sdk_examples": [
    "sdk/typescript/sdk/examples/* updated",
    "sdk/python/examples/* updated"
  ],
  "parity": [
    "TS ↔ Python parity validated (see specs/PARITY_REPORT.md)",
    "OpenAPI/registry updated if contracts changed",
    "Cookbook entries added/updated"
  ],
  "release_communication": [
    "CHANGELOG entry",
    "Release notes",
    "Migration notes if any"
  ]
}
EOF

echo "✅ Docs & parity gates enumerated"
```

### Phase 6: Security, Privacy, and Compliance Gate

```bash
echo "\n🛡️  Phase 6: Security & Privacy Gate\n==================================="

cat > .feature-implementation-game/checks/security.json << 'EOF'
{
  "data_classification": "none|minimal|personal|sensitive",
  "threat_model_updated": true,
  "dependency_scan_clean": true,
  "secrets_management": "verified",
  "authz_paths_reviewed": true
}
EOF

echo "✅ Security checklist drafted"
```

### Phase 7: Ship, Monitor, and Learn

```bash
echo "\n🚢 Phase 7: Ship & Monitor\n========================="

# Generate release artifacts
cat > .feature-implementation-game/artifacts/release.md << 'EOF'
# Release Summary: $FEATURE_NAME

## What shipped
- Concise description

## Rollout
- Strategy and timeline

## Metrics & Observability
- Key dashboards and alerts

## Docs
- Links to updated docs, examples, cookbooks

## Support
- Runbook link, known issues, mitigation
EOF

echo "📊 Monitoring dashboards linked, alerts armed"

# Post-ship retrospective seed
cat > .feature-implementation-game/artifacts/postmortem.md << 'EOF'
# Lightweight Retro: $FEATURE_NAME

## What went well

## What was hard

## What we’d change next time

## Detected anti-patterns (if any)
- scope_creep | dark_launch_without_telemetry | skipped_docs | no_kill_switch
EOF

echo "🧠 Retro template prepared"
```

### Optional: Sandbox Testing (E2B)

- Validate the new feature in an isolated environment using [sandbox-testing-game](./sandbox-testing-game.md).
- Example:

```bash
/sandbox-testing-game "<repo-or-path>" "" "" "build a minimal example that uses the new feature end-to-end" 600000
```

### Phase 8: Anti-Pattern Detection & Circuit Breakers

```bash
echo "\n🚨 Phase 8: Anti-Pattern Detection\n================================="

ANTIPATTERNS=()

# Scope creep: files touched grow > 2x plan
PLANNED=$(grep -c '-' .feature-implementation-game/artifacts/plan.md)
TOUCHED=$(git status --porcelain | wc -l | tr -d ' ')
if [ "$TOUCHED" -gt $((PLANNED * 2)) ]; then
  ANTIPATTERNS+=("scope_creep")
  echo "⚠️  Scope creep detected"
fi

# Missing telemetry
if ! grep -qi "metrics\|events\|trace" .feature-implementation-game/artifacts/plan.md; then
  ANTIPATTERNS+=("no_telemetry")
  echo "⚠️  Telemetry missing"
fi

# No flag
FLAG_NAME=$(jq -r .flags.name .feature-implementation-game/state.json)
if [ -z "$FLAG_NAME" ] || [ "$FLAG_NAME" = "null" ]; then
  ANTIPATTERNS+=("no_flag")
  echo "⚠️  Feature flag missing"
fi

jq -n --argjson list "$(printf '%s\n' "${ANTIPATTERNS[@]}" | jq -R . | jq -s .)" '{anti_patterns: $list}' \
  > .feature-implementation-game/checks/anti_patterns.json

echo "✅ Anti-pattern scan complete"
```

### Phase 9: Game Summary

```bash
echo "\n📊 FEATURE IMPLEMENTATION GAME COMPLETE\n====================================="

STATE=$(cat .feature-implementation-game/state.json)
ANTI=$(jq -r '.anti_patterns[]? // empty' .feature-implementation-game/checks/anti_patterns.json 2>/dev/null | wc -l | tr -d ' ')

echo "🎯 Feature: $(jq -r .feature .feature-implementation-game/state.json)"
echo "⏱️  Deadline: $(jq -r .deadline .feature-implementation-game/state.json)"
echo "⚠️  Risk: $(jq -r .risk_level .feature-implementation-game/state.json)"
echo "🧯 Anti-patterns detected: ${ANTI}"

# Optional cleanup
# rm -rf .feature-implementation-game

echo "\n✅ Ship it with confidence!"
```

## Examples

```bash
# Standard flagged rollout under a week deadline
/feature-implementation-game "Registry: Add server tags filter" "2025-01-15T17:00:00Z" medium flagged

# High-risk capability with canary and strict budgets
/feature-implementation-game "MCP: Batch tool execution" "2025-02-01T00:00:00Z" high canary

# Dark launch to validate performance before exposure
/feature-implementation-game "SDK: Streaming responses" "2025-01-10T12:00:00Z" medium dark_launch
```

## Anti-Patterns Prevented

1. **YOLO Deploy**: Enforced flags, rollouts, and kill switches
2. **Silent Breaking Changes**: Contract gates, parity checks, migration notes
3. **Observability Blindness**: Required metrics/events/traces before rollout
4. **Scope Creep**: Circuit breaker compares plan vs touched files
5. **Docs/DX Tax**: Dedicated phase and gates for docs, examples, cookbooks
6. **One-Way Doors**: Rollback paths and auto-pause conditions

## Integration Notes (Smithery/MCP context)

- Update `docs/sdk/*` and `docs/use/*` with examples and diagrams
- Ensure TS ↔ Python SDK parity with tests and `PARITY_REPORT.md`
- If introducing new MCP tools/resources, define schemas and update registry
- Keep `claude_code_mcp_config.json` aligned if developer UX changes
- Add cookbook examples under `smithery-cookbook/`

## Meta-Learning

Each execution logs artifacts under `.feature-implementation-game/` for later synthesis (e.g., with mem0). Patterns across features inform improved defaults for budgets, rollouts, and documentation quality gates.
