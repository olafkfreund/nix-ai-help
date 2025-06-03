package nixos

import (
	"errors"
	"regexp"
	"sort"
	"strings"
)

// LogEntry represents a single log entry with structured fields.
type LogEntry struct {
	Timestamp string
	Level     string
	Unit      string
	Message   string
}

// ParseLog parses log output into structured entries.
//
// Features:
// - Extracts timestamp, level, unit, and message fields from common log formats:
//   - systemd/journalctl: "Jun  3 12:34:56 host unit[PID]: message"
//   - generic:           "[timestamp] LEVEL unit: message"
//   - simple:            "LEVEL message"
//
// - Groups multi-line log entries (indented lines are appended to previous message)
// - Falls back to treating unknown lines as message-only entries
// - Easily extensible for new log formats by adding regexes
//
// Returns a slice of LogEntry for downstream diagnostics and analysis.
func ParseLog(log string) ([]LogEntry, error) {
	if log == "" {
		return nil, errors.New("log input is empty")
	}

	var entries []LogEntry
	lines := strings.Split(log, "\n")

	// Improved systemd/journalctl regex: allow flexible whitespace and host/unit fields
	systemdRe := regexp.MustCompile(`^(\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2})\s+([^ ]+)\s+([^\[]+)(?:\[\d+\])?:\s+(.*)$`)
	// Generic: "[timestamp] [level] [unit]: message"
	genericRe := regexp.MustCompile(`^\[([^\]]+)\]\s+([A-Z]+)\s+([^:]+):\s+(.*)$`)
	// Fallback: LEVEL message (e.g., 'INFO something happened')
	simpleLevelRe := regexp.MustCompile(`^([A-Z]+)\s+(.*)$`)

	var current LogEntry
	inMultiline := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if m := systemdRe.FindStringSubmatch(line); m != nil {
			if inMultiline {
				entries = append(entries, current)
				inMultiline = false
			}
			current = LogEntry{
				Timestamp: m[1],
				Level:     "", // Not present in systemd format
				Unit:      m[3],
				Message:   m[4],
			}
			inMultiline = true
		} else if m := genericRe.FindStringSubmatch(line); m != nil {
			if inMultiline {
				entries = append(entries, current)
				inMultiline = false
			}
			current = LogEntry{
				Timestamp: m[1],
				Level:     m[2],
				Unit:      m[3],
				Message:   m[4],
			}
			inMultiline = true
		} else if m := simpleLevelRe.FindStringSubmatch(line); m != nil {
			if inMultiline {
				entries = append(entries, current)
				inMultiline = false
			}
			current = LogEntry{
				Timestamp: "",
				Level:     m[1],
				Unit:      "",
				Message:   m[2],
			}
			inMultiline = true
		} else if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			// Indented line: treat as continuation of previous message
			if inMultiline {
				current.Message += "\n" + strings.TrimLeft(line, " \t")
			}
		} else {
			// Fallback: treat as a new message with unknown fields
			if inMultiline {
				entries = append(entries, current)
				inMultiline = false
			}
			current = LogEntry{
				Timestamp: "",
				Level:     "",
				Unit:      "",
				Message:   line,
			}
			inMultiline = true
		}
	}
	if inMultiline {
		entries = append(entries, current)
	}

	return entries, nil
}

// ParseLogStream parses log lines as they arrive (real-time streaming).
// It returns a channel of LogEntry, emitting entries as soon as a new complete entry is detected.
// Usage: feed lines to the input channel; receive parsed LogEntry from the output channel.
func ParseLogStream(input <-chan string) <-chan LogEntry {
	output := make(chan LogEntry)
	go func() {
		var current LogEntry
		inMultiline := false
		// Regexes as in ParseLog
		systemdRe := regexp.MustCompile(`^([\w]{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2})\s+([^ ]+)\s+([^\[]+)(?:\[\d+\])?:\s+(.*)$`)
		genericRe := regexp.MustCompile(`^\[([^\]]+)\]\s+([A-Z]+)\s+([^:]+):\s+(.*)$`)
		simpleLevelRe := regexp.MustCompile(`^([A-Z]+)\s+(.*)$`)
		for line := range input {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			if m := systemdRe.FindStringSubmatch(line); m != nil {
				if inMultiline {
					output <- current
					inMultiline = false
				}
				current = LogEntry{Timestamp: m[1], Level: "", Unit: m[3], Message: m[4]}
				inMultiline = true
			} else if m := genericRe.FindStringSubmatch(line); m != nil {
				if inMultiline {
					output <- current
					inMultiline = false
				}
				current = LogEntry{Timestamp: m[1], Level: m[2], Unit: m[3], Message: m[4]}
				inMultiline = true
			} else if m := simpleLevelRe.FindStringSubmatch(line); m != nil {
				if inMultiline {
					output <- current
					inMultiline = false
				}
				current = LogEntry{Timestamp: "", Level: m[1], Unit: "", Message: m[2]}
				inMultiline = true
			} else if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
				if inMultiline {
					current.Message += "\n" + strings.TrimLeft(line, " \t")
				}
			} else {
				if inMultiline {
					output <- current
					inMultiline = false
				}
				current = LogEntry{Timestamp: "", Level: "", Unit: "", Message: line}
				inMultiline = true
			}
		}
		if inMultiline {
			output <- current
		}
		close(output)
	}()
	return output
}

// ParseLogWithTimeline parses log output and returns entries sorted by timestamp (if available).
// If timestamps are missing or unparseable, original order is preserved.
func ParseLogWithTimeline(log string) ([]LogEntry, error) {
	entries, err := ParseLog(log)
	if err != nil {
		return nil, err
	}
	// Simple sort: if all entries have timestamps, sort by them (lexical sort is sufficient for syslog-style)
	hasTimestamps := true
	for _, e := range entries {
		if e.Timestamp == "" {
			hasTimestamps = false
			break
		}
	}
	if hasTimestamps {
		sort.SliceStable(entries, func(i, j int) bool {
			return entries[i].Timestamp < entries[j].Timestamp
		})
	}
	return entries, nil
}

// CorrelateLogEntriesByUnit groups log entries by their Unit field for root cause analysis.
func CorrelateLogEntriesByUnit(entries []LogEntry) map[string][]LogEntry {
	byUnit := make(map[string][]LogEntry)
	for _, e := range entries {
		unit := e.Unit
		if unit == "" {
			unit = "(unknown)"
		}
		byUnit[unit] = append(byUnit[unit], e)
	}
	return byUnit
}

// ParseNixConfig takes a NixOS configuration file content as a string and returns a structured representation.
func ParseNixConfig(config string) (map[string]interface{}, error) {
	if config == "" {
		return nil, errors.New("configuration input is empty")
	}

	parsedConfig := make(map[string]interface{})
	lines := strings.Split(config, "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}
		// Simple parsing logic: split by the first '='
		parts := strings.SplitN(line, "=", 2)
		if len(parts) < 2 {
			parsedConfig[parts[0]] = nil
		} else {
			parsedConfig[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	return parsedConfig, nil
}
