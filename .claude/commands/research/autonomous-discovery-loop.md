# Autonomous Discovery Loop
**Pattern**: Systematic hypothesis exploration with iterative refinement
**Based on**: "Agentic Science" framework (arXiv:2508.14111)
**Determinism**: High - Each phase has programmatic validation gates

---

## Usage

```bash
/autonomous-discovery-loop "Research question or domain"
```

**Arguments**:
- `$ARGUMENTS` (required): Research question or domain to explore
- `max_iterations` (optional): Maximum discovery loops (default: 5)
- `hypothesis_breadth` (optional): Number of hypotheses per iteration (default: 10)
- `convergence_threshold` (optional): Stop when X% of hypotheses validated (default: 0.80)

---

## Variables

**Session State** (tracked in `.discovery-loop/state.json`):
```json
{
  "research_question": "$ARGUMENTS",
  "iteration": 0,
  "max_iterations": 5,
  "hypotheses_generated": [],
  "hypotheses_tested": [],
  "hypotheses_validated": [],
  "evidence_collected": [],
  "current_phase": "observe",
  "convergence_score": 0.0
}
```

---

## Workflow Phases

### Phase 0: Initialize Discovery Space

**Deterministic Hook**: `.claude/hooks/research/init-discovery-loop.sh`

```bash
#!/bin/bash
# Creates research state directory and validates inputs

RESEARCH_QUESTION="$1"
MAX_ITER="${2:-5}"
HYPOTHESIS_BREADTH="${3:-10}"

mkdir -p .discovery-loop/{hypotheses,evidence,tests,synthesis}

cat > .discovery-loop/state.json << EOF
{
  "research_question": "$RESEARCH_QUESTION",
  "iteration": 0,
  "max_iterations": $MAX_ITER,
  "hypothesis_breadth": $HYPOTHESIS_BREADTH,
  "current_phase": "observe",
  "gates": {
    "observe": {"passed": false, "evidence_count": 0},
    "hypothesize": {"passed": false, "hypotheses_count": 0},
    "design": {"passed": false, "tests_defined": 0},
    "execute": {"passed": false, "tests_run": 0},
    "analyze": {"passed": false, "results_recorded": 0}
  }
}
EOF

echo "✅ Discovery loop initialized for: $RESEARCH_QUESTION"
```

**Gate**: State file created, inputs validated

---

### Phase 1: Observe (LLM Intelligence + Agent Execution)

**LLM Task** (Intelligence - Non-Deterministic):
```
Generate initial research queries to gather domain knowledge:

1. What are the key concepts in [research_question]?
2. What existing research exists?
3. What are the known controversies?
4. What methodologies are used?
5. What gaps exist?

Provide 5-10 specific search queries for agents to execute.
```

**Agent Task** (Deterministic - Hook Enforced):

PreToolUse Hook validates:
- Query is from LLM (not agent-generated)
- Query is specific and executable
- No interpretation in query

Agent executes:
- `arxiv-paper-mcp__search_papers` with LLM-provided keywords
- `exa__web_search_exa` for domain context
- `exa__get_code_context_exa` for technical details
- `firecrawl__firecrawl_scrape` on key URLs

**Returns**:
```json
{
  "phase": "observe",
  "evidence_collected": 47,
  "sources": ["arxiv", "web", "code", "docs"],
  "raw_data": "[All retrieved content]",
  "metadata": {
    "unique_concepts": 125,
    "date_range": "2020-2025",
    "domains": ["cs.AI", "cs.LG", "cs.MA"]
  }
}
```

**Gate Check** (`.claude/hooks/research/check-observe-gate.sh`):
```bash
#!/bin/bash
STATE_FILE=".discovery-loop/state.json"

EVIDENCE_COUNT=$(jq -r '.gates.observe.evidence_count' "$STATE_FILE")
MIN_REQUIRED=20

if [ "$EVIDENCE_COUNT" -ge "$MIN_REQUIRED" ]; then
  jq '.gates.observe.passed = true | .current_phase = "hypothesize"' "$STATE_FILE" > tmp && mv tmp "$STATE_FILE"
  echo "✅ Observe gate passed: $EVIDENCE_COUNT sources collected"
  exit 0
else
  echo "❌ Observe gate failed: Only $EVIDENCE_COUNT sources (need $MIN_REQUIRED)"
  exit 1
fi
```

