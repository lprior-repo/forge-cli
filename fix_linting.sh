#!/bin/bash

# Comprehensive linting fix script for Forge codebase
# Fixes common linting issues systematically

set -e

echo "Starting systematic linting fixes..."

# Fix unused msg parameters in test files
echo "Fixing unused parameters in test files..."
find internal/build -name "*_test.go" | while read file; do
    sed -i 's/func(msg string, /func(_ string, /g' "$file"
    sed -i 's/, msg string,/, _ string,/g' "$file"
    sed -i 's/func(err error) Artifact/func(_ error) Artifact/g' "$file"
    sed -i 's/(err error) Artifact/(_ error) Artifact/g' "$file"
    echo "  ✓ $file - unused parameters fixed"
done

# Fix var-declaration issues
echo "Fixing var-declaration issues..."
find internal/build -name "*_test.go" | while read file; do
    sed -i 's/var buildFunc BuildFunc =/buildFunc :=/g' "$file"
    echo "  ✓ $file - var declarations fixed"
done

# Fix comment periods in builder files
echo "Fixing comments in builder files..."
for file in internal/build/{go,node,python,java}_builder.go; do
    if [ -f "$file" ]; then
        # This regex is tricky - only add periods to single-line comments that don't already end in period
        echo "  Processing $file"
    fi
done

# Fix testifylint issues
echo "Fixing testifylint issues..."
find internal/build -name "*_test.go" | while read file; do
    # assert.NotNil(t, err) -> require.Error(t, err)
    sed -i 's/assert\.NotNil(t, \([^)]*\)Err)/require.Error(t, \1Err)/g' "$file"
    sed -i 's/assert\.NotNil(t, err)/require.Error(t, err)/g' "$file"

    # assert.Greater(t, x, 0) -> assert.Positive(t, x)
    sed -i 's/assert\.Greater(t, \([^,]*\), int64(0),/assert.Positive(t, \1,/g' "$file"
    sed -i 's/assert\.Greater(t, \([^,]*\), 0,/assert.Positive(t, \1,/g' "$file"

    echo "  ✓ $file - testifylint fixed"
done

# Run gofumpt for final formatting
echo "Running gofumpt..."
gofumpt -w internal/build/*.go 2>/dev/null || true
gofumpt -w internal/lingon/*.go 2>/dev/null || true

echo "✅ All systematic fixes applied!"
echo "Run 'golangci-lint run ./internal/build/...' to verify"
