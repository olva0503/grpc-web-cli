package template

import (
	"testing"
)

func TestSubstitute(t *testing.T) {
	vars := map[string]interface{}{
		"token": "secret-123",
		"id":    42,
		"host":  "localhost",
	}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"No subs", "plain text", "plain text"},
		{"Single sub", "Bearer {{token}}", "Bearer secret-123"},
		{"Multiple subs", "http://{{host}}/api/{{id}}", "http://localhost/api/42"},
		{"Unknown var", "Hello {{name}}", "Hello {{name}}"},
		{"Mixed known unknown", "{{token}} and {{name}}", "secret-123 and {{name}}"},
		{"Nil map", "Hello {{token}}", "Hello {{token}}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputVars := vars
			if tt.name == "Nil map" {
				inputVars = nil
			}
			if got := Substitute(tt.input, inputVars); got != tt.want {
				t.Errorf("Substitute() = %v, want %v", got, tt.want)
			}
		})
	}
}