**Gate**: Minimum 20 evidence sources collected (deterministic count)

---

### Phase 2: Hypothesize (LLM Intelligence)

**LLM Task**:
```
Based on evidence from Phase 1, generate hypotheses:

Input: [Raw evidence from agents]
Generate: 10 testable hypotheses about [research_question]

Each hypothesis must:
1. Be specific and testable
2. Explain a pattern in the evidence
3. Make a falsifiable prediction
4. Include success criteria

Output format:
{
  "hypothesis_id": "H001",
  "statement": "[Specific claim]",
  "test_criteria": "[How to validate]",
  "expected_evidence": "[What would confirm this]",
  "confidence": 0.0-1.0
}
```

**Agent Task** (Deterministic):
- Store hypotheses to `.discovery-loop/hypotheses/iteration-N.json`
- Count: Track number of hypotheses
- No evaluation, no filtering, no intelligence

**Gate Check**:
```bash
#!/bin/bash
HYPOTHESIS_FILE=".discovery-loop/hypotheses/iteration-$(jq -r '.iteration' .discovery-loop/state.json).json"

if [ ! -f "$HYPOTHESIS_FILE" ]; then
  echo "❌ No hypotheses file found"
  exit 1
fi

HYPO_COUNT=$(jq '. | length' "$HYPOTHESIS_FILE")
MIN_REQUIRED=5

if [ "$HYPO_COUNT" -ge "$MIN_REQUIRED" ]; then
  jq ".gates.hypothesize.passed = true | .gates.hypothesize.hypotheses_count = $HYPO_COUNT | .current_phase = \"design\"" .discovery-loop/state.json > tmp && mv tmp .discovery-loop/state.json
  echo "✅ Hypothesize gate passed: $HYPO_COUNT hypotheses generated"
  exit 0
else
  echo "❌ Need at least $MIN_REQUIRED hypotheses, got $HYPO_COUNT"
  exit 1
fi
```

**Gate**: Minimum 5 hypotheses generated (deterministic count)

---

### Phase 3: Design Tests (LLM Intelligence)

**LLM Task**:
```
For each hypothesis, design validation approach:

Hypothesis: [H001]
Design:
  - What evidence would confirm this?
  - What evidence would refute this?
  - What sources to check?
  - What queries to run?
  - What metrics to measure?

Output: Executable test specification (search queries, comparison criteria)
```

**Agent Task** (Deterministic):
- Store test designs
- Validate test specifications have required fields
- Track: Number of tests designed

**Gate Check**:
```bash
#!/bin/bash
# Each hypothesis must have a test design

HYPO_COUNT=$(jq -r '.gates.hypothesize.hypotheses_count' .discovery-loop/state.json)
TEST_COUNT=$(find .discovery-loop/tests -name "*.json" | wc -l)

if [ "$TEST_COUNT" -ge "$HYPO_COUNT" ]; then
  jq ".gates.design.passed = true | .gates.design.tests_defined = $TEST_COUNT | .current_phase = \"execute\"" .discovery-loop/state.json > tmp && mv tmp .discovery-loop/state.json
  echo "✅ Design gate passed: $TEST_COUNT tests defined"
  exit 0
else
  echo "❌ Need $HYPO_COUNT tests, only $TEST_COUNT defined"
  exit 1
fi
```

**Gate**: One test per hypothesis (deterministic count match)

---

### Phase 4: Execute Tests (Agent Execution)

**Agent Task** (Fully Deterministic):

For each test in `.discovery-loop/tests/`:
1. Read test specification
2. Execute specified queries (no intelligence)
3. Store raw results
4. Track: Test ID, status (pending/complete), result count

**Example test execution**:
```bash
# Agent reads: test-H001.json
{
  "hypothesis_id": "H001",
  "queries": [
    {"tool": "arxiv-paper-mcp__search_papers", "keyword": "X"},
    {"tool": "exa__web_search_exa", "query": "Y"}
  ],
  "success_criteria": "Found > 5 papers supporting claim"
}

# Agent executes queries deterministically
# Stores results to .discovery-loop/evidence/H001-results.json
# Updates state: tests_run += 1
```

