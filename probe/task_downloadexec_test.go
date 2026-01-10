package probe

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestDownloadExecTaskConfigure(t *testing.T) {
	task := &DownloadExecTask{}
	
	config := map[string]interface{}{
		"url":    "https://example.com/file",
		"sha256": "abc123",
	}
	
	err := task.Configure(config)
	if err != nil {
		t.Fatalf("Configure failed: %v", err)
	}
	
	if task.URL != "https://example.com/file" {
		t.Errorf("Expected URL to be set")
	}
	
	if task.SHA256 != "abc123" {
		t.Errorf("Expected SHA256 to be set")
	}
}

func TestDownloadExecTaskMissingURL(t *testing.T) {
	task := &DownloadExecTask{}
	
	config := map[string]interface{}{
		"sha256": "abc123",
	}
	
	err := task.Configure(config)
	if err == nil {
		t.Errorf("Expected error for missing URL")
	}
}

func TestDownloadExecTaskMissingSHA256(t *testing.T) {
	task := &DownloadExecTask{}
	
	config := map[string]interface{}{
		"url": "https://example.com/file",
	}
	
	err := task.Configure(config)
	if err == nil {
		t.Errorf("Expected error for missing SHA256")
	}
}

func TestDownloadExecTaskDownload(t *testing.T) {
	// Create test server
	content := []byte("test content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer server.Close()
	
	// Calculate SHA256 of test content
	hash := sha256.Sum256(content)
	sha256Hash := hex.EncodeToString(hash[:])
	
	task := &DownloadExecTask{}
	config := map[string]interface{}{
		"url":    server.URL,
		"sha256": sha256Hash,
	}
	
	err := task.Configure(config)
	if err != nil {
		t.Fatalf("Configure failed: %v", err)
	}
	
	// Test download
	filePath, err := task.download(context.Background())
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}
	defer os.Remove(filePath)
	
	// Verify file contents
	downloaded, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}
	
	if string(downloaded) != string(content) {
		t.Errorf("Downloaded content mismatch")
	}
}

func TestDownloadExecTaskSHA256Verification(t *testing.T) {
	// Create temp file with known content
	content := []byte("test content for sha256")
	tmpFile, err := os.CreateTemp("", "probe-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	tmpFile.Write(content)
	tmpFile.Close()
	
	// Calculate correct SHA256
	hash := sha256.Sum256(content)
	correctSHA256 := hex.EncodeToString(hash[:])
	
	task := &DownloadExecTask{
		SHA256: correctSHA256,
	}
	
	// Test with correct SHA256
	err = task.verifySHA256(tmpFile.Name())
	if err != nil {
		t.Errorf("Verification failed with correct SHA256: %v", err)
	}
	
	// Test with incorrect SHA256
	task.SHA256 = "invalid_hash"
	err = task.verifySHA256(tmpFile.Name())
	if err == nil {
		t.Errorf("Expected error with incorrect SHA256")
	}
}

func TestDownloadExecTaskSignatureVerification(t *testing.T) {
	// Generate Ed25519 key pair
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}
	
	// Create temp file with known content
	content := []byte("test content for signature")
	tmpFile, err := os.CreateTemp("", "probe-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	tmpFile.Write(content)
	tmpFile.Close()
	
	// Sign the content
	signature := ed25519.Sign(privateKey, content)
	
	task := &DownloadExecTask{
		Signature: base64.StdEncoding.EncodeToString(signature),
		PublicKey: base64.StdEncoding.EncodeToString(publicKey),
	}
	
	// Test with correct signature
	err = task.verifySignature(tmpFile.Name())
	if err != nil {
		t.Errorf("Verification failed with correct signature: %v", err)
	}
	
	// Test with incorrect signature
	task.Signature = base64.StdEncoding.EncodeToString([]byte("invalid_signature_12345678901234567890123456789012345678901234567890123456"))
	err = task.verifySignature(tmpFile.Name())
	if err == nil {
		t.Errorf("Expected error with incorrect signature")
	}
}
