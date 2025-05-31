#!/bin/bash
#
# Test script for all AI providers integration
# Part of the NixAI test suite
#

# Determine path to nixai executable (handles running from any directory)
REPO_ROOT=$(git rev-parse --show-toplevel 2>/dev/null || echo "$HOME/Source/NIX/nix-ai-help")
NIXAI_BIN="$REPO_ROOT/nixai"
TEST_LOGS_DIR="$REPO_ROOT/tests/providers/logs"

# Create logs directory if it doesn't exist
mkdir -p "$TEST_LOGS_DIR"

echo "🧪 TESTING ALL AI PROVIDERS"
echo "==========================="

# Test Ollama
echo "📋 1. Testing Ollama provider..."
echo "set ai ollama llama3" | $NIXAI_BIN interactive > /dev/null 2>&1
echo "  - Switched to Ollama provider"
$NIXAI_BIN explain-option services.openssh.enable > "$TEST_LOGS_DIR/ollama_test.log" 2>&1
if grep -q "Complete!" "$TEST_LOGS_DIR/ollama_test.log"; then
    echo "  ✅ Ollama test PASSED"
else
    echo "  ❌ Ollama test FAILED"
fi

echo ""

# Test Gemini
echo "📋 2. Testing Gemini provider..."
echo "set ai gemini" | $NIXAI_BIN interactive > /dev/null 2>&1
echo "  - Switched to Gemini provider"
$NIXAI_BIN explain-option services.openssh.enable > "$TEST_LOGS_DIR/gemini_test.log" 2>&1
if grep -q "Complete!" "$TEST_LOGS_DIR/gemini_test.log"; then
    echo "  ✅ Gemini test PASSED"
else
    echo "  ❌ Gemini test FAILED"
fi

echo ""

# Test OpenAI
echo "📋 3. Testing OpenAI provider..."
echo "set ai openai" | $NIXAI_BIN interactive > /dev/null 2>&1
echo "  - Switched to OpenAI provider"
$NIXAI_BIN explain-option services.openssh.enable > "$TEST_LOGS_DIR/openai_test.log" 2>&1
if grep -q "Complete!" "$TEST_LOGS_DIR/openai_test.log"; then
    echo "  ✅ OpenAI test PASSED"
else
    echo "  ❌ OpenAI test FAILED"
fi

echo ""
echo "=== TEST SUMMARY ==="
echo "Ollama: $(grep -q "Complete!" "$TEST_LOGS_DIR/ollama_test.log" && echo "✅ PASSED" || echo "❌ FAILED")"
echo "Gemini: $(grep -q "Complete!" "$TEST_LOGS_DIR/gemini_test.log" && echo "✅ PASSED" || echo "❌ FAILED")"
echo "OpenAI: $(grep -q "Complete!" "$TEST_LOGS_DIR/openai_test.log" && echo "✅ PASSED" || echo "❌ FAILED")"

# Calculate overall status
OLLAMA_STATUS=$(grep -q "Complete!" "$TEST_LOGS_DIR/ollama_test.log" && echo "0" || echo "1")
GEMINI_STATUS=$(grep -q "Complete!" "$TEST_LOGS_DIR/gemini_test.log" && echo "0" || echo "1")
OPENAI_STATUS=$(grep -q "Complete!" "$TEST_LOGS_DIR/openai_test.log" && echo "0" || echo "1")
OVERALL_STATUS=$((OLLAMA_STATUS + GEMINI_STATUS + OPENAI_STATUS))

echo ""
echo "Test logs saved in: $TEST_LOGS_DIR/"
echo "MCP server status: $(curl -s http://localhost:8081/healthz || echo "NOT RUNNING")"

# Exit with appropriate status
exit $((OVERALL_STATUS > 0))
