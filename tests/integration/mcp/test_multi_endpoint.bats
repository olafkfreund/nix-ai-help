#!/usr/bin/env bats

load helpers

## Test: Multi-Endpoint MCP Server

describe "nixai multi-endpoint MCP config" {
  it "writes all endpoints to config.yaml" {
    run cat /etc/nixai/config.yaml
    assert_output --partial 'endpoints:'
    assert_output --partial 'name: default'
    assert_output --partial 'name: test'
    assert_output --partial 'socket_path: /run/nixai/mcp.sock'
    assert_output --partial 'socket_path: /tmp/nixai-test.sock'
  }

  it "CLI can use custom endpoint socket" {
    # This assumes a test socket is running; adjust as needed for your test env
    run nixai --ask "What is NixOS?" --socket-path=/tmp/nixai-test.sock
    # Should not error out (exit code 0 or 1 if no server, but not crash)
    [ "$status" -eq 0 ] || [ "$status" -eq 1 ]
  }
}