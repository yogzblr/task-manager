package probe

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProbeHTTPTask(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	p := New()

	yaml := fmt.Sprintf(`
name: test-http
tasks:
  - name: check-test-server
    type: http
    config:
      url: %s
      expected_status: [200]
`, server.URL)

	result, err := p.ExecuteYAML(context.Background(), []byte(yaml))
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got failure")
	}

	if len(result.Tasks) != 1 {
		t.Errorf("Expected 1 task result, got %d", len(result.Tasks))
	}
}

func TestProbeCommandTask(t *testing.T) {
	p := New()
	
	yaml := `
name: test-command
tasks:
  - name: echo-test
    type: command
    config:
      command: echo
      args: ["hello", "world"]
`
	
	result, err := p.ExecuteYAML(context.Background(), []byte(yaml))
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}
	
	if !result.Success {
		t.Errorf("Expected success, got failure")
	}
}

func TestProbeUnknownTask(t *testing.T) {
	p := New()
	
	yaml := `
name: test-unknown
tasks:
  - name: unknown
    type: nonexistent
    config:
      foo: bar
`
	
	_, err := p.ExecuteYAML(context.Background(), []byte(yaml))
	if err == nil {
		t.Errorf("Expected error for unknown task type")
	}
}

func TestProbeInvalidYAML(t *testing.T) {
	p := New()
	
	yaml := `invalid yaml: [[[`
	
	_, err := p.ExecuteYAML(context.Background(), []byte(yaml))
	if err == nil {
		t.Errorf("Expected error for invalid YAML")
	}
}
