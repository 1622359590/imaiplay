package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type persistedStorageConfig struct {
	Config          `json:",inline"`
	SecretEncrypted string `json:"secret_encrypted,omitempty"`
}

type ConfigStore struct {
	mu     sync.Mutex
	path   string
	key    []byte
	config Config
}

func NewConfigStore(file, secret string, defaults ...Config) (*ConfigStore, error) {
	store := &ConfigStore{path: file, key: keyFromSecret(secret)}
	if len(defaults) > 0 {
		store.config = defaults[0]
	}
	return store, store.load()
}
func (store *ConfigStore) internal() Config {
	store.mu.Lock()
	defer store.mu.Unlock()
	return store.config
}
func (store *ConfigStore) Get() Config {
	store.mu.Lock()
	defer store.mu.Unlock()
	config := store.config
	config.S3.SecretKey = ""
	return config
}
func (store *ConfigStore) Save(config Config) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	if config.S3.SecretKey == "" {
		config.S3.SecretKey = store.config.S3.SecretKey
	}
	if err := store.write(config); err != nil {
		return err
	}
	store.config = config
	return nil
}
func (store *ConfigStore) load() error {
	store.mu.Lock()
	defer store.mu.Unlock()
	data, err := os.ReadFile(store.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var persisted persistedStorageConfig
	if err := json.Unmarshal(data, &persisted); err != nil {
		return err
	}
	secret, err := decrypt(store.key, persisted.SecretEncrypted)
	if err != nil {
		return err
	}
	persisted.Config.S3.SecretKey = secret
	store.config = persisted.Config
	return nil
}
func (store *ConfigStore) write(config Config) error {
	encrypted, err := encrypt(store.key, config.S3.SecretKey)
	if err != nil {
		return err
	}
	persisted := persistedStorageConfig{Config: config, SecretEncrypted: encrypted}
	persisted.Config.S3.SecretKey = ""
	data, err := json.MarshalIndent(persisted, "", "  ")
	if err != nil {
		return err
	}
	if dir := filepath.Dir(store.path); dir != "." {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return err
		}
	}
	temporary := store.path + ".tmp"
	if err := os.WriteFile(temporary, data, 0o600); err != nil {
		return err
	}
	return os.Rename(temporary, store.path)
}
func keyFromSecret(secret string) []byte { sum := sha256.Sum256([]byte(secret)); return sum[:] }
func encrypt(key []byte, value string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(gcm.Seal(nonce, nonce, []byte(value), nil)), nil
}
func decrypt(key []byte, value string) (string, error) {
	if value == "" {
		return "", nil
	}
	data, err := base64.RawStdEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(data) < gcm.NonceSize() {
		return "", errors.New("invalid encrypted storage secret")
	}
	plain, err := gcm.Open(nil, data[:gcm.NonceSize()], data[gcm.NonceSize():], nil)
	return string(plain), err
}
