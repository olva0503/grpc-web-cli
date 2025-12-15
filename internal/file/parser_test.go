package file

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseMultiple_SingleRequest(t *testing.T) {
	content := `# Test request
GRPC http://localhost:8080/api
Service: example.UserService
Method: GetUser
Authorization: Bearer token123
{
    "user_id": "123"
}`

	requests := parseTestContent(t, content)

	if len(requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(requests))
	}

	req := requests[0]
	if req.Name != "Test request" {
		t.Errorf("expected name 'Test request', got %q", req.Name)
	}
	if req.Address != "http://localhost:8080/api" {
		t.Errorf("expected address 'http://localhost:8080/api', got %q", req.Address)
	}
	if req.Service != "example.UserService" {
		t.Errorf("expected service 'example.UserService', got %q", req.Service)
	}
	if req.Method != "GetUser" {
		t.Errorf("expected method 'GetUser', got %q", req.Method)
	}
	if req.Headers["Authorization"] != "Bearer token123" {
		t.Errorf("expected Authorization header, got %v", req.Headers)
	}
}

func TestParseMultiple_MultipleRequests(t *testing.T) {
	content := `# First request
GRPC http://localhost:8080/api
Service: example.UserService
Method: GetUser
{
    "user_id": "1"
}

---

# Second request
GRPC http://localhost:8080/api
Service: example.UserService
Method: UpdateUser
{
    "user_id": "2"
}

---

# Third request
GRPC http://localhost:8080/api
Service: example.OrderService
Method: CreateOrder
{
    "order_id": "3"
}`

	requests := parseTestContent(t, content)

	if len(requests) != 3 {
		t.Fatalf("expected 3 requests, got %d", len(requests))
	}

	// Verify first request
	if requests[0].Name != "First request" {
		t.Errorf("request 1: expected name 'First request', got %q", requests[0].Name)
	}
	if requests[0].Method != "GetUser" {
		t.Errorf("request 1: expected method 'GetUser', got %q", requests[0].Method)
	}

	// Verify second request
	if requests[1].Name != "Second request" {
		t.Errorf("request 2: expected name 'Second request', got %q", requests[1].Name)
	}
	if requests[1].Method != "UpdateUser" {
		t.Errorf("request 2: expected method 'UpdateUser', got %q", requests[1].Method)
	}

	// Verify third request
	if requests[2].Name != "Third request" {
		t.Errorf("request 3: expected name 'Third request', got %q", requests[2].Name)
	}
	if requests[2].Service != "example.OrderService" {
		t.Errorf("request 3: expected service 'example.OrderService', got %q", requests[2].Service)
	}
}

func TestParseMultiple_DefaultValues(t *testing.T) {
	content := `GRPC http://localhost:8080
Service: example.Service
Method: DoSomething`

	requests := parseTestContent(t, content)

	if len(requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(requests))
	}

	req := requests[0]
	if req.Protocol != "grpc-web" {
		t.Errorf("expected default protocol 'grpc-web', got %q", req.Protocol)
	}
	if req.Timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %v", req.Timeout)
	}
	if req.Body != "{}" {
		t.Errorf("expected default empty body '{}', got %q", req.Body)
	}
}

func TestParseMultiple_CustomTimeout(t *testing.T) {
	content := `GRPC http://localhost:8080
Service: example.Service
Method: DoSomething
Timeout: 5s
{}`

	requests := parseTestContent(t, content)

	if requests[0].Timeout != 5*time.Second {
		t.Errorf("expected timeout 5s, got %v", requests[0].Timeout)
	}
}

func TestParseMultiple_CustomProtocol(t *testing.T) {
	content := `GRPC http://localhost:8080
Service: example.Service
Method: DoSomething
Protocol: connect
{}`

	requests := parseTestContent(t, content)

	if requests[0].Protocol != "connect" {
		t.Errorf("expected protocol 'connect', got %q", requests[0].Protocol)
	}
}