**Gate Check**:
```bash
#!/bin/bash
TEST_COUNT=$(jq -r '.gates.design.tests_defined' .discovery-loop/state.json)
COMPLETED=$(find .discovery-loop/evidence -name "*-results.json" | wc -l)

if [ "$COMPLETED" -ge "$TEST_COUNT" ]; then
  jq ".gates.execute.passed = true | .gates.execute.tests_run = $COMPLETED | .current_phase = \"analyze\"" .discovery-loop/state.json > tmp && mv tmp .discovery-loop/state.json
  echo "✅ Execute gate passed: All $TEST_COUNT tests completed"
  exit 0
else
  echo "❌ Only $COMPLETED/$TEST_COUNT tests completed"
  exit 1
fi
```

**Gate**: All tests executed (deterministic completion count)

---

### Phase 5: Analyze Results (LLM Intelligence)

**LLM Task**:
```
Evaluate each hypothesis against test results:

For H001:
  Evidence for: [List supporting findings]
  Evidence against: [List refuting findings]
  Strength of evidence: [Strong/Moderate/Weak]
  Conclusion: [Validated/Refuted/Inconclusive]
  Confidence: [0.0-1.0]
  Next steps: [What to do with this finding]

Synthesis:
  - Which hypotheses validated?
  - What new questions emerged?
  - What patterns across hypotheses?
  - Should we iterate or conclude?
```

**Agent Task** (Deterministic):
- Store LLM analysis to `.discovery-loop/synthesis/iteration-N.json`
- Count validated vs. refuted hypotheses
- Calculate convergence score

**Gate Check**:
```bash
#!/bin/bash
ANALYSIS_FILE=".discovery-loop/synthesis/iteration-$(jq -r '.iteration' .discovery-loop/state.json).json"

if [ ! -f "$ANALYSIS_FILE" ]; then
  echo "❌ No analysis file found"
  exit 1
fi

# Check if LLM made decision to iterate or conclude
DECISION=$(jq -r '.decision' "$ANALYSIS_FILE")

if [ "$DECISION" = "conclude" ] || [ "$DECISION" = "iterate" ]; then
  jq ".gates.analyze.passed = true | .current_phase = \"decide\"" .discovery-loop/state.json > tmp && mv tmp .discovery-loop/state.json
  echo "✅ Analyze gate passed: Decision = $DECISION"
  exit 0
else
  echo "❌ Analysis incomplete: No decision (iterate/conclude) made"
  exit 1
fi
```

**Gate**: LLM must make explicit decision (deterministic presence check)

---

### Phase 6: Decide Next Action (LLM Intelligence with Deterministic Constraints)

**LLM Decision** (Intelligence):
```
Based on analysis, choose:

A. CONCLUDE: If convergence_score > threshold
   → Proceed to final synthesis

B. ITERATE: If iteration < max_iterations AND new questions emerged
   → Generate new hypotheses and loop to Phase 1

C. EXPAND: If promising branch needs deeper exploration
   → Focus next iteration on specific sub-domain

Provide: Explicit decision + rationale
```

**Deterministic Constraint Check** (Hook):
```bash
#!/bin/bash
# Enforce iteration limits

CURRENT_ITER=$(jq -r '.iteration' .discovery-loop/state.json)
MAX_ITER=$(jq -r '.max_iterations' .discovery-loop/state.json)
DECISION=$(jq -r '.decision' .discovery-loop/synthesis/iteration-$CURRENT_ITER.json)

if [ "$DECISION" = "iterate" ]; then
  if [ "$CURRENT_ITER" -ge "$MAX_ITER" ]; then
    echo "🚫 BLOCKED: Cannot iterate (at max_iterations=$MAX_ITER)"
    echo "Forcing decision to CONCLUDE"
    jq '.decision = "conclude" | .decision_override = "max_iterations_reached"' .discovery-loop/synthesis/iteration-$CURRENT_ITER.json > tmp && mv tmp .discovery-loop/synthesis/iteration-$CURRENT_ITER.json
    exit 1
  fi
fi

# Update iteration counter if iterating
if [ "$DECISION" = "iterate" ]; then
  jq ".iteration += 1 | .current_phase = \"observe\"" .discovery-loop/state.json > tmp && mv tmp .discovery-loop/state.json
  echo "✅ Proceeding to iteration $(jq -r '.iteration' .discovery-loop/state.json)"
fi

exit 0
```

**Gate**: Iteration limit enforced (deterministic), but LLM decides strategy (intelligence)

