#!/bin/bash
# Post-Edit Hook: Run relevant tests after code changes
# Provides immediate feedback on whether changes break tests

set -e

TOOL_RESULT=$(cat)
FILE_PATH=$(echo "$TOOL_RESULT" | jq -r '.arguments.file_path // .arguments.filePath // empty')

[ -z "$FILE_PATH" ] && exit 0

# Only test actual code files, not config/docs
case "$FILE_PATH" in
  *.go|*.js|*.ts|*.tsx|*.py)
    ;;
  *)
    exit 0
    ;;
esac

echo "🧪 Running tests for $FILE_PATH..." >&2

# Go files - run package tests
if [[ "$FILE_PATH" =~ \.go$ ]]; then
  PACKAGE=$(dirname "$FILE_PATH")

  if [ -f "$PACKAGE"/*_test.go ] 2>/dev/null; then
    if ! go test "./$PACKAGE" -v 2>&1 | tail -20; then
      cat << EOF
{
  "append": "\n\n❌ **Tests Failed for $PACKAGE**\n\nRun \`go test -v ./$PACKAGE\` to see full output.\n\n**Action Required**: Fix failing tests before proceeding."
}
EOF
    else
      cat << EOF
{
  "append": "\n✅ Tests passed for $PACKAGE"
}
EOF
    fi
  else
    cat << EOF
{
  "append": "\n⚠️ No tests found for $PACKAGE. Consider adding tests."
}
EOF
  fi
fi

# TypeScript/JavaScript - run related tests
if [[ "$FILE_PATH" =~ \.(ts|tsx|js|jsx)$ ]]; then
  # Try to find and run relevant test file
  TEST_FILE="${FILE_PATH%.*}.test.${FILE_PATH##*.}"

  if [ -f "$TEST_FILE" ]; then
    if command -v jest &> /dev/null; then
      if ! jest "$TEST_FILE" --verbose 2>&1 | tail -20; then
        cat << EOF
{
  "append": "\n\n❌ **Tests Failed for $FILE_PATH**\n\nRun \`jest $TEST_FILE\` for details."
}
EOF
      else
        cat << EOF
{
  "append": "\n✅ Tests passed for $FILE_PATH"
}
EOF
      fi
    fi
  fi
fi

# Python files - run pytest on test file if exists
if [[ "$FILE_PATH" =~ \.py$ ]]; then
  TEST_FILE="${FILE_PATH%.*}_test.py"

  if [ -f "$TEST_FILE" ]; then
    if command -v pytest &> /dev/null; then
      if ! pytest "$TEST_FILE" -v 2>&1 | tail -20; then
        cat << EOF
{
  "append": "\n\n❌ **Tests Failed for $FILE_PATH**\n\nRun \`pytest $TEST_FILE -v\` for details."
}
EOF
      else
        cat << EOF
{
  "append": "\n✅ Tests passed for $FILE_PATH"
}
EOF
      fi
    fi
  fi
fi

exit 0
