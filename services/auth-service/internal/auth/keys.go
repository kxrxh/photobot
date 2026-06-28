package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sync"

	"csort.ru/auth-service/internal/logger"
	jwks "github.com/intility/go-jwks"
)

var (
	keyManager     *KeyManager
	keyManagerOnce sync.Once
	keyLogger      = logger.GetLogger("auth.keys")
)

// KeyManager handles RSA key pair generation, loading, and JWKS generation
type KeyManager struct {
	privateKey   *rsa.PrivateKey
	publicKey    *rsa.PublicKey
	keyID        string
	mu           sync.RWMutex
	initialized  bool
	jwksJSON     []byte
	jwksJSONOnce sync.Once
}

// GetKeyManager returns the singleton KeyManager instance
func GetKeyManager() *KeyManager {
	keyManagerOnce.Do(func() {
		keyManager = &KeyManager{}
	})
	return keyManager
}

// Initialize loads or generates RSA keys.
func (km *KeyManager) Initialize(privateKeyPath, publicKeyPath string) error {
	km.mu.Lock()
	defer km.mu.Unlock()

	if km.initialized {
		return errors.New("key manager already initialized")
	}

	if err := km.loadKeys(privateKeyPath); err == nil {
		keyLogger.Info().Msg("Loaded existing RSA keys")
		km.initialized = true
		return nil
	}

	keyLogger.Info().Msg("Generating new RSA key pair")
	if err := km.generateKeys(); err != nil {
		return fmt.Errorf("failed to generate keys: %w", err)
	}

	if err := km.saveKeys(privateKeyPath, publicKeyPath); err != nil {
		keyLogger.Warn().
			Err(err).
			Msg("Failed to save keys to files, but continuing with in-memory keys")
	}

	km.initialized = true
	return nil
}

func (km *KeyManager) loadKeys(privateKeyPath string) error {
	// Load private key (path from config, not user input)
	privateKeyData, err := os.ReadFile(privateKeyPath) // #nosec G304 -- path from config
	if err != nil {
		return fmt.Errorf("failed to read private key file: %w", err)
	}

	block, rest := pem.Decode(privateKeyData)
	if block == nil {
		return errors.New("failed to decode private key PEM")
	}
	if len(rest) > 0 {
		keyLogger.Warn().Msg("Extra data found after PEM block, ignoring")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return errors.New("private key is not RSA")
		}
	}

	km.privateKey = privateKey
	km.publicKey = &privateKey.PublicKey

	keyID, err := km.generateKeyID()
	if err != nil {
		return fmt.Errorf("failed to generate key ID: %w", err)
	}
	km.keyID = keyID

	return nil
}

func (km *KeyManager) generateKeys() error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate RSA key: %w", err)
	}

	km.privateKey = privateKey
	km.publicKey = &privateKey.PublicKey

	keyID, err := km.generateKeyID()
	if err != nil {
		return fmt.Errorf("failed to generate key ID: %w", err)
	}
	km.keyID = keyID

	return nil
}

func (km *KeyManager) saveKeys(privateKeyPath, publicKeyPath string) error {
	keysDir := filepath.Dir(privateKeyPath)
	if err := os.MkdirAll(keysDir, 0o750); err != nil {
		return fmt.Errorf("failed to create keys directory: %w", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(km.privateKey),
	})

	if err := os.WriteFile(privateKeyPath, privateKeyPEM, 0o600); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	publicKeyDER, err := x509.MarshalPKIXPublicKey(km.publicKey)
	if err != nil {
		return fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyDER,
	})

	if err := os.WriteFile(
		publicKeyPath,
		publicKeyPEM,
		0o644, // #nosec G306 -- public key is meant to be readable
	); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}

	return nil
}

func (km *KeyManager) generateKeyID() (string, error) {
	publicKeyDER, err := x509.MarshalPKIXPublicKey(km.publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key for key ID generation: %w", err)
	}

	hash := sha256Hash(publicKeyDER)
	return base64.RawURLEncoding.EncodeToString(hash[:8]), nil
}

// GetSigningKey returns the private key for signing tokens
func (km *KeyManager) GetSigningKey() *rsa.PrivateKey {
	km.mu.RLock()
	defer km.mu.RUnlock()
	if !km.initialized {
		return nil
	}
	return km.privateKey
}

// GetPublicKey returns the public key for validation
func (km *KeyManager) GetPublicKey() *rsa.PublicKey {
	km.mu.RLock()
	defer km.mu.RUnlock()
	if !km.initialized {
		return nil
	}
	return km.publicKey
}

// GetKeyID returns the active JWK key id (kid).
func (km *KeyManager) GetKeyID() string {
	km.mu.RLock()
	defer km.mu.RUnlock()
	if !km.initialized {
		return ""
	}
	return km.keyID
}

// GetJWKS returns the JSON Web Key Set for this key
func (km *KeyManager) GetJWKS() *jwks.JWKS {
	km.mu.RLock()
	defer km.mu.RUnlock()

	if km.publicKey == nil {
		return nil
	}

	// JWK RSA fields: unsigned big-endian modulus and exponent.
	nBytes := km.publicKey.N.Bytes()
	n := base64.RawURLEncoding.EncodeToString(nBytes)

	eBytes := bigIntToBytes(km.publicKey.E)
	e := base64.RawURLEncoding.EncodeToString(eBytes)

	return &jwks.JWKS{
		Keys: []jwks.JSONWebKey{
			{
				Kty: "RSA",
				Use: "sig",
				Kid: km.keyID,
				Alg: "RS256",
				N:   n,
				E:   e,
			},
		},
	}
}

// GetJWKSJSON returns cached pre-serialized JWKS JSON.
// Call after Initialize. Returns nil, nil if keys are not initialized.
func (km *KeyManager) GetJWKSJSON() ([]byte, error) {
	km.mu.RLock()
	if !km.initialized || km.publicKey == nil {
		km.mu.RUnlock()
		return nil, nil
	}

	// Check if already cached without holding write lock
	if km.jwksJSON != nil {
		km.mu.RUnlock()
		return km.jwksJSON, nil
	}
	km.mu.RUnlock()

	km.mu.Lock()
	defer km.mu.Unlock()

	// Double-check after acquiring write lock
	if !km.initialized || km.publicKey == nil {
		return nil, nil
	}

	km.jwksJSONOnce.Do(func() {
		// Generate JWKS directly without calling GetJWKS() to avoid deadlock
		nBytes := km.publicKey.N.Bytes()
		n := base64.RawURLEncoding.EncodeToString(nBytes)
		eBytes := bigIntToBytes(km.publicKey.E)
		e := base64.RawURLEncoding.EncodeToString(eBytes)

		jwksObj := &jwks.JWKS{
			Keys: []jwks.JSONWebKey{
				{
					Kty: "RSA",
					Use: "sig",
					Kid: km.keyID,
					Alg: "RS256",
					N:   n,
					E:   e,
				},
			},
		}
		km.jwksJSON, _ = json.Marshal(jwksObj)
	})
	return km.jwksJSON, nil
}

func sha256Hash(data []byte) [32]byte {
	hash := sha256.Sum256(data)
	return hash
}

func bigIntToBytes(e int) []byte {
	bigE := big.NewInt(int64(e))
	return bigE.Bytes()
}
