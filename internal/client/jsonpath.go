package client

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// EvaluateJSONPath extracts a value from a JSON string using a simple path syntax.
// Supported syntax:
// - Dot notation: user.details.name
// - Array indexing: users[0].id
func EvaluateJSONPath(jsonStr string, path string) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", fmt.Errorf("invalid JSON response: %w", err)
	}

	result, err := evaluatePath(data, path)
	if err != nil {
		return "", err
	}

	// Convert result to string
	return fmt.Sprintf("%v", result), nil
}

func evaluatePath(data interface{}, path string) (interface{}, error) {
	// Strip optional root selector
	if strings.HasPrefix(path, "$.") {
		path = strings.TrimPrefix(path, "$.")
	} else if strings.HasPrefix(path, "$") {
		path = strings.TrimPrefix(path, "$")
	}

	if path == "" {
		return data, nil
	}

	// Handle array indexing at the start of the path, e.g. [0].name
	if strings.HasPrefix(path, "[") {
		endIdx := strings.Index(path, "]")
		if endIdx == -1 {
			return nil, fmt.Errorf("unclosed array index in path: %s", path)
		}

		idxStr := path[1:endIdx]
		idx, err := strconv.Atoi(idxStr)
		if err != nil {
			return nil, fmt.Errorf("invalid array index '%s': %w", idxStr, err)
		}

		slice, ok := data.([]interface{})
		if !ok {
			return nil, fmt.Errorf("expected array but got %T", data)
		}

		if idx < 0 || idx >= len(slice) {
			return nil, fmt.Errorf("array index out of bounds: %d", idx)
		}

		remainingPath := path[endIdx+1:]
		remainingPath = strings.TrimPrefix(remainingPath, ".")

		return evaluatePath(slice[idx], remainingPath)
	}

	// Handle dot notation
	parts := strings.SplitN(path, ".", 2)
	key := parts[0]

	// Check if key has array index like users[0]
	bracketIdx := strings.Index(key, "[")
	if bracketIdx != -1 {
		// This handle cases like users[0] where [0] is part of the first segment
		// Logic needs to be careful.
		// Actually, simplest is to treat "users[0]" as property "users" then index [0]
		// So if we find [, we split there.
		realKey := key[:bracketIdx]
		arrayPart := key[bracketIdx:]

		// Update path to process array part next
		remainingPath := arrayPart
		if len(parts) > 1 {
			remainingPath += "." + parts[1]
		}

		// Process key access first
		obj, ok := data.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("expected object for key '%s' but got %T", realKey, data)
		}

		val, ok := obj[realKey]
		if !ok {
			return nil, fmt.Errorf("key '%s' not found", realKey)
		}

		return evaluatePath(val, remainingPath)
	}

	obj, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected object for key '%s' but got %T", key, data)
	}

	val, ok := obj[key]
	if !ok {
		return nil, fmt.Errorf("key '%s' not found", key)
	}

	if len(parts) > 1 {
		return evaluatePath(val, parts[1])
	}

	return val, nil
}
