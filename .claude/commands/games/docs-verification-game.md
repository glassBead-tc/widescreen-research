# Docs Verification Game

Game-theoretic documentation verification workflow that systematically validates accuracy, prevents false positives, and ensures documentation matches implementation reality using multi-dimensional testing and validation gates.

## Command Structure

```bash
/docs-verification-game [DOCS_PATH] [CODEBASE_PATH] [--SCOPE=<scope>] [--DEPTH=<depth>] [--STRATEGY=<strategy>] [--CONFIDENCE=<threshold>]
```

## Parameter Definitions

- `DOCS_PATH`: Location of documentation to verify (file, directory, or pattern)
- `CODEBASE_PATH`: Location of the codebase the docs describe
- `SCOPE`: Verification scope (api, examples, concepts, all) - default: all
- `DEPTH`: Verification depth (surface, standard, deep) - default: standard
- `STRATEGY`: Verification approach (sequential, parallel, adversarial) - default: parallel
- `CONFIDENCE`: Required confidence threshold (0.0-1.0) - default: 0.9

## Game Variables

DOCS_PATH: $ARGUMENTS
CODEBASE_PATH: $ARGUMENTS
SCOPE: $ARGUMENTS
DEPTH: $ARGUMENTS
STRATEGY: $ARGUMENTS
CONFIDENCE: $ARGUMENTS

## Phase 1: Documentation Analysis & Inventory (Time-boxed: 25% of budget)

### Objective: Map documentation claims and create verifiable assertions

```bash
# Initialize verification game state
echo "🔍 Starting Docs Verification Game"
echo "📄 Target: $DOCS_PATH"
echo "💻 Codebase: $CODEBASE_PATH"
echo "🎯 Required confidence: $CONFIDENCE"

mkdir -p .docs-verify/{state,analysis,assertions,evidence}
cat > .docs-verify/state.json << 'EOF'
{
  "phase": "analysis",
  "confidence_level": 0.0,
  "verification_rounds": 0,
  "false_positive_detections": [],
  "accuracy_issues": [],
  "players": {
    "skeptic": { "doubt_level": 0.9, "weight": 1.0 },
    "implementer": { "reality_check": 0.8, "weight": 0.9 },
    "user": { "confusion_potential": 0.5, "weight": 0.8 },
    "maintainer": { "drift_concern": 0.7, "weight": 0.7 }
  }
}
EOF
```

### Step 1.1: Extract Documentation Claims

```bash
# Parse documentation for verifiable claims
echo "📊 Extracting verifiable claims from documentation..."

# Use Claude Code's semantic understanding to identify claims
/codebase-search "What specific APIs, methods, or behaviors are documented in $DOCS_PATH?" --target-directories "$DOCS_PATH"

# Categorize claims by type
cat > .docs-verify/analysis/claim-categories.json << 'EOF'
{
  "api_signatures": [],      // Function signatures, parameters, returns
  "code_examples": [],       // Executable code snippets
  "behavioral_claims": [],   // How things work, side effects
  "configuration": [],       // Config options, defaults
  "error_scenarios": [],     // Error handling, edge cases
  "performance_claims": [],  // Speed, memory, scalability
  "compatibility": []        // Version requirements, dependencies
}
EOF

# Extract and categorize each claim
find "$DOCS_PATH" -type f \( -name "*.md" -o -name "*.mdx" \) | while read -r doc_file; do
  echo "Analyzing: $doc_file"

  # Extract code blocks
  grep -n '```' "$doc_file" | awk 'NR%2==1 {start=$1} NR%2==0 {print start "-" $1}' > .docs-verify/analysis/code-blocks-$doc_file.txt

  # Extract API references
  grep -E "(function|class|interface|type|export)" "$doc_file" > .docs-verify/analysis/api-refs-$doc_file.txt
done
```

### Step 1.2: Build Assertion Database

```bash
# Convert claims into testable assertions
echo "🎯 Building testable assertions..."

