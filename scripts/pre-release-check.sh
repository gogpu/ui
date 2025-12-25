#!/usr/bin/env bash
# Pre-Release Validation Script for gogpu/ui
# This script runs all quality checks before creating a release
# EXACTLY matches CI checks + additional validations
#
# Usage:
#   bash scripts/pre-release-check.sh          # Full check before release
#   bash scripts/pre-release-check.sh --quick  # Quick check during development
#
# On Windows with multiple Go versions, set GOROOT:
#   GOROOT="/c/Program Files/Go" bash scripts/pre-release-check.sh

set -e  # Exit on first error

# Handle GOROOT for Windows with multiple Go versions
if [[ -n "$GOROOT" ]]; then
    export PATH="$GOROOT/bin:$PATH"
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Header
echo ""
echo "================================================"
echo "  gogpu/ui - Pre-Release Check"
echo "================================================"
echo ""

# Track overall status
ERRORS=0
WARNINGS=0

# 1. Check Go version
log_info "Checking Go version..."
GO_VERSION=$(go version | awk '{print $3}')
REQUIRED_VERSION="go1.25"
if [[ "$GO_VERSION" < "$REQUIRED_VERSION" ]]; then
    log_error "Go version $REQUIRED_VERSION+ required, found $GO_VERSION"
    ERRORS=$((ERRORS + 1))
else
    log_success "Go version: $GO_VERSION"
fi
echo ""

# 2. Check git status
log_info "Checking git status..."
if git diff-index --quiet HEAD --; then
    log_success "Working directory is clean"
else
    log_warning "Uncommitted changes detected"
    git status --short
    WARNINGS=$((WARNINGS + 1))
fi
echo ""

# 3. Code formatting check (EXACT CI command)
log_info "Checking code formatting (gofmt -l .)..."
UNFORMATTED=$(gofmt -l .)
if [ -n "$UNFORMATTED" ]; then
    log_error "The following files need formatting:"
    echo "$UNFORMATTED"
    echo ""
    log_info "Run: go fmt ./..."
    ERRORS=$((ERRORS + 1))
else
    log_success "All files are properly formatted"
fi
echo ""

# 4. Go vet
log_info "Running go vet..."
if go vet ./... 2>&1; then
    log_success "go vet passed"
else
    log_error "go vet failed"
    ERRORS=$((ERRORS + 1))
fi
echo ""

# 5. Build all packages
log_info "Building all packages..."
if go build ./... 2>&1; then
    log_success "Build successful"
else
    log_error "Build failed"
    ERRORS=$((ERRORS + 1))
fi
echo ""

# 5.1 Build examples
log_info "Building examples..."
if go build ./examples/... 2>&1; then
    log_success "Examples build successful"
else
    log_error "Examples build failed"
    ERRORS=$((ERRORS + 1))
fi
echo ""

# 6. go.mod validation
log_info "Validating go.mod..."
go mod verify
if [ $? -eq 0 ]; then
    log_success "go.mod verified"
else
    log_error "go.mod verification failed"
    ERRORS=$((ERRORS + 1))
fi

# Check if go.mod needs tidying
go mod tidy
if git diff --quiet go.mod go.sum 2>/dev/null; then
    log_success "go.mod is tidy"
else
    log_warning "go.mod needs tidying (run 'go mod tidy')"
    git diff go.mod go.sum 2>/dev/null || true
    WARNINGS=$((WARNINGS + 1))
fi
echo ""

# 6.5. Verify golangci-lint configuration
log_info "Verifying golangci-lint configuration..."
if command -v golangci-lint &> /dev/null; then
    if golangci-lint config verify 2>&1; then
        log_success "golangci-lint config is valid"
    else
        log_error "golangci-lint config is invalid"
        ERRORS=$((ERRORS + 1))
    fi
else
    log_warning "golangci-lint not installed (optional but recommended)"
    log_info "Install: https://golangci-lint.run/welcome/install/"
    WARNINGS=$((WARNINGS + 1))
fi
echo ""

# 7. Run tests with race detector
# Priority: 1) kolkov/racedetector (Pure Go), 2) go test -race (requires CGO), 3) no race detection
RACE_MODE="none"

if command -v racedetector &> /dev/null; then
    log_info "Using kolkov/racedetector (Pure Go, no CGO required)..."
    RACE_MODE="racedetector"
elif command -v gcc &> /dev/null || command -v clang &> /dev/null; then
    log_info "Using standard race detector (CGO enabled)..."
    RACE_MODE="standard"
else
    log_warning "No race detector available"
    log_info "Install: go install github.com/kolkov/racedetector/cmd/racedetector@latest"
    WARNINGS=$((WARNINGS + 1))
fi

log_info "Running tests..."
case "$RACE_MODE" in
    "racedetector")
        TEST_OUTPUT=$(racedetector test ./... 2>&1 || true)
        ;;
    "standard")
        TEST_OUTPUT=$(go test -race ./... 2>&1 || true)
        ;;
    *)
        TEST_OUTPUT=$(go test ./... 2>&1 || true)
        ;;
esac

if echo "$TEST_OUTPUT" | grep -q "FAIL"; then
    log_error "Tests failed"
    echo "$TEST_OUTPUT"
    ERRORS=$((ERRORS + 1))
elif echo "$TEST_OUTPUT" | grep -q "PASS\|ok"; then
    if [ "$RACE_MODE" != "none" ]; then
        log_success "All tests passed with race detector"
    else
        log_success "All tests passed (race detector not available)"
    fi
