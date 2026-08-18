package crypto

import "testing"

func TestEncryptDecryptRoundTrip(t *testing.T) {
	c, err := NewCipher("test-master-key")
	if err != nil {
		t.Fatalf("NewCipher: %v", err)
	}
	plain := "ovs-api-key-中文-123"
	enc, err := c.Encrypt(plain)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if enc == plain || !IsCipherText(enc) {
		t.Fatalf("unexpected cipher text: %q", enc)
	}
	got, err := c.Decrypt(enc)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if got != plain {
		t.Fatalf("round trip mismatch: got %q want %q", got, plain)
	}
}

func TestEncryptNondeterministic(t *testing.T) {
	c, _ := NewCipher("k")
	a, _ := c.Encrypt("same")
	b, _ := c.Encrypt("same")
	if a == b {
		t.Fatal("cipher text should differ across calls (random nonce)")
	}
}

func TestDecryptPlainPassthrough(t *testing.T) {
	c, _ := NewCipher("k")
	got, err := c.Decrypt("legacy-plain-value")
	if err != nil || got != "legacy-plain-value" {
		t.Fatalf("passthrough failed: %q %v", got, err)
	}
}

func TestDecryptWrongKey(t *testing.T) {
	c1, _ := NewCipher("k1")
	c2, _ := NewCipher("k2")
	enc, _ := c1.Encrypt("secret")
	if _, err := c2.Decrypt(enc); err == nil {
		t.Fatal("expected error decrypting with wrong key")
	}
}

func TestEmptyMasterKey(t *testing.T) {
	if _, err := NewCipher("   "); err == nil {
		t.Fatal("expected ErrEmptyMasterKey")
	}
}