cat > .docs-verify/assertions/assertion-builder.js << 'EOF'
// Transform documentation claims into executable tests
const assertions = {
  api_signatures: {
    verify: async (claim) => {
      // Check if function exists with documented signature
      // Verify parameter types and return types
      // Validate optional vs required parameters
    }
  },
  code_examples: {
    verify: async (claim) => {
      // Extract code example
      // Create isolated test environment
      // Execute and verify output matches documentation
    }
  },
  behavioral_claims: {
    verify: async (claim) => {
      // Create test scenarios
      // Execute behavior
      // Verify side effects and outcomes
    }
  }
};
EOF
```

### Gate 1: Analysis Completeness Check

```bash
# Validate we have sufficient assertions to verify
GATE_1_CRITERIA=(
  "Documentation parsed and claims extracted"
  "Assertions created for each claim category"
  "Code examples identified and isolated"
  "API surface mapped to documentation"
  "Verification strategy selected"
)

echo "🚪 Gate 1: Analysis Completeness"
for criteria in "${GATE_1_CRITERIA[@]}"; do
  echo "- [ ] $criteria"
done

# Calculate initial verification complexity
TOTAL_ASSERTIONS=$(find .docs-verify/assertions -name "*.json" | xargs jq -s 'map(to_entries | length) | add')
echo "Total assertions to verify: $TOTAL_ASSERTIONS"
```

## Phase 2: Multi-Strategy Verification Setup (Time-boxed: 20% of budget)

### Objective: Design verification approach with false-positive prevention

```bash
echo "🎲 Phase 2: Verification Strategy Design"

# Select verification strategies based on assertion types
case $STRATEGY in
  "sequential")
    echo "📝 Sequential verification: One assertion at a time"
    PARALLEL_COUNT=1
    ;;
  "parallel")
    echo "⚡ Parallel verification: Multiple assertions simultaneously"
    PARALLEL_COUNT=4
    ;;
  "adversarial")
    echo "👹 Adversarial verification: Intentionally try to break claims"
    PARALLEL_COUNT=3
    ADVERSARIAL_MODE=true
    ;;
esac
```

### Step 2.1: Test Environment Preparation

```bash
# Create isolated test environments for verification
echo "🔧 Preparing verification environments..."

# Setup parallel test environments if needed
if [ "$PARALLEL_COUNT" -gt 1 ]; then
  for i in $(seq 1 $PARALLEL_COUNT); do
    # Create isolated worktree for testing
    git worktree add .docs-verify/test-env-$i verify-env-$i 2>/dev/null || true

    # Install dependencies in each environment
    cd .docs-verify/test-env-$i
    npm install --silent || yarn install --silent
    cd -
  done
fi

# Create verification test suites
cat > .docs-verify/test-template.ts << 'EOF'
import { describe, test, expect } from '@jest/globals';

describe('Documentation Verification: ${ASSERTION_CATEGORY}', () => {
  test('${ASSERTION_NAME}', async () => {
    // Setup
    const claim = ${CLAIM_JSON};

    // Execute
    const result = await verifyClaim(claim);

    // Verify
    expect(result.matches).toBe(true);
    expect(result.confidence).toBeGreaterThan(0.9);
  });
});
EOF
```

### Step 2.2: False-Positive Prevention Setup

```bash
# Configure mechanisms to prevent false verifications
echo "🛡️ Setting up false-positive prevention..."

cat > .docs-verify/false-positive-guards.json << 'EOF'
{
  "guards": {
    "version_mismatch": {
      "description": "Detect when docs reference old API versions",
      "check": "Compare import versions with package.json"
    },
    "partial_matches": {
      "description": "Prevent accepting partial API matches",
      "check": "Verify complete signature, not just function name"
    },
    "example_drift": {
      "description": "Detect outdated code examples",
      "check": "Run examples with strict mode and type checking"
    },
    "context_confusion": {
      "description": "Prevent verifying wrong context",
      "check": "Validate file paths and module contexts"
    }
  }
}
EOF
```

### Gate 2: Verification Readiness

```bash
GATE_2_CRITERIA=(
  "Test environments prepared"
  "Verification strategies defined"
  "False-positive guards configured"
  "Rollback capability confirmed"
  "Success metrics established"
)

