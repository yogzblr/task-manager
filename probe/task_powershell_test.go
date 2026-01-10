package probe

import (
	"context"
	"runtime"
	"testing"
)

func TestPowerShellTask(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("PowerShell task only works on Windows")
	}
	
	task := &PowerShellTask{}
	
	config := map[string]interface{}{
		"script": "Write-Output 'Hello from PowerShell'",
	}
	
	err := task.Configure(config)
	if err != nil {
		t.Fatalf("Configure failed: %v", err)
	}
	
	result, err := task.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	
	resultMap := result.(map[string]interface{})
	output := resultMap["output"].(string)
	
	if output == "" {
		t.Errorf("Expected non-empty output")
	}
}

func TestPowerShellTaskNonWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("This test is for non-Windows platforms")
	}
	
	task := &PowerShellTask{}
	
	config := map[string]interface{}{
		"script": "Write-Output 'test'",
	}
	
	err := task.Configure(config)
	if err == nil {
		t.Errorf("Expected error on non-Windows platform")
	}
}

func TestPowerShellTaskMissingScript(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("PowerShell task only works on Windows")
	}
	
	task := &PowerShellTask{}
	
	config := map[string]interface{}{}
	
	err := task.Configure(config)
	if err == nil {
		t.Errorf("Expected error for missing script")
	}
}

func TestPowerShellTaskExitCode(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("PowerShell task only works on Windows")
	}
	
	task := &PowerShellTask{}
	
	config := map[string]interface{}{
		"script": "exit 1",
	}
	
	err := task.Configure(config)
	if err != nil {
		t.Fatalf("Configure failed: %v", err)
	}
	
	_, err = task.Execute(context.Background())
	if err == nil {
		t.Errorf("Expected error for non-zero exit code")
	}
}
