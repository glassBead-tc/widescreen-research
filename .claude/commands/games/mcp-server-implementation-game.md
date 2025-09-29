# /mcp-server-implementation-game

Implement and validate a new MCP server end-to-end with instrumentation, registry integration, and a first-class dogfooding workflow that exercises every exposed capability through the agent.

## Usage

```
/mcp-server-implementation-game $ARGUMENTS [ship_deadline] [risk_level] [transport] [registry_visibility]
```

## Arguments

- `$ARGUMENTS` (required): High-information server specification (PRD/RFC/URL/brief); see Variables
- `ship_deadline` (optional): ISO 8601 deadline (default: 1 week from now)
- `risk_level` (optional): "low" | "medium" | "high" (default: "medium")
- `transport` (optional): "stdio" | "shttp" (default: "stdio")
- `registry_visibility` (optional): "local" | "team" | "public" (default: "local")

## Variables

- SPEC_DOCUMENT: $ARGUMENTS (prewritten specification document)
  - Must include: tool/resource schemas, prompts/models (if any), auth/secret handling, error taxonomy, retries/timeouts, perf goals, observability, rollout
- SERVER_ID: derived from spec first line, slugged
- TRANSPORT: stdio/shttp per argument
- REGISTRY: where to register (local/team/public)

## Algorithm

### Phase 0: Initialize Server Game Space

```bash
mkdir -p .mcp-server-game/{state,artifacts,checks,logs}

cat > .mcp-server-game/state.json << EOF
{
  "server_id": "$SERVER_ID",
  "transport": "$TRANSPORT",
  "registry_visibility": "$REGISTRY",
  "deadline": "$SHIP_DEADLINE",
  "risk_level": "$RISK_LEVEL",
  "phase": 0,
  "budgets": { "error_budget": 0.02, "p95_latency_ms": 250 },
  "instrumentation": { "metrics": true, "events": true, "traces": true },
  "connectivity": { "configured": false, "verified": false },
  "dogfood": { "planned": false, "executed": false, "results_path": null }
}
EOF

# Persist raw spec
cat > .mcp-server-game/spec.raw.txt << 'EOF'
$ARGUMENTS
EOF
```

### Phase 1: Readiness Gate (Spec-Provided Contracts)

```bash
cat > .mcp-server-game/checks/readiness.json << 'EOF'
{
  "contracts": [
    "Spec provides tool schemas (names, params, result types)",
    "Spec provides resource schemas (uris, mime types, metadata)",
    "Spec provides error taxonomy + retry/timeout strategy",
    "Spec documents auth/secret handling",
    "Transport selected (stdio/shttp) with bootstrap",
    "Observability plan (metrics/events/traces) present in spec"
  ],
  "registry": [
    "Server metadata prepared for registry",
    "Visibility chosen (local/team/public)"
  ]
}
EOF

echo "✅ Readiness checklist created"
```

### Phase 2: Minimal Vertical Slice Behind a Flag

```bash
echo "\n🔧 Implement minimal slice: one tool + optional resource"

# Server bootstrap (placeholder)
echo "  - Bootstrapping $SERVER_ID over $TRANSPORT"
echo "  - Registering example tool: health_check"
echo "  - Adding metrics/events/traces hooks"

echo "\n🧪 Unit tests for slice"
echo "  - Add tests for tool schema validation and error paths"
```

### Phase 3: Connectivity and Registry Integration

```bash
echo "\n🔌 Connectivity"
echo "  - Configure client connection to $SERVER_ID"
echo "  - Verify handshake and list capabilities"

jq '.connectivity.configured = true' .mcp-server-game/state.json > tmp && mv tmp .mcp-server-game/state.json

echo "\n📇 Registry"
echo "  - Prepare registry entry (name, description, transport, tags)"
echo "  - Visibility: $REGISTRY"
```

### Phase 4: Dogfood Test Plan Generation

```bash
echo "\n🐶 Dogfood plan generation"

# Create plan input for /mcp-dogfood
cat > .mcp-server-game/artifacts/dogfood.input.json << 'EOF'
{
  "target_server": "$SERVER_ID",
  "reasoner": "sequentialthinking",
  "time_window": "15m"
}
EOF

jq '.dogfood.planned = true' .mcp-server-game/state.json > tmp && mv tmp .mcp-server-game/state.json

echo "  - Invoking /mcp-dogfood to enumerate capabilities and draft scenario"
/mcp-dogfood "$SERVER_ID" --run_scenarios=true --reasoner=sequentialthinking 2>&1 | tee .mcp-server-game/logs/dogfood.log || true
```

### Phase 5: Rollout & Safety (Internal)

```bash
echo "\n🚦 Internal rollout (developer-only)"
echo "  - Enable server for local agents only"
echo "  - Auto-pause on error budget breach or p95 > budget"
```

### Phase 6: Docs & Examples

```bash
echo "\n📚 Docs & examples"
cat > .mcp-server-game/artifacts/README.snippet.md << 'EOF'
# $SERVER_ID MCP Server

## Capabilities
- Tools: [...]
- Resources: [...]

## Connect
```bash
smithery connect $SERVER_ID
```

## Examples

```bash
# Call a tool
/call "$SERVER_ID" tool_name '{"param": "value"}'
```

EOF

```

### Phase 7: Execute Dogfood and Gate

```bash
echo "\n🏃 Running dogfood end-to-end"

# Expect /mcp-dogfood to write a RESULTS.md and state JSON under .mcp-dogfood/
RESULTS=".mcp-dogfood/RESULTS.md"
if [ -f "$RESULTS" ]; then
  jq --arg p "$RESULTS" '.dogfood.executed = true | .dogfood.results_path = $p' \
    .mcp-server-game/state.json > tmp && mv tmp .mcp-server-game/state.json
  echo "✅ Dogfood completed: $RESULTS"
else
  echo "⚠️  Dogfood results not found; investigate .mcp-server-game/logs/dogfood.log"
fi
```

### Phase 8: Summary

```bash
cat > .mcp-server-game/SUMMARY.md << EOF
# MCP Server Implementation Game Summary

- Server: $SERVER_ID
- Transport: $TRANSPORT
- Registry: $REGISTRY
- Dogfood: $(jq -r '.dogfood.executed' .mcp-server-game/state.json)
- Results: $(jq -r '.dogfood.results_path' .mcp-server-game/state.json)

## Next Steps
- Expand capabilities and update schemas
- Broaden rollout beyond internal if stable
- Address issues from dogfood RESULTS.md
EOF

echo "✅ MCP server implementation game complete"
```

## Related Commands

- `/mcp-dogfood` — enumerate capabilities, draft scenario, execute simulations, and generate results
- `/debug-suite` — orchestrated debugging if issues arise during dogfood
- `/meta-learning-suite` — harvest learnings after the run
- `/refactoring-game` — apply targeted improvements identified by dogfood
