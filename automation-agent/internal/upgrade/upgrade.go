package upgrade

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

// Upgrader handles agent binary upgrades
type Upgrader struct {
	binaryPath string
	verifier   Verifier
}

// Verifier verifies upgrade signatures
type Verifier interface {
	VerifySignature(data []byte, signature, keyID string) error
}

// NewUpgrader creates a new upgrader
func NewUpgrader(binaryPath string, verifier Verifier) *Upgrader {
	return &Upgrader{
		binaryPath: binaryPath,
		verifier:   verifier,
	}
}

// UpgradeInfo represents upgrade information
type UpgradeInfo struct {
	Version   string
	URL       string
	SHA256    string
	Signature string
	KeyID     string
}

// Upgrade performs an atomic binary upgrade
func (u *Upgrader) Upgrade(ctx context.Context, info UpgradeInfo) error {
	// Download new binary to temp location
	tempFile, err := u.download(ctx, info.URL)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer os.Remove(tempFile)
	
	// Verify SHA256
	if err := u.verifySHA256(tempFile, info.SHA256); err != nil {
		return fmt.Errorf("SHA256 verification failed: %w", err)
	}
	
	// Verify signature
	if u.verifier != nil {
		data, err := os.ReadFile(tempFile)
		if err != nil {
			return fmt.Errorf("failed to read temp file: %w", err)
		}
		
		if err := u.verifier.VerifySignature(data, info.Signature, info.KeyID); err != nil {
			return fmt.Errorf("signature verification failed: %w", err)
		}
	}
	
	// Perform atomic swap
	if err := u.atomicSwap(tempFile, u.binaryPath); err != nil {
		return fmt.Errorf("atomic swap failed: %w", err)
	}
	
	return nil
}

func (u *Upgrader) download(ctx context.Context, url string) (string, error) {
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
	
	tmpFile, err := os.CreateTemp("", "automation-agent-upgrade-*")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()
	
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}
	
	return tmpFile.Name(), nil
}

func (u *Upgrader) verifySHA256(filePath, expected string) error {
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

func (u *Upgrader) atomicSwap(tempFile, targetFile string) error {
	if runtime.GOOS == "windows" {
		// Windows: Use MoveFileEx with MOVEFILE_REPLACE_EXISTING
		// This requires syscall which is platform-specific
		// For now, use a simple rename (not truly atomic on Windows)
		return os.Rename(tempFile, targetFile)
	} else {
		// Unix: Use rename which is atomic
		return os.Rename(tempFile, targetFile)
	}
}

// GetCurrentBinaryPath returns the path to the current binary
func GetCurrentBinaryPath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	
	// Resolve symlinks
	resolved, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		return execPath, nil
	}
	
	return resolved, nil
}
