package probe

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"
)

// DownloadExecTask downloads and executes files with verification
type DownloadExecTask struct {
	URL       string
	SHA256    string
	Signature string
	PublicKey string
	Args      []string
	Timeout   time.Duration
	Cleanup   bool
}

// Configure sets up the DownloadExec task
func (t *DownloadExecTask) Configure(config map[string]interface{}) error {
	// URL is required
	url, ok := config["url"].(string)
	if !ok || url == "" {
		return fmt.Errorf("url is required")
	}
	t.URL = url
	
	// SHA256 is required
	sha256, ok := config["sha256"].(string)
	if !ok || sha256 == "" {
		return fmt.Errorf("sha256 is required")
	}
	t.SHA256 = sha256
	
	// Signature is optional (but recommended)
	if signature, ok := config["signature"].(string); ok {
		t.Signature = signature
	}
	
	// Public key is required if signature is provided
	if t.Signature != "" {
		if publicKey, ok := config["public_key"].(string); ok {
			t.PublicKey = publicKey
		} else {
			return fmt.Errorf("public_key is required when signature is provided")
		}
	}
	
	// Args (optional)
	if args, ok := config["args"].([]interface{}); ok {
		t.Args = make([]string, len(args))
		for i, arg := range args {
			if strArg, ok := arg.(string); ok {
				t.Args[i] = strArg
			}
		}
	}
	
	// Timeout (default: 60s)
	if timeoutStr, ok := config["timeout"].(string); ok {
		duration, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return fmt.Errorf("invalid timeout: %w", err)
		}
		t.Timeout = duration
	} else {
		t.Timeout = 60 * time.Second
	}
	
	// Cleanup (default: true)
	if cleanup, ok := config["cleanup"].(bool); ok {
		t.Cleanup = cleanup
	} else {
		t.Cleanup = true
	}
	
	return nil
}

// Execute downloads, verifies, and executes the file
func (t *DownloadExecTask) Execute(ctx context.Context) (interface{}, error) {
	// Download file
	tempFile, err := t.download(ctx)
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}
	
	// Clean up if requested
	if t.Cleanup {
		defer os.Remove(tempFile)
	}
	
	// Verify SHA256 checksum
	if err := t.verifySHA256(tempFile); err != nil {
		return nil, fmt.Errorf("SHA256 verification failed: %w", err)
	}
	
	// Verify signature if provided
	if t.Signature != "" {
		if err := t.verifySignature(tempFile); err != nil {
			return nil, fmt.Errorf("signature verification failed: %w", err)
		}
	}
	
	// Make executable (Unix only)
	if err := os.Chmod(tempFile, 0755); err != nil {
		// Ignore error on Windows
	}
	
	// Execute file
	return t.execute(ctx, tempFile)
}

func (t *DownloadExecTask) download(ctx context.Context) (string, error) {
	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", t.URL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	// Perform request
	client := &http.Client{
		Timeout: t.Timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "probe-downloadexec-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()
	
	// Copy response body to file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	
	return tmpFile.Name(), nil
}

func (t *DownloadExecTask) verifySHA256(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("failed to compute hash: %w", err)
	}
	
	actual := hex.EncodeToString(hash.Sum(nil))
	if actual != t.SHA256 {
		return fmt.Errorf("SHA256 mismatch: expected %s, got %s", t.SHA256, actual)
	}
	
	return nil
}

func (t *DownloadExecTask) verifySignature(filePath string) error {
	// Read file contents
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	// Decode signature
	sigBytes, err := base64.StdEncoding.DecodeString(t.Signature)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}
	
	// Decode public key
	pubKeyBytes, err := base64.StdEncoding.DecodeString(t.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to decode public key: %w", err)
	}
	
	if len(pubKeyBytes) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid public key size: expected %d, got %d", ed25519.PublicKeySize, len(pubKeyBytes))
	}
	
	publicKey := ed25519.PublicKey(pubKeyBytes)
	
	// Verify signature
	if !ed25519.Verify(publicKey, data, sigBytes) {
		return fmt.Errorf("signature verification failed")
	}
	
	return nil
}

func (t *DownloadExecTask) execute(ctx context.Context, filePath string) (interface{}, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, t.Timeout)
	defer cancel()
	
	// Execute file
	cmd := exec.CommandContext(ctx, filePath, t.Args...)
	
	output, err := cmd.CombinedOutput()
	exitCode := 0
	
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return nil, fmt.Errorf("execution failed: %w", err)
		}
	}
	
	result := map[string]interface{}{
		"output":    string(output),
		"exit_code": exitCode,
	}
	
	if exitCode != 0 {
		return result, fmt.Errorf("execution exited with code %d", exitCode)
	}
	
	return result, nil
}
