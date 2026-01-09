package downloadexec

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/automation-platform/agent/internal/probe"
	"github.com/automation-platform/agent/internal/security"
)

// DownloadExecPlugin downloads and executes artifacts
type DownloadExecPlugin struct {
	verifier *security.Verifier
}

// New creates a new download_exec plugin
func New(verifier *security.Verifier) *DownloadExecPlugin {
	return &DownloadExecPlugin{
		verifier: verifier,
	}
}

// Name returns the plugin name
func (p *DownloadExecPlugin) Name() string {
	return "download_exec"
}

// Execute downloads, verifies, and executes an artifact
func (p *DownloadExecPlugin) Execute(ctx context.Context, config map[string]interface{}) (*probe.Result, error) {
	artifactMap, ok := config["artifact"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("artifact is required")
	}
	
	artifactURL, _ := artifactMap["url"].(string)
	expectedSHA256, _ := artifactMap["sha256"].(string)
	signature, _ := artifactMap["signature"].(string)
	keyID, _ := artifactMap["key_id"].(string)
	
	if artifactURL == "" {
		return nil, fmt.Errorf("artifact URL is required")
	}
	
	// Download artifact
	tempFile, err := p.download(ctx, artifactURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download artifact: %w", err)
	}
	defer os.Remove(tempFile)
	
	// Verify SHA256
	if expectedSHA256 != "" {
		if err := p.verifySHA256(tempFile, expectedSHA256); err != nil {
			return nil, fmt.Errorf("SHA256 verification failed: %w", err)
		}
	}
	
	// Verify signature
	if signature != "" && keyID != "" && p.verifier != nil {
		if err := p.verifier.VerifySignature(tempFile, signature, keyID); err != nil {
			return nil, fmt.Errorf("signature verification failed: %w", err)
		}
	}
	
	// Make executable
	if err := os.Chmod(tempFile, 0755); err != nil {
		return nil, fmt.Errorf("failed to make executable: %w", err)
	}
	
	// Execute
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, tempFile)
	} else {
		cmd = exec.CommandContext(ctx, tempFile)
	}
	
	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return nil, err
		}
	}
	
	return &probe.Result{
		ExitCode: exitCode,
		Stdout:   string(output),
		Stderr:   "",
		Success:  exitCode == 0,
	}, nil
}

func (p *DownloadExecPlugin) download(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	// Create temp file
	tmpFile, err := os.CreateTemp("", "automation-agent-*")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()
	
	// Copy to temp file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}
	
	return tmpFile.Name(), nil
}

func (p *DownloadExecPlugin) verifySHA256(filePath, expected string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}
	
	actual := hex.EncodeToString(hash.Sum(nil))
	if actual != expected {
		return fmt.Errorf("SHA256 mismatch: expected %s, got %s", expected, actual)
	}
	
	return nil
}
