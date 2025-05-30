#!/bin/bash
#
# Run AI provider integration tests for NixAI
#

set -e

echo "🧪 Running AI Provider Tests"
echo "==========================="

# Define color codes for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Track overall status
OVERALL_STATUS=0

# First, run the all providers test
ALL_PROVIDERS_TEST="tests/providers/test-all-providers.sh"
if [ -x "$ALL_PROVIDERS_TEST" ]; then
    echo "Running comprehensive provider test..."
    if $ALL_PROVIDERS_TEST; then
        echo -e "${GREEN}✅ All providers test PASSED${NC}"
    else
        echo -e "${RED}❌ All providers test FAILED${NC}"
        OVERALL_STATUS=1
    fi
    echo ""
fi

# Run any additional provider-specific tests
for test_script in tests/providers/provider_*.{py,sh}; do
    if [ -x "$test_script" ]; then
        echo "Running $test_script..."
        if $test_script; then
            echo -e "${GREEN}✅ $test_script PASSED${NC}"
        else
            echo -e "${RED}❌ $test_script FAILED${NC}"
            OVERALL_STATUS=1
        fi
        echo ""
    fi
done

echo "==========================="
if [ $OVERALL_STATUS -eq 0 ]; then
    echo -e "${GREEN}✅ ALL PROVIDER TESTS PASSED${NC}"
else
    echo -e "${RED}❌ SOME PROVIDER TESTS FAILED${NC}"
fi

exit $OVERALL_STATUS
