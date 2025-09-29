# Sandbox Testing Mechanism

Test newly generated code inside an E2B sandbox (or container-use) by building a plausible application or demonstration.  This Stage-Gate workflow forces the agent to compile, test and use its own code in a safe, disposable environment and exposes failures early.

## Overview

- **Goal** - Verify that freshly written documentation, SDK code or OSS contribution actually works by building something with it in a sandbox.
- **Pattern** - Stage-Gate phases with explicit go/kill gates and anti-pattern detection. Use E2B Sandboxes and the Model Context Protocol (MCP) server to create a virtual desktop, run commands, and capture results. E2B sandboxes let AI agents execute code in a secure isolated environment and provide actions such as create_sandbox, execute_computer_action, get_screenshot and cleanup_sandbox.
- **Scope** - Any codebase (TypeScript, Python, etc.). The workflow is additive: it does not alter the code under test and cleans up resources when done.

## Usage

```bash
/sandbox-testing-game $CODE_SOURCE [build_command] [test_command] [example_goal] [timeout]
```

## Arguments

- **$CODE_SOURCE** (required): Path or URL to the code to test.  Accepts a Git repository URL, local directory/file, or inline snippet.  For Git URLs the code is cloned; for local paths it is copied; for inline snippets it is written to a file.
- **build_command** (optional): Shell command to build or install the code.  If omitted the workflow infers a default (e.g. npm install for Node projects or pip install . for Python).
- **test_command** (optional): Shell command to run automated tests.  Defaults are inferred (e.g. npm test, pytest -q).
- **example_goal** (optional): Plain-text description of a simple but real example that exercises the code (e.g. “build a CLI that lists items” or “display a chart”).  If omitted the workflow uses the project’s quickstart or README.
- **timeout** (optional): Maximum sandbox lifetime in milliseconds.  Default: 600000 (10 minutes).

## Variables

Before starting, fill these environment variables or allow the workflow to infer them:

- **CODE_SOURCE** - input code specification (string or URL).
- **BUILD_COMMAND** - command to build/install the code.  Inferred if empty.
- **TEST_COMMAND** - command to run tests.  Inferred if empty.
- **EXAMPLE_GOAL** - description of the demonstration to build.
- **TIMEOUT** - sandbox lifetime in milliseconds (default 600000).
- **SANDBOX_ID** - unique ID of the created E2B sandbox.
- **RESULTS_DIR** - local directory .sandbox-testing-game for state, logs and artifacts.

## Algorithm

### Phase 0 - Initialize Testing Space

Create a working directory and record the initial state.  Persist the parameters for traceability and set conservative budgets for time and failures.

```bash
mkdir -p .sandbox-testing-game/{logs,artifacts,examples}
cat > .sandbox-testing-game/state.json << EOF
{
  "code_source": "$CODE_SOURCE",
  "build_command": "${BUILD_COMMAND:-""}",
  "test_command": "${TEST_COMMAND:-""}",
  "example_goal": "${EXAMPLE_GOAL:-""}",
  "timeout": ${TIMEOUT:-600000},
  "phase": 0,
  "created": "$(date -Iseconds)"
}
EOF
echo "🧪 Sandbox Testing Game initialized for code: $CODE_SOURCE"
```

### Phase 1 - Fetch & Prepare Code

Download or copy the specified code into the working directory and attempt to infer default build and test commands based on project structure (e.g. presence of package.json or setup.py).  If the code comes as a multi-line snippet, save it as main.txt.  Update state.json with inferred commands.