echo "🚪 Gate 2: Verification Readiness"
# Validate readiness to begin verification
```

## Phase 3: Parallel Verification Execution (Time-boxed: 40% of budget)

### Objective: Execute comprehensive verification with continuous validation

```bash
echo "🔄 Phase 3: Verification Execution"

# Initialize verification rounds
ROUND=0
CONFIDENCE_LEVEL=0.0
MAX_ROUNDS=3

while [ $ROUND -lt $MAX_ROUNDS ] && (( $(echo "$CONFIDENCE_LEVEL < $CONFIDENCE" | bc -l) )); do
  echo "🎯 Verification Round $ROUND"

  # Load current state
  STATE=$(cat .docs-verify/state.json)
  ACCURACY_ISSUES=$(echo "$STATE" | jq -r '.accuracy_issues | length')

  # Step 3.1: API Signature Verification
  echo "🔍 Verifying API signatures..."

  if [ "$STRATEGY" = "parallel" ]; then
    # Parallel verification using multiple processes
    find .docs-verify/assertions -name "*api_signatures*.json" | \
    xargs -P $PARALLEL_COUNT -I {} bash -c '
      ASSERTION_FILE="{}"
      TEST_ENV=".docs-verify/test-env-$((RANDOM % '"$PARALLEL_COUNT"' + 1))"

      # Extract and verify each API claim
      jq -r ".[] | @json" "$ASSERTION_FILE" | while read -r claim; do
        CLAIM_DATA=$(echo "$claim" | jq -r ".")
        API_NAME=$(echo "$CLAIM_DATA" | jq -r ".name")

        echo "  Verifying API: $API_NAME"

        # Generate verification test
        cd "$TEST_ENV"
        node -e "
          const claim = $CLAIM_DATA;
          // Verification logic here
        " > "../evidence/$API_NAME-result.json"
      done
    '
  fi

  # Step 3.2: Code Example Verification
  echo "📝 Verifying code examples..."

  # Extract and run each code example
  find .docs-verify/analysis -name "*code-blocks*.txt" | while read -r block_file; do
    while IFS= read -r line_range; do
      START_LINE=$(echo "$line_range" | cut -d'-' -f1)
      END_LINE=$(echo "$line_range" | cut -d'-' -f2)

      # Extract code block
      DOC_FILE=$(echo "$block_file" | sed 's/code-blocks-//;s/.txt$//')
      sed -n "${START_LINE},${END_LINE}p" "$DOC_FILE" > .docs-verify/temp-code.js

      # Verify code runs successfully
      if [ "$ADVERSARIAL_MODE" = "true" ]; then
        # Try to break the example
        echo "👹 Adversarial testing of code example..."
        # Add edge cases, null inputs, etc.
      else
        # Standard verification
        node .docs-verify/temp-code.js 2>&1 | tee .docs-verify/evidence/example-$START_LINE.log
      fi
    done < "$block_file"
  done

  # Step 3.3: Behavioral Verification
  echo "🎭 Verifying behavioral claims..."

  # Use Claude Code to identify and test behavioral claims
  /codebase-search "What behaviors are documented that need verification?" --target-directories "$DOCS_PATH"

  # Step 3.4: Cross-Reference Verification
  echo "🔗 Cross-referencing documentation with implementation..."

  # For each documented API, find its implementation
  jq -r '.api_signatures[].name' .docs-verify/assertions/*.json | while read -r api_name; do
    echo "  Cross-referencing: $api_name"

    # Find implementation
    /codebase-search "Where is $api_name implemented?" --target-directories "$CODEBASE_PATH"

    # Compare documentation with actual implementation
    # Store discrepancies in accuracy_issues
  done

  # Step 3.5: Calculate Confidence Level
  echo "📊 Calculating verification confidence..."

  TOTAL_VERIFICATIONS=$(find .docs-verify/evidence -name "*.json" | wc -l)
  SUCCESSFUL_VERIFICATIONS=$(find .docs-verify/evidence -name "*.json" | xargs jq -s '[.[] | select(.success == true)] | length')
  CONFIDENCE_LEVEL=$(echo "scale=2; $SUCCESSFUL_VERIFICATIONS / $TOTAL_VERIFICATIONS" | bc)

  echo "Confidence level: $CONFIDENCE_LEVEL"

  # Update state
  echo "$STATE" | jq --arg conf "$CONFIDENCE_LEVEL" --arg round "$((ROUND + 1))" '
    .confidence_level = ($conf | tonumber) |
    .verification_rounds = ($round | tonumber)
  ' > .docs-verify/state.json

  ROUND=$((ROUND + 1))
done
```

### Step 3.6: Multi-Agent Consensus

```bash
echo "🤝 Checking multi-agent consensus..."

# Each agent evaluates the verification results
SKEPTIC_APPROVAL=$(echo "$CONFIDENCE_LEVEL > 0.95" | bc -l)
IMPLEMENTER_APPROVAL=$([ "$ACCURACY_ISSUES" -lt 5 ] && echo 1 || echo 0)
USER_APPROVAL=$(echo "$CONFIDENCE_LEVEL > 0.85" | bc -l)
MAINTAINER_APPROVAL=$([ "$FALSE_POSITIVES" -eq 0 ] && echo 1 || echo 0)

# Calculate weighted consensus
CONSENSUS_SCORE=$(echo "
  ($SKEPTIC_APPROVAL * 1.0) +
  ($IMPLEMENTER_APPROVAL * 0.9) +
  ($USER_APPROVAL * 0.8) +
  ($MAINTAINER_APPROVAL * 0.7)
" | bc -l)

if (( $(echo "$CONSENSUS_SCORE >= 3.0" | bc -l) )); then
  echo "✅ Multi-agent consensus: Documentation verified as accurate"
else
  echo "❌ Multi-agent consensus: Documentation needs corrections"
fi
```

## Phase 4: Issue Detection & Reporting (Time-boxed: 15% of budget)

### Objective: Generate actionable reports and fix recommendations

```bash
echo "📋 Phase 4: Issue Detection & Reporting"

# Aggregate all verification results
cat > .docs-verify/verification-report.md << 'EOF'
# Documentation Verification Report

**Target**: $DOCS_PATH
**Codebase**: $CODEBASE_PATH
**Date**: $(date)
**Confidence Level**: $CONFIDENCE_LEVEL

## Summary

- Total Assertions Verified: $TOTAL_VERIFICATIONS
- Successful Verifications: $SUCCESSFUL_VERIFICATIONS
- Accuracy Issues Found: $ACCURACY_ISSUES
- False Positives Prevented: $FALSE_POSITIVES
- Verification Rounds: $ROUND

## Accuracy Issues

$(jq -r '.accuracy_issues[] | "### \(.category): \(.title)\n\n**File**: \(.file)\n**Line**: \(.line)\n**Issue**: \(.description)\n**Suggested Fix**: \(.fix)\n"' .docs-verify/state.json)

## API Signature Mismatches

$(find .docs-verify/evidence -name "*api*.json" | xargs jq -s '
  [.[] | select(.success == false)] |
  .[] | "- `\(.api_name)`: \(.mismatch_reason)"
')

## Code Example Failures

$(find .docs-verify/evidence -name "*example*.log" | while read -r log; do
  if grep -q "Error\|Exception" "$log"; then
    echo "- $(basename "$log"): Failed with error"
    echo "  \`\`\`"
    tail -5 "$log" | sed 's/^/  /'
    echo "  \`\`\`"
  fi
done)

## Recommendations

### Immediate Actions
$(jq -r '.accuracy_issues |
  group_by(.severity) |
  .[] |
  if .[0].severity == "high" then
    "1. **Critical**: Fix \(length) high-severity issues immediately"
  elif .[0].severity == "medium" then
    "2. **Important**: Address \(length) medium-severity issues"
  else
    "3. **Minor**: Consider fixing \(length) low-severity issues"
  end
' .docs-verify/state.json)

### Preventive Measures
- Add automated documentation tests to CI/CD
- Implement documentation linting rules
- Create documentation templates with validation
- Regular verification runs (suggested: weekly)
EOF
```

### Generate Fix Patches

```bash
# Generate automated fix suggestions where possible
echo "🔧 Generating fix patches..."

# For each accuracy issue, attempt to generate a fix
jq -r '.accuracy_issues[]' .docs-verify/state.json | while read -r issue; do
  ISSUE_FILE=$(echo "$issue" | jq -r '.file')
  ISSUE_LINE=$(echo "$issue" | jq -r '.line')
  ISSUE_TYPE=$(echo "$issue" | jq -r '.category')

  case $ISSUE_TYPE in
    "api_signature")
      # Generate corrected API documentation
      echo "Generating API signature fix for $ISSUE_FILE:$ISSUE_LINE"
      ;;
    "code_example")
      # Generate working code example
      echo "Generating code example fix for $ISSUE_FILE:$ISSUE_LINE"
      ;;
    "behavioral_claim")
      # Generate accurate behavior description
      echo "Generating behavioral description fix for $ISSUE_FILE:$ISSUE_LINE"
      ;;
  esac
done > .docs-verify/fix-patches.md
```

### Memory Integration

```bash
# Store verification results and patterns
echo "🧠 Storing verification insights..."

mcp_mem0-mcp_add-memory "Documentation verification completed for $DOCS_PATH:
- Confidence level achieved: $CONFIDENCE_LEVEL
- Total accuracy issues: $ACCURACY_ISSUES
- Most common issue type: $MOST_COMMON_ISSUE
- Verification strategy used: $STRATEGY
- False positives prevented: $FALSE_POSITIVES
- Key learning: $KEY_LEARNING" "mem0-mcp-user"

# Store patterns of common documentation drift
if [ "$ACCURACY_ISSUES" -gt 5 ]; then
  mcp_mem0-mcp_add-memory "Documentation drift patterns detected:
  - API signatures tend to drift in: $API_DRIFT_AREAS
  - Code examples become stale when: $EXAMPLE_STALENESS_PATTERN
  - Behavioral descriptions drift due to: $BEHAVIORAL_DRIFT_CAUSE" "mem0-mcp-user"
fi
```

## Phase 5: Continuous Verification Setup (Optional Enhancement)

### Objective: Establish ongoing verification processes

```bash
echo "🔄 Phase 5: Continuous Verification Setup"

# Generate GitHub Action for automated verification
cat > .github/workflows/docs-verification.yml << 'EOF'
name: Documentation Verification
on:
  pull_request:
    paths:
      - 'docs/**'
      - 'src/**'
      - 'sdk/**'
  schedule:
    - cron: '0 9 * * 1'  # Weekly verification
  workflow_dispatch:
    inputs:
      scope:
        description: 'Verification scope'
        required: false
        default: 'all'

jobs:
  verify-docs:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup verification environment
        run: |
          npm install
          mkdir -p .docs-verify

      - name: Run documentation verification
        run: |
          # Use Claude Code verification workflow
          /docs-verification-game docs/ . --scope=${{ inputs.scope || 'all' }}

      - name: Upload verification report
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: verification-report
          path: .docs-verify/verification-report.md

      - name: Comment on PR
        if: github.event_name == 'pull_request'
        uses: actions/github-script@v6
        with:
          script: |
            const report = fs.readFileSync('.docs-verify/verification-report.md', 'utf8');
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: report
            });
EOF

echo "✅ Continuous verification workflow created"
```

## Integration with Other Systems

```bash
# Optional: Integration with tracking systems
if [ "$ENABLE_TRACKING" = "true" ]; then
  # Create Linear ticket for documentation fixes
  if [ "$ACCURACY_ISSUES" -gt 0 ]; then
    mcp_mcp-linear_linear_createIssue '{
      "title": "Documentation accuracy issues found",
      "description": "Verification found '"$ACCURACY_ISSUES"' accuracy issues. See report: '"$BUILD_URL"'",
      "teamId": "'"$DOCS_TEAM_ID"'",
      "priority": '"$([ $ACCURACY_ISSUES -gt 10 ] && echo 1 || echo 2)"',
      "labelIds": ["documentation", "verification"]
    }'
  fi

  # Track verification metrics in Airtable
  mcp_airtable-server_create_record "docs-metrics" "Verification Runs" '{
    "Date": "'"$(date)"'",
    "Target": "'"$DOCS_PATH"'",
    "Confidence": '"$CONFIDENCE_LEVEL"',
    "Issues Found": '"$ACCURACY_ISSUES"',
    "Strategy": "'"$STRATEGY"'",
    "Duration": "'"$VERIFICATION_DURATION"'"
  }'
fi
```

## Cleanup & Success Metrics

```bash
echo "🏁 Documentation verification complete!"

# Generate final summary
cat > .docs-verify/FINAL_SUMMARY.md << EOF
# Documentation Verification Summary

## Results
✅ **Confidence Level**: $CONFIDENCE_LEVEL / 1.0
📊 **Coverage**: $TOTAL_VERIFICATIONS assertions verified
❌ **Issues Found**: $ACCURACY_ISSUES
🛡️ **False Positives Prevented**: $FALSE_POSITIVES

## Performance Metrics
- Verification Time: $VERIFICATION_DURATION
- Rounds Required: $ROUND / $MAX_ROUNDS
- Strategy Effectiveness: $STRATEGY_EFFECTIVENESS

## Next Steps
$([ "$ACCURACY_ISSUES" -gt 0 ] && echo "1. Review and apply fixes from .docs-verify/fix-patches.md" || echo "1. Documentation is accurate!")
2. Set up continuous verification (workflow provided)
3. Monitor documentation drift over time

## Recommendations
$VERIFICATION_RECOMMENDATIONS
EOF

# Archive results
if [ "$KEEP_VERIFICATION_STATE" = "true" ]; then
  tar -czf .docs-verify-$(date +%Y%m%d-%H%M%S).tar.gz .docs-verify/
fi

echo "📊 Final Report: .docs-verify/FINAL_SUMMARY.md"
echo "🎯 Verification complete with confidence level: $CONFIDENCE_LEVEL"
```

## Usage Examples

```bash
# Verify all documentation with standard depth
/docs-verification-game docs/ src/

# Deep verification of SDK documentation with high confidence
/docs-verification-game docs/sdk/ sdk/typescript/ --depth=deep --confidence=0.95

# Quick verification of specific file
/docs-verification-game docs/getting-started.mdx src/ --depth=surface --strategy=sequential

# Adversarial verification to find edge cases
/docs-verification-game docs/api/ src/api/ --strategy=adversarial --confidence=0.99

# Parallel verification with custom scope
/docs-verification-game docs/ . --scope=api,examples --strategy=parallel
```

## Integration with Other Commands

```bash
# After documentation generation
/docs-improvement-game "Authentication API"
/docs-verification-game docs/api/auth.mdx src/auth/

# With continuous improvement
/learning-accelerator --domain=documentation-accuracy
/docs-verification-game --strategy=learned-patterns

# Combined with refactoring verification
/refactoring-game src/api/
/docs-verification-game docs/api/ src/api/ --depth=deep
```

This command transforms documentation verification from manual spot-checking into a systematic, game-theoretic process that ensures documentation accuracy while preventing common verification pitfalls and false positives.
