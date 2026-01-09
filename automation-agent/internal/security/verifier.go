package security

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"os"
)

// Verifier verifies Ed25519 signatures
type Verifier struct {
	publicKeys map[string]ed25519.PublicKey
}

// NewVerifier creates a new signature verifier
func NewVerifier() *Verifier {
	return &Verifier{
		publicKeys: make(map[string]ed25519.PublicKey),
	}
}

// AddPublicKey adds a public key for verification
func (v *Verifier) AddPublicKey(keyID string, publicKey ed25519.PublicKey) {
	v.publicKeys[keyID] = publicKey
}

// VerifySignature verifies an Ed25519 signature
func (v *Verifier) VerifySignature(filePath, signature, keyID string) error {
	publicKey, ok := v.publicKeys[keyID]
	if !ok {
		return fmt.Errorf("public key not found for key_id: %s", keyID)
	}
	
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	// Decode signature
	sigBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}
	
	// Verify
	if !ed25519.Verify(publicKey, data, sigBytes) {
		return fmt.Errorf("signature verification failed")
	}
	
	return nil
}
