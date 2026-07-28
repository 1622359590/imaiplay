package sms

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

type persistedConfig struct {
	Config
	AccessKeySecretEncrypted string `json:"access_key_secret_encrypted,omitempty"`
}

type ConfigStore struct {
	mu     sync.Mutex
	path   string
	key    []byte
	config Config
	sender *ConfigurableSender
}

func NewConfigStore(path, encryptionSecret string, logger *slog.Logger) (*ConfigStore, error) {
	store := &ConfigStore{path: path, key: keyFromSecret(encryptionSecret), sender: NewConfigurableSender(Config{}, logger)}
	if err := store.load(); err != nil {
		return nil, err
	}
	store.sender.Update(store.config)
	return store, nil
}

func (store *ConfigStore) Sender() *ConfigurableSender { return store.sender }

func (store *ConfigStore) Get() Config {
	store.mu.Lock()
	defer store.mu.Unlock()
	config := store.config
	config.AccessKeySecret = ""
	return config
}

func (store *ConfigStore) Save(config Config) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	if config.AccessKeySecret == "" {
		config.AccessKeySecret = store.config.AccessKeySecret
	}
	if err := store.write(config); err != nil {
		return err
	}
	store.config = config
	store.sender.Update(config)
	return nil
}

func (store *ConfigStore) SendTest(ctx context.Context, phone string) error {
	return store.sender.Send(ctx, phone, "", map[string]string{"code": "123456"})
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
	var persisted persistedConfig
	if err := json.Unmarshal(data, &persisted); err != nil {
		return err
	}
	secret, err := decrypt(store.key, persisted.AccessKeySecretEncrypted)
	if err != nil {
		return err
	}
	persisted.Config.AccessKeySecret = secret
	store.config = persisted.Config
	return nil
}

func (store *ConfigStore) write(config Config) error {
	secret, err := encrypt(store.key, config.AccessKeySecret)
	if err != nil {
		return err
	}
	persisted := persistedConfig{Config: config, AccessKeySecretEncrypted: secret}
	persisted.Config.AccessKeySecret = ""
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

func keyFromSecret(secret string) []byte { hash := sha256.Sum256([]byte(secret)); return hash[:] }

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
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
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
		return "", errors.New("invalid encrypted sms secret")
	}
	plain, err := gcm.Open(nil, data[:gcm.NonceSize()], data[gcm.NonceSize():], nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
