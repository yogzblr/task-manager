package probe

import (
	"context"
	"testing"
)

func TestProbeHTTPTask(t *testing.T) {
	p := New()
	
	yaml := `
name: test-http
tasks:
  - name: check-google
    type: http
    config:
      url: https://www.google.com
      expected_status: [200]
`
	
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
