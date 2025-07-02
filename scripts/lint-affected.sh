#!/bin/bash
set -e

# Move to the repository root directory
cd "$(git rev-parse --show-toplevel)"

# Get staged Go files
CHANGED_GO_FILES=$(git diff --name-only --cached --diff-filter=ACMRT | grep -E '\.go$' || true)

if [ -z "$CHANGED_GO_FILES" ]; then
    echo "No staged Go files detected."
    exit 0
fi

# Run golangci-lint on staged files
echo "Linting staged Go files:"
echo "$CHANGED_GO_FILES"
echo "---------------------"

# Run golangci-lint with each file explicitly specified
# The --no-config flag ensures we don't use any existing config inadvertently
if ! echo "$CHANGED_GO_FILES" | xargs -n1 -r ~/golangci-lint/golangci-lint run --fix; then
    echo "Linting failed! Fix the issues before committing."
    exit 1
fi

# If linting made any changes, add them back to staging
for file in $CHANGED_GO_FILES; do
    if git diff --quiet -- "$file"; then
        echo "No lint fixes for $file"
    else
        echo "Adding lint fixes for $file"
        git add "$file"
    fi
done

echo "Linting completed successfully!"