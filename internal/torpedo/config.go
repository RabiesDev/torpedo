package torpedo

import (
	"encoding/json"
	"os"
	"torpedo/internal/helpers"
)

type Config struct {
	ServerAddress string `json:"server_address"`
	Proxies       string `json:"proxies"`
	PointOffset   int    `json:"point_offset"`
}

func (config *Config) ParseProxies() []string {
	proxies, err := helpers.ReadLines(config.Proxies)
	if err != nil {
		return nil
	}
	return proxies
}

func ParseConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err = json.Unmarshal(file, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
