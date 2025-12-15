package assert

import (
	"fmt"
	"grpc_client/internal/client"
	"grpc_client/internal/file"
	"strings"
)

// Result represents the outcome of an assertion
type Result struct {
	Pass    bool
	Message string
}

// Check evaluates a single assertion against the JSON output
func Check(assert file.Assertion, jsonOutput string) (Result, error) {
	if assert.Type != "jsonpath" {
		return Result{
			Pass:    true,
			Message: fmt.Sprintf("Warning: skipping unknown assertion type '%s'", assert.Type),
		}, nil
	}

	val, err := client.EvaluateJSONPath(jsonOutput, assert.Key)
	if err != nil {
		return Result{
			Pass:    false,
			Message: fmt.Sprintf("failed to evaluate jsonpath '%s': %v", assert.Key, err),
		}, nil
	}

	pass := false
	switch assert.Operator {
	case "==":
		pass = val == assert.Value
	case "!=":
		pass = val != assert.Value
	case "contains":
		pass = strings.Contains(val, assert.Value)
	default:
		return Result{
			Pass:    false,
			Message: fmt.Sprintf("unknown operator '%s'", assert.Operator),
		}, nil
	}

	status := "FAIL"
	if pass {
		status = "PASS"
	}

	// Format: PASS: jsonpath "$.id" == "123"
	// Format: FAIL: jsonpath "$.id" == "123" (actual: "456")
	msg := fmt.Sprintf("%s: jsonpath \"%s\" %s \"%s\"", status, assert.Key, assert.Operator, assert.Value)
	if !pass {
		msg += fmt.Sprintf(" (actual: \"%s\")", val)
	}

	return Result{
		Pass:    pass,
		Message: msg,
	}, nil
}
