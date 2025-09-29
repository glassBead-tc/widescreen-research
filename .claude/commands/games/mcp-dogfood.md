# /mcp-dogfood

Enumerate, plan, and execute a full dogfooding pass against an MCP server from within the agent. Produces a machine-updated game state JSON and a RESULTS.md with observations and next steps.

## Usage

```
/mcp-dogfood "$ARGUMENTS" [--run_scenarios=true] [--reasoner=sequentialthinking] [--time_window=15m]
```

## Parameters

- `$ARGUMENTS`: Target server id or connection name
- `--run_scenarios`: If true, the agent will execute the planned scenarios (default: true)
- `--reasoner`: Optional reasoning server to structure scenario planning (e.g., sequentialthinking, clear-thought). Default: sequentialthinking
- `--time_window`: Optional time window for logs/metrics correlation

## Reasoning workflow (sequentialthinking example)

When `--reasoner=sequentialthinking` is used, the agent plans dogfood scenarios through an explicit chain-of-thought loop using the server's `sequentialthinking` tool. The workflow looks like this:

1) Initialize thinking context

- Start with `thoughtNumber = 1`, an initial `totalThoughts` estimate (e.g., 3), and `nextThoughtNeeded = true`.

2) Iterative reasoning calls

- For each step, call the tool with:
  - `thought` (free-form plan/reflection content)
  - `thoughtNumber`, `totalThoughts`, `nextThoughtNeeded`
  - Optional flags to control structure:
    - `isRevision`, `revisesThought` to correct prior steps
    - `branchFromThought`, `branchId` to explore alternatives in parallel
    - `needsMoreThoughts` to signal increasing scope mid-run

3) Read back tool output (returned as text JSON)

- Fields include: `thoughtNumber`, `totalThoughts`, `nextThoughtNeeded`, `branches`, `thoughtHistoryLength`.
- Adjust counters and decide whether to continue, revise, or branch.

4) Convergence

- Set `nextThoughtNeeded = false` when the scenario ordering and dependency plan are ready (e.g., entity-creation before dependent operations).

Example call payload (conceptual):

```json
{
  "thought": "Enumerate capabilities; identify create→list→delete dependencies; schedule entity creation first.",
  "thoughtNumber": 1,
  "totalThoughts": 3,
  "nextThoughtNeeded": true
}
```

Example follow-up (revision):

```json
{
  "thought": "Revise plan: websets requires create before enhance; add cleanup step.",
  "thoughtNumber": 2,
  "totalThoughts": 4,
  "isRevision": true,
  "revisesThought": 1,
  "nextThoughtNeeded": true
}
```

The agent uses this structured loop to produce the scenario list stored in `.mcp-dogfood/state.json` under `scenarios`, ensuring preconditions (e.g., create before list/delete) are respected before execution.

## Algorithm

### Phase 0: Initialize Dogfood Arena

```bash
mkdir -p .mcp-dogfood/{state,plans,logs,artifacts}

# Prerequisite: if using `--reasoner=sequentialthinking`, ensure the MCP server at `mcp-servers/sequentialthinking` is running and connected in Claude Code

cat > .mcp-dogfood/state.json << EOF
{
  "target_server": "$ARGUMENTS",
  "enumerated": false,
  "planned": false,
  "executed": false,
  "capabilities": [],
  "scenarios": [],
  "runs": []
}
EOF
```

### Phase 1: Enumerate All Capabilities

```bash
echo "🔎 Enumerating MCP capabilities for: $ARGUMENTS"

# Pseudocode hooks (replace with actual MCP client calls in your environment)
echo "  - Listing tools, resources, and nested operations"

cat > .mcp-dogfood/artifacts/capabilities.json << 'EOF'
{
  "server": "$ARGUMENTS",
  "tools": [
    { "name": "example_tool", "operations": ["opA", "opB"], "schema": {"input": {}, "output": {}} }
  ],
  "resources": [
    { "uri": "example://resource", "mime": "application/json", "metadata": {} }
  ]
}
EOF

jq '.enumerated = true | .capabilities = input' .mcp-dogfood/state.json \
  < .mcp-dogfood/artifacts/capabilities.json > tmp && mv tmp .mcp-dogfood/state.json
```

### Phase 2: Plan Scenarios (Optionally via Reasoner)

```bash
echo "🧠 Planning scenarios (reasoner: $REASONER)"

cat > .mcp-dogfood/plans/scenario-plan.md << 'EOF'
# Dogfood Scenario Plan

## Goals
- Exercise each tool operation and resource access
- Respect dependencies (e.g., create before list/delete)

## Scenario Outline
1. example_tool.opA with valid inputs
2. example_tool.opB (dependent on opA entity)
3. Read example resource
EOF

jq '.planned = true | .scenarios = ["example_tool.opA", "example_tool.opB", "resource.read"]' \
  .mcp-dogfood/state.json > tmp && mv tmp .mcp-dogfood/state.json
```

### Phase 3: Execute Scenarios

```bash
if [ "$RUN_SCENARIOS" = "false" ]; then
  echo "⏭️  Skipping execution per flag"
else
  echo "🏃 Executing scenarios..."

  # Example execution loop (replace with real MCP invocation)
  while read -r scenario; do
    echo "  ▶️  $scenario"
    START=$(date +%s%3N)
    # Simulate MCP call and capture result
    echo '{"success":true,"details":"ok"}' > .mcp-dogfood/logs/$(echo "$scenario" | tr ' /.' '_').json
    END=$(date +%s%3N)
    DURATION=$((END-START))

    # Append run record
    jq --arg s "$scenario" --argjson d $DURATION --slurpfile r .mcp-dogfood/logs/$(echo "$scenario" | tr ' /.' '_').json \
      '.runs += [{scenario: $s, duration_ms: $d, result: $r[0]}]' \
      .mcp-dogfood/state.json > tmp && mv tmp .mcp-dogfood/state.json
  done < <(jq -r '.scenarios[]' .mcp-dogfood/state.json)
fi

jq '.executed = true' .mcp-dogfood/state.json > tmp && mv tmp .mcp-dogfood/state.json
```

### Phase 4: Results Report

```bash
cat > .mcp-dogfood/RESULTS.md << EOF
# MCP Dogfood Results: $ARGUMENTS

## Summary
- Capabilities enumerated: $(jq -r '.capabilities.tools | length' .mcp-dogfood/state.json) tools
- Scenarios executed: $(jq -r '.runs | length' .mcp-dogfood/state.json)

## Runs
$(jq -r '.runs[] | "- \(.scenario): success=\(.result.success), duration_ms=\(.duration_ms)"' .mcp-dogfood/state.json)

## Observations & Next Steps
- Validate error paths and edge cases
- Add negative tests and latency budgets
- Integrate with debug-suite on failures
EOF

echo "✅ MCP dogfood completed"
```

## Integration

- Triggered by `/mcp-server-implementation-game`
- Can be run standalone to validate existing servers: `/mcp-dogfood server_id`
- Pair with `/debug-suite` if issues are detected
