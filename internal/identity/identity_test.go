package identity

import (
	"crypto/ed25519"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestKeypairGenerateAndPEM(t *testing.T) {
	kp, err := GenerateKeypair()
	if err != nil {
		t.Fatalf("failed to generate keypair: %v", err)
	}

	didStr := kp.DID()
	if !strings.HasPrefix(didStr, "did:key:z6Mk") {
		t.Fatalf("expected DID starting with did:key:z6Mk, got %s", didStr)
	}

	pemBytes, err := kp.ToPEM()
	if err != nil {
		t.Fatalf("failed to serialize to PEM: %v", err)
	}

	recovered, err := FromPEM(pemBytes)
	if err != nil {
		t.Fatalf("failed to parse PEM: %v", err)
	}

	if !kp.PrivateKey.Equal(recovered.PrivateKey) {
		t.Fatalf("recovered key does not match original")
	}

	msg := []byte("test message for signing")
	sig := kp.Sign(msg)
	if !Verify(kp.PublicKey, msg, sig) {
		t.Fatalf("signature verification failed")
	}

	b64Sig := kp.SignB64(msg)
	if len(b64Sig) == 0 {
		t.Fatalf("empty base64 signature")
	}
}

func TestSaveAndLoadKeypair(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "twig-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	kp, err := GenerateKeypair()
	if err != nil {
		t.Fatalf("failed to generate keypair: %v", err)
	}

	if err := SaveKeypair(tempDir, kp); err != nil {
		t.Fatalf("failed to save keypair: %v", err)
	}

	loaded, err := LoadKeypair(tempDir)
	if err != nil {
		t.Fatalf("failed to load keypair: %v", err)
	}

	if loaded.DID() != kp.DID() {
		t.Fatalf("loaded DID %s != original DID %s", loaded.DID(), kp.DID())
	}

	// Verify key file exists
	if _, err := os.Stat(filepath.Join(tempDir, "identity.pem")); err != nil {
		t.Fatalf("identity.pem file does not exist: %v", err)
	}
}

func TestFromSeed(t *testing.T) {
	seed := make([]byte, ed25519.SeedSize)
	for i := range seed {
		seed[i] = byte(i)
	}

	kp1, err := FromSeed(seed)
	if err != nil {
		t.Fatalf("failed to create keypair from seed: %v", err)
	}

	kp2, err := FromSeed(seed)
	if err != nil {
		t.Fatalf("failed to create keypair from seed: %v", err)
	}

	if kp1.DID() != kp2.DID() {
		t.Fatalf("deterministic DIDs do not match: %s != %s", kp1.DID(), kp2.DID())
	}
}