```bash
mkdir -p .sandbox-testing-game/code
# fetch or copy code (simplified)
if [[ "$CODE_SOURCE" =~ ^https:// ]] && [[ "$CODE_SOURCE" =~ \.git$ ]]; then
  git clone "$CODE_SOURCE" .sandbox-testing-game/code
elif [[ -d "$CODE_SOURCE" ]]; then
  cp -r "$CODE_SOURCE"/. .sandbox-testing-game/code
elif [[ "$CODE_SOURCE" == *$'\n'* ]]; then
  echo "$CODE_SOURCE" > .sandbox-testing-game/code/main.txt
else
  echo "🚫 Unknown code source type" && exit 1
fi

# detect language heuristics
DEFAULT_BUILD=""
DEFAULT_TEST=""
if [ -f .sandbox-testing-game/code/package.json ]; then
  DEFAULT_BUILD="npm install"
  DEFAULT_TEST="npm test"
elif [ -f .sandbox-testing-game/code/setup.py ] || [ -f .sandbox-testing-game/code/pyproject.toml ]; then
  DEFAULT_BUILD="pip install ."
  DEFAULT_TEST="pytest -q"
fi

# persist default commands if none provided
jq --arg build "${BUILD_COMMAND:-$DEFAULT_BUILD}" \
   --arg test "${TEST_COMMAND:-$DEFAULT_TEST}" \
   '.build_command = $build | .test_command = $test' \
   .sandbox-testing-game/state.json > tmp && mv tmp .sandbox-testing-game/state.json

echo "📦 Code prepared; default build='$DEFAULT_BUILD', test='$DEFAULT_TEST'"
```

### Phase 2 - Create E2B Sandbox

Request a new sandbox from the MCP server using the create_sandbox tool.  Specify resolution and timeout.  The server returns a sandbox ID and a streaming URL.  Save these in state.json.

```bash
SANDBOX_RES="[1280, 720]"
SANDBOX_TIMEOUT=$(jq -r .timeout .sandbox-testing-game/state.json)

# pseudo-call to create_sandbox (replace with actual MCP client)
response=$(mcp_call create_sandbox '{"resolution": [1280, 720], "timeout": '"$SANDBOX_TIMEOUT"'}')
SANDBOX_ID=$(echo "$response" | jq -r '.sandboxId')
STREAM_URL=$(echo "$response" | jq -r '.streamUrl')

jq --arg id "$SANDBOX_ID" --arg url "$STREAM_URL" '.sandbox_id = $id | .stream_url = $url' \
   .sandbox-testing-game/state.json > tmp && mv tmp .sandbox-testing-game/state.json

echo "🖥️  Sandbox created: $SANDBOX_ID"
echo "🔗 Stream URL: $STREAM_URL"
```

### Phase 3 - Upload & Build Code in Sandbox

Compress the prepared code and upload it into the sandbox.  Use the execute_computer_action tool to open a terminal inside the sandbox and run build commands.  E2B provides various actions—keypress to send keys, type to send text and double_click, drag, etc.—that allow automation.

```bash
tar -czf code.tar.gz -C .sandbox-testing-game code

# pseudo-upload (actual API may differ)
mcp_call upload_file '{"sandboxId":"'$SANDBOX_ID'","fileName":"code.tar.gz","fileData": <base64-data>}'

# open terminal via keyboard shortcut
mcp_call execute_computer_action '{"sandboxId":"'$SANDBOX_ID'","action":{"type":"keypress","keys":"Ctrl+Alt+t"}}'

# untar and build
for cmd in "mkdir -p ~/code && tar -xzvf code.tar.gz -C ~/code" "$(jq -r .build_command .sandbox-testing-game/state.json)"; do
  mcp_call execute_computer_action '{"sandboxId":"'$SANDBOX_ID'","action":{"type":"type","text":"'$cmd'"}}'
  mcp_call execute_computer_action '{"sandboxId":"'$SANDBOX_ID'","action":{"type":"keypress","keys":"Return"}}'
  # Consider waiting or capturing output between commands
done
echo "🔨 Build command executed in sandbox"
```

### Phase 4 - Run Tests & Example Goal

Execute the test command using the same action loop.  If an example goal is provided, script a minimal demonstration using the code—for instance, compile and run a sample CLI, call an API, or render a chart.  Capture a screenshot of the final state via get_screenshot.

```bash
TEST_CMD=$(jq -r .test_command .sandbox-testing-game/state.json)
if [ -n "$TEST_CMD" ]; then
  mcp_call execute_computer_action '{"sandboxId":"'$SANDBOX_ID'","action":{"type":"type","text":"'$TEST_CMD'"}}'
  mcp_call execute_computer_action '{"sandboxId":"'$SANDBOX_ID'","action":{"type":"keypress","keys":"Return"}}'
  echo "🧪 Tests started"
fi

if [ -n "$EXAMPLE_GOAL" ]; then
  echo "🎯 Executing example goal: $EXAMPLE_GOAL"
  # Implement domain-specific commands or scripts here
  # For example: write a quick script that imports the library and performs the described task
fi

# capture screenshot
shot=$(mcp_call get_screenshot '{"sandboxId":"'$SANDBOX_ID'"}')
echo "$shot" | jq -r '.screenshot' | base64 -d > .sandbox-testing-game/artifacts/final.png
echo "📸 Screenshot saved to artifacts/final.png"
```

