package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

const ConfigFilePath = "config.json"

const (
	VisionIP  = "224.5.23.2"
	TrackedIP = "224.5.23.2"
	RefereeIP = "224.5.23.1"
)

// Config アプリケーション設定
type Config struct {
	VisionPort  int `json:"visionPort"`
	TrackedPort int `json:"trackedPort"`
	RefereePort int `json:"refereePort"`
	mu          sync.RWMutex
}

// DefaultConfig デフォルト設定を返す
func DefaultConfig() *Config {
	return &Config{
		VisionPort:  10006,
		TrackedPort: 10010,
		RefereePort: 10003,
	}
}

// LoadConfig 設定ファイルから設定を読み込む
func LoadConfig() (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(ConfigFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// ファイルが存在しない場合はデフォルト設定を使用
			return cfg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save 設定をファイルに保存
func (c *Config) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(ConfigFilePath, data, 0644)
}

// Update 設定を更新
func (c *Config) Update(visionPort, trackedPort, refereePort int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if visionPort > 0 {
		c.VisionPort = visionPort
	}
	if trackedPort > 0 {
		c.TrackedPort = trackedPort
	}
	if refereePort > 0 {
		c.RefereePort = refereePort
	}
}

// GetAddresses 現在のアドレス設定を取得
func (c *Config) GetAddresses() (vision, tracked, referee string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return fmt.Sprintf("%s:%d", VisionIP, c.VisionPort),
		fmt.Sprintf("%s:%d", TrackedIP, c.TrackedPort),
		fmt.Sprintf("%s:%d", RefereeIP, c.RefereePort)
}

// GetPorts 現在のポート設定を取得
func (c *Config) GetPorts() (vision, tracked, referee int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.VisionPort, c.TrackedPort, c.RefereePort
}
