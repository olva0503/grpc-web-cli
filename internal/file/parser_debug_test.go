package file

import (
	"strings"
	"testing"
)

func TestParseCapturesReproduction(t *testing.T) {
	// Read the actual file causing issues
	filePath := "../../test.grpc"
	reqs, err := ParseMultiple(filePath)
	if err != nil {
		t.Fatalf("ParseMultiple failed: %v", err)
	}

	if len(reqs) == 0 {
		t.Fatalf("No requests parsed")
	}

	req := reqs[0]

	if strings.Contains(req.Body, "[Captures]") {
		t.Errorf("Body should not contain [Captures], got:\n%s", req.Body)
		// Debug: print hex of the body
		t.Logf("Body bytes: % x", req.Body)
	}

	// Also check captures
	if len(req.Captures) == 0 {
		t.Errorf("Expected captures to be parsed, got empty")
	}
}