---

## Hook Configuration

Add to `.claude/hooks.json`:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "mcp__arxiv-paper-mcp__.*|mcp__exa__.*|mcp__firecrawl-mcp__.*",
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/research/validate-discovery-phase.sh",
            "description": "Ensure we're in correct discovery phase before retrieval"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "mcp__arxiv-paper-mcp__.*|mcp__exa__.*|mcp__firecrawl-mcp__.*",
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/research/record-evidence.sh",
            "description": "Track evidence collection and update phase gates"
          }
        ]
      }
    ]
  }
}
```

---

## Balance: Determinism vs. Agency

### What's Deterministic (Enforced by Hooks):

✅ **Phase progression**: Can't skip phases (observe → hypothesize → design → execute → analyze)
✅ **Iteration limits**: Max iterations enforced programmatically
✅ **Gate criteria**: Minimum evidence counts, test completion, etc.
✅ **State tracking**: All agent actions recorded to state.json
✅ **Evidence storage**: Raw data stored without interpretation

### What's Intelligent (LLM Decides):

🧠 **What to observe**: Which queries to run, which sources to check
🧠 **Hypothesis generation**: What explanations to propose
🧠 **Test design**: How to validate each hypothesis
🧠 **Result interpretation**: What evidence means, confidence levels
🧠 **Strategy decisions**: Iterate, conclude, or expand focus

---

## Example Execution

```bash
$ /autonomous-discovery-loop "Why do some neural networks exhibit subliminal learning?"

# Phase 0: Initialize
✅ Discovery loop initialized
📊 State: .discovery-loop/state.json created

# Phase 1: Observe
🧠 LLM generates 8 search queries
🤖 Agent executes queries (arxiv, exa, firecrawl)
📊 Collected 52 sources
✅ Observe gate: PASSED (52 >= 20)

# Phase 2: Hypothesize
🧠 LLM analyzes evidence, generates 10 hypotheses:
   H001: Same-base-model requirement suggests geometric signal
   H002: Transmission scales with model size
   H003: Non-semantic patterns in token co-occurrence
   ... (7 more)
📊 Stored to .discovery-loop/hypotheses/iteration-0.json
✅ Hypothesize gate: PASSED (10 >= 5)

# Phase 3: Design Tests
🧠 LLM designs validation for each hypothesis
📊 10 test specifications created
✅ Design gate: PASSED (10 tests = 10 hypotheses)

# Phase 4: Execute Tests
🤖 Agent runs all tests deterministically
📊 Results: 10/10 complete
✅ Execute gate: PASSED (all complete)

# Phase 5: Analyze
🧠 LLM evaluates results:
   H001: VALIDATED (high confidence)
   H002: VALIDATED (moderate confidence)
   H003: REFUTED (strong counter-evidence)
   H004: INCONCLUSIVE
   ...
📊 Convergence: 6/10 validated = 60%
🧠 LLM decides: ITERATE (new questions emerged from H001, H002)

# Gate Check: Iteration allowed? (0 < 5)
✅ Proceeding to iteration 1

[Loop continues...]
```

---

## Hook Scripts

### `.claude/hooks/research/validate-discovery-phase.sh`

```bash
#!/bin/bash
# Validates phase progression before allowing retrieval

TOOL_INPUT=$(cat)
TOOL_NAME=$(echo "$TOOL_INPUT" | jq -r '.toolName')

STATE_FILE=".discovery-loop/state.json"
[ ! -f "$STATE_FILE" ] && exit 0  # No active loop

CURRENT_PHASE=$(jq -r '.current_phase' "$STATE_FILE")

# Check if current phase allows this tool use
case "$CURRENT_PHASE" in
  "observe"|"execute")
    # Retrieval allowed in these phases
    exit 0
    ;;
  "hypothesize"|"design"|"analyze")
    # Block retrieval in thinking phases
    cat << EOF
{
  "block": true,
  "message": "🚫 Phase Violation: Cannot use retrieval tools during '$CURRENT_PHASE' phase.\n\nCurrent phase requires LLM intelligence only.\nAgent execution allowed in: observe, execute phases.\n\nCurrent state: $(cat $STATE_FILE | jq -c)"
}
EOF
    exit 1
    ;;
esac
```

### `.claude/hooks/research/record-evidence.sh`

```bash
#!/bin/bash
# Records evidence collection and checks gates