else
    log_warning "No tests found or unexpected output"
    WARNINGS=$((WARNINGS + 1))
fi
echo ""

# 8. Test coverage check
log_info "Checking test coverage..."
COVERAGE=$(go test -cover ./... 2>&1 | grep "coverage:" | tail -1 | awk '{print $5}' | sed 's/%//')
if [ -n "$COVERAGE" ]; then
    echo "  overall coverage: ${COVERAGE}%"
    if awk -v cov="$COVERAGE" 'BEGIN {exit !(cov >= 70.0)}'; then
        log_success "Coverage meets requirement (>70%)"
    else
        log_warning "Coverage below 70% (${COVERAGE}%) - acceptable for early versions"
        WARNINGS=$((WARNINGS + 1))
    fi
else
    log_warning "Could not determine coverage (no tests yet)"
    WARNINGS=$((WARNINGS + 1))
fi
echo ""

# 9. Dependency check
log_info "Checking dependencies..."

GG_VERSION=$(grep "gogpu/gg" go.mod | awk '{print $2}')
if [ -n "$GG_VERSION" ]; then
    log_success "gogpu/gg dependency: $GG_VERSION"
else
    log_warning "gogpu/gg not found (2D rendering won't work)"
    WARNINGS=$((WARNINGS + 1))
fi

GOGPU_VERSION=$(grep "gogpu/gogpu" go.mod | awk '{print $2}')
if [ -n "$GOGPU_VERSION" ]; then
    log_success "gogpu/gogpu dependency: $GOGPU_VERSION"
else
    log_warning "gogpu/gogpu not found (windowing won't work)"
    WARNINGS=$((WARNINGS + 1))
fi

SIGNALS_VERSION=$(grep "coregx/signals" go.mod | awk '{print $2}')
if [ -n "$SIGNALS_VERSION" ]; then
    log_success "coregx/signals dependency: $SIGNALS_VERSION"
else
    log_warning "coregx/signals not found (state management won't work)"
    WARNINGS=$((WARNINGS + 1))
fi
echo ""

# 10. golangci-lint (same as CI)
log_info "Running golangci-lint..."
if command -v golangci-lint &> /dev/null; then
    LINT_OUTPUT=$(golangci-lint run --timeout=5m ./... 2>&1 || true)
    # Check for "0 issues" or empty output (success)
    if echo "$LINT_OUTPUT" | grep -qE "(^0 issues|no issues)"; then
        log_success "golangci-lint passed with 0 issues"
    elif [ -z "$LINT_OUTPUT" ]; then
        log_success "golangci-lint passed"
    elif echo "$LINT_OUTPUT" | grep -q "issues:"; then
        # Has issues - extract count
        log_error "Linter found issues"
        echo "$LINT_OUTPUT" | tail -20
        ERRORS=$((ERRORS + 1))
    else
        log_success "golangci-lint passed"
    fi
else
    log_warning "golangci-lint not installed"
    log_info "Install: https://golangci-lint.run/welcome/install/"
    WARNINGS=$((WARNINGS + 1))
fi
echo ""

# 11. Check for TODO/FIXME comments
log_info "Checking for TODO/FIXME comments..."
TODO_COUNT=$(grep -r "TODO\|FIXME" --include="*.go" --exclude-dir=vendor . 2>/dev/null | wc -l)
if [ "$TODO_COUNT" -gt 0 ]; then
    log_warning "Found $TODO_COUNT TODO/FIXME comments"
    grep -r "TODO\|FIXME" --include="*.go" --exclude-dir=vendor . 2>/dev/null | head -5
    WARNINGS=$((WARNINGS + 1))
else
    log_success "No TODO/FIXME comments found"
fi
echo ""

# 12. Check critical documentation files
log_info "Checking documentation..."
DOCS_MISSING=0
REQUIRED_DOCS="README.md LICENSE"

for doc in $REQUIRED_DOCS; do
    if [ ! -f "$doc" ]; then
        log_error "Missing: $doc"
        DOCS_MISSING=1
        ERRORS=$((ERRORS + 1))
    fi
done

if [ $DOCS_MISSING -eq 0 ]; then
    log_success "All critical documentation files present"
fi
echo ""

# Summary
echo "========================================"
echo "  Summary"
echo "========================================"
echo ""

if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    log_success "All checks passed! Ready for release."
    echo ""
    log_info "Next steps for release (GitHub Flow):"
    echo ""
    echo "  1. Update CHANGELOG.md and README.md if needed"
    echo ""
    echo "  2. Commit changes:"
    echo "     git add -A"
    echo "     git commit -m \"chore: prepare vX.Y.Z release\""
    echo "     git push"
    echo ""
    echo "  3. Wait for CI to pass on main"
    echo ""
    echo "  4. Create and push tag:"
    echo "     git tag -a vX.Y.Z -m \"Release vX.Y.Z\""
    echo "     git push origin vX.Y.Z"
    echo ""
    exit 0
elif [ $ERRORS -eq 0 ]; then
    log_warning "Checks completed with $WARNINGS warning(s)"
    echo ""
    log_info "Review warnings above before proceeding with release"
    echo ""
    exit 0
else
    log_error "Checks failed with $ERRORS error(s) and $WARNINGS warning(s)"
    echo ""
    log_error "Fix errors before creating release"
    echo ""
    exit 1
fi
