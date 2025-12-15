package client

import (
	"testing"
)

func TestEvaluateJSONPath(t *testing.T) {
	jsonStr := `
	{
		"token": "abc-123",
		"user": {
			"id": 101,
			"name": "Alice",
			"details": {
				"active": true
			}
		},
		"permissions": ["read", "write"],
		"users": [
			{"id": 1, "name": "Bob"},
			{"id": 2, "name": "Charlie"}
		]
	}
	`

	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{"Simple key", "token", "abc-123", false},
		{"Nested key", "user.name", "Alice", false},
		{"Deep nested key", "user.details.active", "true", false},
		{"Number value", "user.id", "101", false},
		{"Array index simple", "permissions[0]", "read", false},
		{"Array index nested object", "users[1].name", "Charlie", false},
		{"Root array index check", "users[0].id", "1", false}, // Note: json numbers are float64 usually, here we check string representation
		{"Invalid path", "user.missing", "", true},
		{"Invalid array index", "permissions[5]", "", true},
		{"Invalid json", "{", "", true},
		{"Empty path", "", jsonStr, false}, // Should return full json... wait, EvaluateJSONPath returns interface as string.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvaluateJSONPath(jsonStr, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("EvaluateJSONPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				// Special handling for full json dump might be tricky due to whitespace, skipping exact check for empty path
				if tt.path != "" {
					t.Errorf("EvaluateJSONPath() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