TOOL_RESULT=$(cat)
TOOL_NAME=$(echo "$TOOL_RESULT" | jq -r '.toolName')

STATE_FILE=".discovery-loop/state.json"
[ ! -f "$STATE_FILE" ] && exit 0

# Increment evidence counter
CURRENT_COUNT=$(jq -r '.gates.observe.evidence_count' "$STATE_FILE")
NEW_COUNT=$((CURRENT_COUNT + 1))

jq ".gates.observe.evidence_count = $NEW_COUNT" "$STATE_FILE" > tmp && mv tmp "$STATE_FILE"

echo "📊 Evidence collected: $NEW_COUNT sources"

# Auto-check gate if in observe phase
CURRENT_PHASE=$(jq -r '.current_phase' "$STATE_FILE")
if [ "$CURRENT_PHASE" = "observe" ]; then
  .claude/hooks/research/check-observe-gate.sh
fi

exit 0
```

---

## Output Format

### Final Synthesis (After Convergence)

```markdown
# Discovery Loop Results: [Research Question]

## Iterations Completed: N

## Validated Hypotheses (High Confidence):
1. **H001**: [Statement]
   - Evidence: 15 sources supporting
   - Confidence: 0.92
   - Key insight: [What this tells us]

2. **H002**: [Statement]
   - Evidence: 12 sources supporting
   - Confidence: 0.85
   - Key insight: [What this tells us]

## Refuted Hypotheses:
- **H003**: [Statement] - Refuted by [counter-evidence]
- **H007**: [Statement] - No supporting evidence found

## Inconclusive (Needs More Research):
- **H004**: [Statement] - Conflicting evidence, 3 for / 3 against

## Novel Discoveries:
- [Unexpected pattern found during exploration]
- [Cross-domain connection identified]

## Research Gaps Identified:
- [Gap 1]: No papers address [specific aspect]
- [Gap 2]: Methodology limitations in [area]

## Recommended Next Steps:
1. [Action based on validated hypotheses]
2. [Further research needed for inconclusive]
3. [Experimental validation of key claims]

---
**Methodology**: Autonomous Discovery Loop (5 iterations)
**Evidence Base**: 247 sources across 5 iterations
**Hypotheses Tested**: 42 total (26 validated, 8 refuted, 8 inconclusive)
**Convergence**: 0.81 (above 0.80 threshold)
```

---

## Benefits

### Qualitative Advantages:
- **Systematic hypothesis exploration**: No human cognitive limits
- **Deterministic audit trail**: Every decision recorded
- **Reproducible**: Re-run produces same exploration path
- **Exhaustive**: Tests hypotheses humans would skip as "unlikely"

### vs. Manual Research:
- **Coverage**: 10x more hypotheses explored
- **Rigor**: Every hypothesis gets test design
- **Bias**: No confirmation bias in hypothesis testing
- **Memory**: No lost threads across iterations

### vs. Unstructured Agent Research:
- **Accountability**: Hooks enforce phase discipline
- **Quality**: Gates ensure minimum evidence standards
- **Control**: Iteration limits prevent runaway loops
- **Transparency**: State file shows exactly where we are

---

## Anti-Patterns Prevented by Hooks

❌ **Skipping observation**: Hook blocks hypothesis generation until evidence gate passed
❌ **Untested hypotheses**: Hook blocks iteration until all hypotheses have tests
❌ **Infinite loops**: Hook enforces max_iterations
❌ **Agent interpretation**: Hooks ensure agents only retrieve, never interpret
❌ **Phase confusion**: Hook blocks wrong tools in wrong phases

---

## Customization

### Adjust Gate Criteria

Edit hook scripts to change thresholds:
```bash
# In check-observe-gate.sh
MIN_REQUIRED=20  # Change to 10 for faster loops, 50 for exhaustive
```

### Adjust Iteration Budget

```bash
/autonomous-discovery-loop "Question" 10 20
# 10 iterations max, 20 hypotheses per iteration
```

### Add Custom Validation

Create `.claude/hooks/research/custom-gate.sh` for domain-specific validation

---

**Status**: Active research pattern with deterministic enforcement
**Requires**: arxiv-paper-mcp, exa, firecrawl MCP servers
**Validation**: State tracker pattern compliant ✅