#!/bin/bash

set -e

echo "=========================================="
echo "etlmon Verification Script"
echo "=========================================="
echo ""

# Build verification
echo "Step 1: Building project..."
go build ./...
echo "✓ Build successful"
echo ""

# Run all tests with verbose output
echo "Step 2: Running tests (verbose)..."
go test ./... -v
echo "✓ All tests passed"
echo ""

# Show coverage summary
echo "Step 3: Running tests with coverage summary..."
go test ./... -cover
echo "✓ Coverage summary complete"
echo ""

# Generate coverage profile
echo "Step 4: Generating coverage profile..."
go test ./... -coverprofile=coverage.out
echo "✓ Coverage profile generated"
echo ""

# Generate HTML coverage report
echo "Step 5: Generating HTML coverage report..."
go tool cover -html=coverage.out -o coverage.html
echo "✓ HTML coverage report generated: coverage.html"
echo ""

# Race detector
echo "Step 6: Running race detector..."
go test -race ./internal/tui/...
echo "✓ Race detector passed"
echo ""

# Static analysis
echo "Step 7: Running go vet..."
go vet ./...
echo "✓ Static analysis passed"
echo ""

echo "=========================================="
echo "✓ All verification checks passed!"
echo "=========================================="
echo ""
echo "Coverage report available at: coverage.html"