### Phase 5 - Analyse Results & Collect Artifacts

Parse build and test logs (captured separately) to determine success/failure.  Persist results in state.json and store generated artifacts, coverage reports and the screenshot.

```bash
BUILD_STATUS="success"  # set based on log parsing
TEST_STATUS="success"   # set based on log parsing

jq --arg build "$BUILD_STATUS" --arg test "$TEST_STATUS" '.results = {build: $build, tests: $test}' \
   .sandbox-testing-game/state.json > tmp && mv tmp .sandbox-testing-game/state.json

echo "✅ Results recorded"
```

### Phase 6 - Cleanup Sandbox

Destroy the sandbox to free resources using cleanup_sandbox.

```bash
mcp_call cleanup_sandbox '{"sandboxId":"'$SANDBOX_ID'"}'
echo "🧹 Sandbox cleaned up"
```

### Phase 7 - Anti-Pattern Detection

Check for common failure modes such as missing tests, build errors or unspecified example goals.  Record anti-patterns in anti_patterns.json for later analysis.

```bash
ANTIPATTERNS=()
if [ "$BUILD_STATUS" != "success" ]; then ANTIPATTERNS+=("build_failed"); fi
if [ "$TEST_STATUS" != "success" ]; then ANTIPATTERNS+=("tests_failed"); fi
if [ -z "$EXAMPLE_GOAL" ]; then ANTIPATTERNS+=("no_example_goal"); fi

jq -n --argjson list "$(printf '%s\n' "${ANTIPATTERNS[@]}" | jq -R . | jq -s .)" '{anti_patterns: $list}' > .sandbox-testing-game/artifacts/anti_patterns.json
echo "🚨 Anti-patterns detected: ${ANTIPATTERNS[*]}"
```

Context Engineering Tips for $CODE_SOURCE
 • Provide a complete repository or snippet.  Make sure dependencies are declared in package.json, requirements.txt or pyproject.toml.
 • Include a README with clear build and test instructions.  When omitted, the workflow infers commands but may guess incorrectly.
 • Keep the example goal simple yet realistic.  For example, for an SDK commit, write a script that imports the SDK and calls a basic function.
 • Avoid interactive prompts (pass -y flags) since the terminal inside the sandbox is non-interactive.
 • Use environment variables for any secrets or API keys.  Do not hardcode them in code or build commands.

# Test a Python SDK by building a CLI that lists items from the SDK

/sandbox-testing-game "https://github.com/acme/my-python-sdk.git" "pip install ." "pytest -q" "build a CLI that lists items" 900000

# Quick test of a Node library with default commands and a simple example

/sandbox-testing-game "https://github.com/example/user-lib.git" "" "" "run the sample program in README" 600000

Anti-Patterns Prevented

 1. Unverified Code - Forces building and testing within a fresh sandbox before claiming success.
 2. No Example - Encourages creation of a plausible application to exercise the code.
 3. Missing Tests - Runs tests explicitly; flags when none are provided.
 4. Resource Leaks - Ensures sandboxes are destroyed after use.
 5. Ambiguous Build Commands - Attempts to infer build and test commands but records them for transparency.

Integration Notes (E2B MCP Context)
 • The workflow assumes an E2B API key is configured.  E2B sandboxes allow AI agents to execute code in a safe environment and expose tools for creating and managing virtual desktops.  Once a sandbox is created, you can control it using execute_computer_action to click, type, keypress, drag and scroll.
 • Use get_screenshot to capture the desktop state as a PNG.  Use cleanup_sandbox to destroy the sandbox when finished.
 • Integrate this workflow with an orchestrator or mcp-use to programmatically call MCP tools.  See E2B’s docs for client libraries.

Success Metrics
 • The build completes without errors.
 • All tests pass.  Any failures are logged and surfaced to the agent.
 • The example goal executes successfully, demonstrating the code’s intended use.
 • The sandbox is cleaned up within the specified timeout.
 • The anti-pattern report is empty or contains only acceptable warnings.
