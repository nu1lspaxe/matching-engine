package keys

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGenerateEd25519KeyPair(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		setup   func() string
		cleanup func(string)
	}{
		{
			name: "successful_key_generation",
			args: args{
				filename: "",
			},
			wantErr: false,
			setup: func() string {
				tempDir := os.TempDir()
				return filepath.Join(tempDir, "test_key_success")
			},
			cleanup: func(filename string) {
				os.Remove(filename)
				os.Remove(filename + ".pub")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := tt.setup()
			if filename == "" && tt.args.filename != "" {
				filename = tt.args.filename
			}

			defer tt.cleanup(filename)

			err := GenerateEd25519KeyPair(filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateEd25519KeyPair() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				t.Run("validate_files_created", func(t *testing.T) {
					validateFilesCreated(t, filename)
				})
				t.Run("validate_file_permissions", func(t *testing.T) {
					validateFilePermissions(t, filename)
				})
				t.Run("validate_key_formats", func(t *testing.T) {
					validateKeyFormats(t, filename)
				})
				t.Run("validate_key_pair_relationship", func(t *testing.T) {
					validateKeyPairRelationship(t, filename)
				})
				t.Run("validate_cryptographic_operations", func(t *testing.T) {
					validateCryptographicOperations(t, filename)
				})
			}
		})
	}
}

func validateFilesCreated(t *testing.T, filename string) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("Private key file %s was not created", filename)
	}

	pubFilename := filename + ".pub"
	if _, err := os.Stat(pubFilename); os.IsNotExist(err) {
		t.Errorf("Public key file %s was not created", pubFilename)
	}
}

func validateFilePermissions(t *testing.T, filename string) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission tests on Windows")
	}

	privInfo, err := os.Stat(filename)
	if err != nil {
		t.Fatalf("Failed to get private key file info: %v", err)
	}
	if privInfo.Mode().Perm() != 0600 {
		t.Errorf("Private key file has incorrect permissions: got %o, want 0600", privInfo.Mode().Perm())
	}

	pubInfo, err := os.Stat(filename + ".pub")
	if err != nil {
		t.Fatalf("Failed to get public key file info: %v", err)
	}
	if pubInfo.Mode().Perm() != 0644 {
		t.Errorf("Public key file has incorrect permissions: got %o, want 0644", pubInfo.Mode().Perm())
	}
}

func validateKeyFormats(t *testing.T, filename string) {
	privKeyData, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read private key file: %v", err)
	}

	privBlock, _ := pem.Decode(privKeyData)
	if privBlock == nil {
		t.Fatal("Failed to decode private key PEM block")
	}

	if privBlock.Type != "PRIVATE KEY" {
		t.Errorf("Private key PEM block has wrong type: got %s, want PRIVATE KEY", privBlock.Type)
	}

	privKey, err := x509.ParsePKCS8PrivateKey(privBlock.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}

	if _, ok := privKey.(ed25519.PrivateKey); !ok {
		t.Errorf("Private key is not Ed25519 type")
	}

	pubKeyData, err := os.ReadFile(filename + ".pub")
	if err != nil {
		t.Fatalf("Failed to read public key file: %v", err)
	}

	pubBlock, _ := pem.Decode(pubKeyData)
	if pubBlock == nil {
		t.Fatal("Failed to decode public key PEM block")
	}

	if pubBlock.Type != "PUBLIC KEY" {
		t.Errorf("Public key PEM block has wrong type: got %s, want PUBLIC KEY", pubBlock.Type)
	}

	pubKey, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse public key: %v", err)
	}

	if _, ok := pubKey.(ed25519.PublicKey); !ok {
		t.Errorf("Public key is not Ed25519 type")
	}
}

func validateKeyPairRelationship(t *testing.T, filename string) {
	privKeyData, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read private key: %v", err)
	}

	privBlock, _ := pem.Decode(privKeyData)
	if privBlock == nil {
		t.Fatal("Failed to decode private key PEM")
	}

	privKey, err := x509.ParsePKCS8PrivateKey(privBlock.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}

	ed25519PrivKey := privKey.(ed25519.PrivateKey)

	pubKeyData, err := os.ReadFile(filename + ".pub")
	if err != nil {
		t.Fatalf("Failed to read public key: %v", err)
	}

	pubBlock, _ := pem.Decode(pubKeyData)
	if pubBlock == nil {
		t.Fatal("Failed to decode public key PEM")
	}

	pubKey, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse public key: %v", err)
	}

	ed25519PubKey := pubKey.(ed25519.PublicKey)

	expectedPubKey := ed25519PrivKey.Public().(ed25519.PublicKey)
	if !ed25519PubKey.Equal(expectedPubKey) {
		t.Errorf("Generated public key does not match the public key derived from private key")
	}
}

func validateCryptographicOperations(t *testing.T, filename string) {
	privKeyData, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read private key: %v", err)
	}

	privBlock, _ := pem.Decode(privKeyData)
	privKey, err := x509.ParsePKCS8PrivateKey(privBlock.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}

	ed25519PrivKey := privKey.(ed25519.PrivateKey)

	pubKeyData, err := os.ReadFile(filename + ".pub")
	if err != nil {
		t.Fatalf("Failed to read public key: %v", err)
	}

	pubBlock, _ := pem.Decode(pubKeyData)
	pubKey, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse public key: %v", err)
	}

	ed25519PubKey := pubKey.(ed25519.PublicKey)

	testMessages := [][]byte{
		[]byte("Hello, World!"),
		[]byte(""),
		[]byte("這是一個測試訊息"),
		make([]byte, 1024),
	}

	for i, message := range testMessages {
		t.Run(fmt.Sprintf("sign_verify_message_%d", i), func(t *testing.T) {
			signature := ed25519.Sign(ed25519PrivKey, message)

			if !ed25519.Verify(ed25519PubKey, message, signature) {
				t.Errorf("Signature verification failed for message %d", i)
			}

			wrongMessage := append(message, byte(0xFF))
			if ed25519.Verify(ed25519PubKey, wrongMessage, signature) {
				t.Errorf("Signature verification should have failed for wrong message %d", i)
			}
		})
	}
}
