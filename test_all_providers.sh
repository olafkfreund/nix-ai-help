#!/bin/bash

echo "=== TESTING ALL AI PROVIDERS ==="
echo ""

# Test Ollama
echo "1. Testing Ollama provider..."
echo "set ai ollama llama3" | ./nixai interactive > /dev/null 2>&1
echo "  - Switched to Ollama provider"
./nixai explain-option services.openssh.enable > ollama_test.log 2>&1
if grep -q "Complete!" ollama_test.log; then
    echo "  ✅ Ollama test PASSED"
else
    echo "  ❌ Ollama test FAILED"
fi

echo ""

# Test Gemini
echo "2. Testing Gemini provider..."
echo "set ai gemini" | ./nixai interactive > /dev/null 2>&1
echo "  - Switched to Gemini provider"
./nixai explain-option services.openssh.enable > gemini_test.log 2>&1
if grep -q "Complete!" gemini_test.log; then
    echo "  ✅ Gemini test PASSED"
else
    echo "  ❌ Gemini test FAILED"
fi

echo ""

# Test OpenAI
echo "3. Testing OpenAI provider..."
echo "set ai openai" | ./nixai interactive > /dev/null 2>&1
echo "  - Switched to OpenAI provider"
./nixai explain-option services.openssh.enable > openai_test.log 2>&1
if grep -q "Complete!" openai_test.log; then
    echo "  ✅ OpenAI test PASSED"
else
    echo "  ❌ OpenAI test FAILED"
fi

echo ""
echo "=== TEST SUMMARY ==="
echo "Ollama: $(grep -q "Complete!" ollama_test.log && echo "✅ PASSED" || echo "❌ FAILED")"
echo "Gemini: $(grep -q "Complete!" gemini_test.log && echo "✅ PASSED" || echo "❌ FAILED")"
echo "OpenAI: $(grep -q "Complete!" openai_test.log && echo "✅ PASSED" || echo "❌ FAILED")"

echo ""
echo "Test logs saved as: ollama_test.log, gemini_test.log, openai_test.log"
echo "MCP server status: $(curl -s http://localhost:8081/healthz || echo "NOT RUNNING")"
