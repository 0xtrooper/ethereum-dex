// Copyright © 2026 0xTrooper (on Github)
// 
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Copyright © 2026 0xTrooper (on Github)
// 
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package service

import (
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v2"
)

type Config struct {
	Network  NetworkConfig  `yaml:"network"`
	Contract ContractConfig `yaml:"contract"`
	Tokens   []TokenConfig  `yaml:"tokens,omitempty"`
}

type NetworkConfig struct {
	RPCURL  string `yaml:"rpc_url"`
	ChainID int64  `yaml:"chain_id"`
}

type ContractConfig struct {
	Address string `yaml:"address"`
}

type TokenConfig struct {
	Symbol   string `yaml:"symbol,omitempty"`
	Address  string `yaml:"address"`
	Decimals uint8  `yaml:"decimals,omitempty"`
}

type Service struct {
	config *Config
}

func New() (*Service, error) {
	config, err := lazyLoad()
	if err != nil {
		return nil, err
	}
	return &Service{config: config}, nil
}

func (s *Service) Get() *Config {
	return s.config
}

func (s *Service) Path() string {
	path, _ := configPath()
	return path
}

func (s *Service) Delete() error {
	path, err := configPath()
	if err != nil {
		return err
	}
	return os.Remove(path)
}

func (s *Service) LoadIfExists() (*Config, bool, error) {
	path, err := configPath()
	if err != nil {
		return nil, false, err
	}
	return loadIfExists(path)
}

func (s *Service) Ensure() (*Config, bool, error) {
	path, err := configPath()
	if err != nil {
		return nil, false, err
	}

	cfg, exists, err := loadIfExists(path)
	if err != nil {
		return nil, false, err
	}

	if !exists {
		cfg = defaultConfig()
		if err = save(path, cfg); err != nil {
			return nil, false, err
		}
		return cfg, true, nil
	}

	return cfg, false, nil
}

func (s *Service) Save(c *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	return save(path, c)
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "dex", "config.yaml"), nil
}

func load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err = yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func loadIfExists(path string) (*Config, bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	config, err := load(path)
	if err != nil {
		return nil, false, err
	}
	return config, true, nil
}

func defaultConfig() *Config {
	return &Config{
		Network: NetworkConfig{
			RPCURL:  "",
			ChainID: defaultChainID,
		},
		Contract: ContractConfig{
			Address: defaultContractAddress,
		},
		Tokens: nil,
	}
}

func save(path string, c *Config) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func lazyLoad() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	config, exists, err := loadIfExists(path)
	if err != nil {
		return nil, err
	}

	if !exists {
		config = defaultConfig()
		if err = save(path, config); err != nil {
			return nil, err
		}
	}

	return config, nil
}
