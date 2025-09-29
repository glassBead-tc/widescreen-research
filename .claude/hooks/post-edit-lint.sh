#!/bin/bash
# Post-Edit Hook: Run linter immediately after file changes
# Provides fast feedback to agent on code quality

set -e

# Read tool result from stdin
TOOL_RESULT=$(cat)
FILE_PATH=$(echo "$TOOL_RESULT" | jq -r '.arguments.file_path // .arguments.filePath // empty')

# Exit if no file path or not a code file
if [ -z "$FILE_PATH" ]; then
  exit 0
fi

# Only lint code files
case "$FILE_PATH" in
  *.go|*.js|*.ts|*.tsx|*.py|*.sh)
    ;;
  *)
    exit 0
    ;;
esac

echo "🔍 Running linter on $FILE_PATH..." >&2

# Go files
if [[ "$FILE_PATH" =~ \.go$ ]]; then
  if command -v golangci-lint &> /dev/null; then
    if ! golangci-lint run --fix "$FILE_PATH" 2>&1; then
      cat << EOF
{
  "append": "\n\n⚠️ **Linting Issues Detected in $FILE_PATH**\n\nRun \`golangci-lint run --fix $FILE_PATH\` to see details.\n\nCommon issues:\n- Unused imports\n- Unused variables\n- Missing error checks\n- Formatting issues"
}
EOF
    else
      cat << EOF
{
  "append": "\n✅ Linting passed for $FILE_PATH"
}
EOF
    fi
  fi
fi

# TypeScript/JavaScript files
if [[ "$FILE_PATH" =~ \.(ts|tsx|js|jsx)$ ]]; then
  if command -v eslint &> /dev/null; then
    if ! eslint --fix "$FILE_PATH" 2>&1; then
      cat << EOF
{
  "append": "\n\n⚠️ **Linting Issues Detected in $FILE_PATH**\n\nRun \`eslint $FILE_PATH\` for details."
}
EOF
    else
      cat << EOF
{
  "append": "\n✅ Linting passed for $FILE_PATH"
}
EOF
    fi
  fi
fi

# Python files
if [[ "$FILE_PATH" =~ \.py$ ]]; then
  if command -v ruff &> /dev/null; then
    if ! ruff check --fix "$FILE_PATH" 2>&1; then
      cat << EOF
{
  "append": "\n\n⚠️ **Linting Issues Detected in $FILE_PATH**\n\nRun \`ruff check $FILE_PATH\` for details."
}
EOF
    else
      cat << EOF
{
  "append": "\n✅ Linting passed for $FILE_PATH"
}
EOF
    fi
  fi
fi

# Shell scripts
if [[ "$FILE_PATH" =~ \.sh$ ]]; then
  if command -v shellcheck &> /dev/null; then
    if ! shellcheck "$FILE_PATH" 2>&1; then
      cat << EOF
{
  "append": "\n\n⚠️ **ShellCheck Issues in $FILE_PATH**\n\nRun \`shellcheck $FILE_PATH\` for details."
}
EOF
    else
      cat << EOF
{
  "append": "\n✅ ShellCheck passed for $FILE_PATH"
}
EOF
    fi
  fi
fi

exit 0
