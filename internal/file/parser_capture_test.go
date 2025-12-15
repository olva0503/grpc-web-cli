package file

import (
	"strings"
	"testing"
)

func TestParseCaptures(t *testing.T) {
	content := `
GRPC http://localhost:8080
Service: svc
Method: method
{ "json": "body" }

[Captures]
var1: path.to.val
var2: array[0]
`
	lines := strings.Split(strings.TrimSpace(content), "\n")
	req, err := parseContent(lines, 1)
	if err != nil {
		t.Fatalf("parseContent failed: %v", err)
	}

	if len(req.Captures) != 2 {
		t.Errorf("Expected 2 captures, got %d", len(req.Captures))
	}

	if req.Captures["var1"] != "path.to.val" {
		t.Errorf("Expected var1=path.to.val, got %s", req.Captures["var1"])
	}
	if req.Captures["var2"] != "array[0]" {
		t.Errorf("Expected var2=array[0], got %s", req.Captures["var2"])
	}
}
