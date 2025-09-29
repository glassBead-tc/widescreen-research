#!/bin/bash
# Post-Write Hook: Validate newly created files
# Checks syntax, formatting, and basic correctness immediately

set -e

TOOL_RESULT=$(cat)
FILE_PATH=$(echo "$TOOL_RESULT" | jq -r '.arguments.file_path // .arguments.filePath // empty')

[ -z "$FILE_PATH" ] && exit 0
[ ! -f "$FILE_PATH" ] && exit 0

echo "✨ Validating newly created file: $FILE_PATH..." >&2

ISSUES=""

# Go files - check if it compiles
if [[ "$FILE_PATH" =~ \.go$ ]]; then
  if ! go build "$FILE_PATH" 2>&1 >/dev/null; then
    ISSUES="${ISSUES}\n- ❌ Go syntax errors"
  fi

  # Check gofmt
  if ! gofmt -l "$FILE_PATH" | grep -q .; then
    gofmt -w "$FILE_PATH"
    ISSUES="${ISSUES}\n- ⚠️ Formatted with gofmt"
  fi
fi

# TypeScript files - check syntax
if [[ "$FILE_PATH" =~ \.(ts|tsx)$ ]]; then
  if command -v tsc &> /dev/null; then
    if ! tsc --noEmit "$FILE_PATH" 2>&1 >/dev/null; then
      ISSUES="${ISSUES}\n- ❌ TypeScript syntax errors"
    fi
  fi
fi

# Python files - check syntax
if [[ "$FILE_PATH" =~ \.py$ ]]; then
  if ! python3 -m py_compile "$FILE_PATH" 2>&1 >/dev/null; then
    ISSUES="${ISSUES}\n- ❌ Python syntax errors"
  fi
fi

# JSON files - validate JSON
if [[ "$FILE_PATH" =~ \.json$ ]]; then
  if ! jq empty "$FILE_PATH" 2>&1 >/dev/null; then
    ISSUES="${ISSUES}\n- ❌ Invalid JSON"
  fi
fi

# YAML files - validate YAML
if [[ "$FILE_PATH" =~ \.(yaml|yml)$ ]]; then
  if command -v yamllint &> /dev/null; then
    if ! yamllint "$FILE_PATH" 2>&1 >/dev/null; then
      ISSUES="${ISSUES}\n- ⚠️ YAML linting issues"
    fi
  fi
fi

# Shell scripts - check syntax and make executable
if [[ "$FILE_PATH" =~ \.sh$ ]]; then
  if ! bash -n "$FILE_PATH" 2>&1 >/dev/null; then
    ISSUES="${ISSUES}\n- ❌ Shell syntax errors"
  fi

  # Make executable
  chmod +x "$FILE_PATH"
  ISSUES="${ISSUES}\n- ✅ Made executable"
fi

# Report issues
if [ -n "$ISSUES" ]; then
  cat << EOF
{
  "append": "\n\n📝 **File Validation Results for $FILE_PATH**\n$ISSUES"
}
EOF
fi

exit 0
