package file

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// RequestFile represents a parsed .grpc request file
type RequestFile struct {
	Name     string            // Optional request name (from comment)
	Address  string            // Server address (from GRPC line)
	Service  string            // Fully qualified service name
	Method   string            // Method name
	Protocol string            // grpc, grpc-web, or connect
	Timeout  time.Duration     // Request timeout
	Headers  map[string]string // HTTP headers
	Body     string            // JSON request body
	Captures map[string]string // Captured variables from response
	Asserts  []Assertion       // List of assertions
}

// Assertion represents a check to be performed on the response
type Assertion struct {
	Type     string // "jsonpath", "header", "status"
	Key      string // jsonpath expression or header name
	Operator string // "==", "!=", "contains"
	Value    string // Expected value (as string)
}

// Parse reads and parses a .grpc request file (returns first request only)
func Parse(path string) (*RequestFile, error) {
	requests, err := ParseMultiple(path)
	if err != nil {
		return nil, err
	}
	if len(requests) == 0 {
		return nil, fmt.Errorf("no requests found in file")
	}
	return requests[0], nil
}

// ParseMultiple reads and parses a .grpc file containing one or more requests
// Requests are separated by "---" on its own line
func ParseMultiple(path string) ([]*RequestFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open request file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	var sections [][]string
	var currentSection []string

	for scanner.Scan() {
		line := scanner.Text()
		// Check for separator
		if strings.TrimSpace(line) == "---" {
			if len(currentSection) > 0 {
				sections = append(sections, currentSection)
				currentSection = nil
			}
			continue
		}
		currentSection = append(currentSection, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Don't forget the last section
	if len(currentSection) > 0 {
		sections = append(sections, currentSection)
	}

	if len(sections) == 0 {
		return nil, fmt.Errorf("no requests found in file")
	}

	var requests []*RequestFile
	for i, section := range sections {
		req, err := parseContent(section, i+1)
		if err != nil {
			return nil, fmt.Errorf("request %d: %w", i+1, err)
		}
		requests = append(requests, req)
	}

	return requests, nil
}

// parseContent parses a single request from lines of text
func parseContent(lines []string, requestNum int) (*RequestFile, error) {

	// Move body lines processing earlier or handle logic flow:
	// The previous loop was skipping lines inside `inBody`.
	// We need to refactor the loop to correctly handle sections.
	// However, looking at the previous code, it was mixing line-by-line parsing with a flag.

	// Let's rewrite the parsing loop logic slightly to support sections.
	// Re-parsing the lines is cleaner.

	req := &RequestFile{
		Protocol: "grpc-web",
		Timeout:  30 * time.Second,
		Headers:  make(map[string]string),
		Captures: make(map[string]string),
	}

	var currentSection string // "", "Body", "Captures", "Asserts"
	var bodyLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines if not in a body/block
		if currentSection == "" && trimmed == "" {
			continue
		}

		// Handle comments
		if strings.HasPrefix(trimmed, "#") {
			if req.Name == "" {
				req.Name = strings.TrimPrefix(trimmed, "#")
				req.Name = strings.TrimSpace(req.Name)
			}
			continue
		}

		// Detect section headers
		if trimmed == "[Captures]" {
			currentSection = "Captures"
			continue
		}
		if trimmed == "[Asserts]" {
			currentSection = "Asserts"
			continue
		}

		// If we are in Captures section
		if currentSection == "Captures" {
			if trimmed == "" {
				continue
			}
			// Parse key: value
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				// invalid capture line, maybe log warning or error?
				// content parser returning error is better
				continue // strict parsing might fail on empty lines or comments
			}
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			req.Captures[key] = val
			continue
		}

		// If we are in Asserts section
		if currentSection == "Asserts" {
			if trimmed == "" {
				continue
			}
			// Parse assertion: type "key" op "value"
			// Example: jsonpath "$.id" == "123"
			// Simple parser avoiding regex for now to avoid dependency complexity if possible,
			// but regex is clearer for quoted strings.
			// Let's use specific logic for the expected format.
			parts := strings.Fields(trimmed)
			if len(parts) >= 4 {
				// We need to handle quotes. strings.Fields splits by space, breaking quoted strings.
				// Let's rely on a helper or regex.
				// Given the constraints and simplicity, let's try a custom split function or just regex.
				// A simple way for now: assumes standard formatting jsonpath "key" op "value"

				// Re-parsing line to handle quotes properly
				// Format: <type> <key_q> <op> <value_q>
				// or: <type> <key_q> <op> <value_raw> (if value is number/bool)

				// Let's use a robust approach: find first space, then parse first quoted string, etc.
				// But to keep it simple and consistent with strict hurl-like syntax:

				// 1. Type
				firstSpace := strings.Index(trimmed, " ")
				if firstSpace == -1 {
					continue
				}
				aType := trimmed[:firstSpace]
				rest := strings.TrimSpace(trimmed[firstSpace:])

				// 2. Key (quoted)
				if !strings.HasPrefix(rest, "\"") {
					continue
				}
				rest = rest[1:] // skip open quote
				endQuote := strings.Index(rest, "\"")
				if endQuote == -1 {
					continue
				}
				key := rest[:endQuote]
				rest = strings.TrimSpace(rest[endQuote+1:])

				// 3. Operator
				firstSpace = strings.Index(rest, " ")
				if firstSpace == -1 {
					continue
				}
				op := rest[:firstSpace]
				rest = strings.TrimSpace(rest[firstSpace:])

				// 4. Value (quoted or raw)
				var val string
				if strings.HasPrefix(rest, "\"") {
					// create valid string from quoted
					rest = rest[1:]
					endQuote = strings.LastIndex(rest, "\"") // Use LastIndex to handle simple cases? No, strict.
					// Actually, value might contain quotes.
					// For simple implementation, let's assume valid JSON string or simple string.
					// Let's just take until the end quote?
					endQuote = strings.Index(rest, "\"")
					if endQuote != -1 {
						val = rest[:endQuote]
					}
				} else {
					val = rest
				}

				req.Asserts = append(req.Asserts, Assertion{
					Type:     aType,
					Key:      key,
					Operator: op,
					Value:    val,
				})
			}
			continue
		}

		// Detect Body start (if not already strictly defined, implicit JSON body starts with {)
		if currentSection == "" && strings.HasPrefix(trimmed, "{") {
			currentSection = "Body"
		}

		if currentSection == "Body" {
			bodyLines = append(bodyLines, line)
			continue
		}

		// Parse GRPC address line
		if strings.HasPrefix(line, "GRPC ") {
			req.Address = strings.TrimSpace(strings.TrimPrefix(line, "GRPC"))
			continue
		}

		// Parse key: value pairs for main section
		colonIdx := strings.Index(line, ":")
		if colonIdx != -1 {
			key := strings.TrimSpace(line[:colonIdx])
			value := strings.TrimSpace(line[colonIdx+1:])

			switch key {
			case "Service":
				req.Service = value
			case "Method":
				req.Method = value
			case "Protocol":
				req.Protocol = value
			case "Timeout":
				duration, err := time.ParseDuration(value)
				if err != nil {
					return nil, fmt.Errorf("invalid timeout duration %q: %w", value, err)
				}
				req.Timeout = duration
			default:
				// Treat as HTTP header
				req.Headers[key] = value
			}
			continue
		}
	}

	if len(bodyLines) > 0 {
		req.Body = strings.Join(bodyLines, "\n")
	} else {
		req.Body = "{}"
	}

	// Validate required fields
	if req.Address == "" {
		return nil, fmt.Errorf("missing required 'GRPC <address>' line")
	}
	if req.Service == "" {
		return nil, fmt.Errorf("missing required 'Service:' field")
	}
	if req.Method == "" {
		return nil, fmt.Errorf("missing required 'Method:' field")
	}

	return req, nil
}