func TestParseMultiple_MissingAddress(t *testing.T) {
	content := `Service: example.Service
Method: DoSomething
{}`

	_, err := parseTestContentWithError(content)
	if err == nil {
		t.Error("expected error for missing address")
	}
}

func TestParseMultiple_MissingService(t *testing.T) {
	content := `GRPC http://localhost:8080
Method: DoSomething
{}`

	_, err := parseTestContentWithError(content)
	if err == nil {
		t.Error("expected error for missing service")
	}
}

func TestParseMultiple_MissingMethod(t *testing.T) {
	content := `GRPC http://localhost:8080
Service: example.Service
{}`

	_, err := parseTestContentWithError(content)
	if err == nil {
		t.Error("expected error for missing method")
	}
}

func TestParseMultiple_MultipleHeaders(t *testing.T) {
	content := `GRPC http://localhost:8080
Service: example.Service
Method: DoSomething
Authorization: Bearer token
X-Custom-Header: custom-value
X-Tenant: my-tenant
{}`

	requests := parseTestContent(t, content)

	headers := requests[0].Headers
	if headers["Authorization"] != "Bearer token" {
		t.Errorf("expected Authorization header")
	}
	if headers["X-Tenant"] != "my-tenant" {
		t.Errorf("expected X-Tenant header")
	}
}

func TestParseMultiple_Asserts(t *testing.T) {
	content := `GRPC http://localhost:8080
Service: example.Service
Method: GetData
{}

[Asserts]
jsonpath "$.status" == "active"
jsonpath "$.count" == "10"
jsonpath "$.items[0]" contains "item1"`

	requests := parseTestContent(t, content)

	if len(requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(requests))
	}

	req := requests[0]
	if len(req.Asserts) != 3 {
		t.Fatalf("expected 3 assertions, got %d", len(req.Asserts))
	}

	// Verify first assertion
	a1 := req.Asserts[0]
	if a1.Type != "jsonpath" || a1.Key != "$.status" || a1.Operator != "==" || a1.Value != "active" {
		t.Errorf("assertion 1 mismatch: %+v", a1)
	}

	// Verify second assertion
	a2 := req.Asserts[1]
	if a2.Type != "jsonpath" || a2.Key != "$.count" || a2.Operator != "==" || a2.Value != "10" {
		t.Errorf("assertion 2 mismatch: %+v", a2)
	}

	// Verify third assertion
	a3 := req.Asserts[2]
	if a3.Type != "jsonpath" || a3.Key != "$.items[0]" || a3.Operator != "contains" || a3.Value != "item1" {
		t.Errorf("assertion 3 mismatch: %+v", a3)
	}
}

func TestParse_BackwardCompatibility(t *testing.T) {
	content := `# Single request
GRPC http://localhost:8080
Service: example.Service
Method: DoSomething
{}`

	tmpFile := createTempFile(t, content)
	defer func() {
		_ = os.Remove(tmpFile)
	}()

	// Test that Parse still works (returns first request)
	req, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if req.Name != "Single request" {
		t.Errorf("expected name 'Single request', got %q", req.Name)
	}
	if req.Method != "DoSomething" {
		t.Errorf("expected method 'DoSomething', got %q", req.Method)
	}
}

// Helper functions

func parseTestContent(t *testing.T, content string) []*RequestFile {
	t.Helper()
	requests, err := parseTestContentWithError(content)
	if err != nil {
		t.Fatalf("ParseMultiple failed: %v", err)
	}
	return requests
}

func parseTestContentWithError(content string) ([]*RequestFile, error) {
	tmpFile := createTempFileHelper(content)
	defer func() {
		_ = os.Remove(tmpFile)
	}()
	return ParseMultiple(tmpFile)
}

func createTempFile(t *testing.T, content string) string {
	t.Helper()
	return createTempFileHelper(content)
}

func createTempFileHelper(content string) string {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_request.grpc")
	_ = os.WriteFile(tmpFile, []byte(content), 0644)
	return tmpFile
}
