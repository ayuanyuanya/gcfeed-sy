package infraconfig

import (
	"errors"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
)

// LoadConfig 读取 YAML 配置文件，并反序列化为应用启动配置。
func LoadConfig(path string) (*Config, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, errors.New("config file path is empty")
	}
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New("config file read error")
	}
	cfg := &Config{}
	err = yaml.Unmarshal(file, cfg)
	if err != nil {
		return nil, errors.New("config file unmarshal error")
	}
	return cfg, nil
}
