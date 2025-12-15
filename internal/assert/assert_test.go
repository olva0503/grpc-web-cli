package assert

import (
	"grpc_client/internal/file"
	"testing"
)

func TestCheck(t *testing.T) {
	jsonOutput := `{"id": "123", "status": "active", "items": ["item1", "item2"]}`

	tests := []struct {
		name      string
		assertion file.Assertion
		wantPass  bool
		wantMsg   string
	}{
		{
			name: "Equals match",
			assertion: file.Assertion{
				Type:     "jsonpath",
				Key:      "$.id",
				Operator: "==",
				Value:    "123",
			},
			wantPass: true,
			wantMsg:  `PASS: jsonpath "$.id" == "123"`,
		},
		{
			name: "Equals mismatch",
			assertion: file.Assertion{
				Type:     "jsonpath",
				Key:      "$.id",
				Operator: "==",
				Value:    "456",
			},
			wantPass: false,
			wantMsg:  `FAIL: jsonpath "$.id" == "456" (actual: "123")`,
		},
		{
			name: "Not equals match",
			assertion: file.Assertion{
				Type:     "jsonpath",
				Key:      "$.status",
				Operator: "!=",
				Value:    "inactive",
			},
			wantPass: true,
			wantMsg:  `PASS: jsonpath "$.status" != "inactive"`,
		},
		{
			name: "Not equals mismatch",
			assertion: file.Assertion{
				Type:     "jsonpath",
				Key:      "$.status",
				Operator: "!=",
				Value:    "active",
			},
			wantPass: false,
			wantMsg:  `FAIL: jsonpath "$.status" != "active" (actual: "active")`,
		},
		{
			name: "Contains match",
			assertion: file.Assertion{
				Type:     "jsonpath",
				Key:      "$.items[0]",
				Operator: "contains",
				Value:    "item",
			},
			wantPass: true,
			wantMsg:  `PASS: jsonpath "$.items[0]" contains "item"`,
		},
		{
			name: "Contains mismatch",
			assertion: file.Assertion{
				Type:     "jsonpath",
				Key:      "$.items[0]",
				Operator: "contains",
				Value:    "xyz",
			},
			wantPass: false,
			wantMsg:  `FAIL: jsonpath "$.items[0]" contains "xyz" (actual: "item1")`,
		},
		{
			name: "Unknown operator",
			assertion: file.Assertion{
				Type:     "jsonpath",
				Key:      "$.id",
				Operator: "unknown",
				Value:    "123",
			},
			wantPass: false,
			wantMsg:  "unknown operator 'unknown'",
		},
		{
			name: "Invalid JSONPath",
			assertion: file.Assertion{
				Type:     "jsonpath",
				Key:      "$.items[",
				Operator: "==",
				Value:    "123",
			},
			wantPass: false,
			wantMsg:  "failed to evaluate jsonpath '$.items[': unclosed array index in path: [",
		},
		{
			name: "Unknown assertion type",
			assertion: file.Assertion{
				Type:     "header",
				Key:      "Content-Type",
				Operator: "==",
				Value:    "application/json",
			},
			wantPass: true, // Treated as warning
			wantMsg:  "Warning: skipping unknown assertion type 'header'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := Check(tt.assertion, jsonOutput)
			if result.Pass != tt.wantPass {
				t.Errorf("Check() pass = %v, want %v", result.Pass, tt.wantPass)
			}
			if result.Message != tt.wantMsg {
				t.Errorf("Check() message = %q, want %q", result.Message, tt.wantMsg)
			}
		})
	}
}
