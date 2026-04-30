package service

import (
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v2"
)

type Config struct {
	Network  NetworkConfig  `yaml:"network"`
	Contract ContractConfig `yaml:"contract"`
}

type NetworkConfig struct {
	RPCURL  string `yaml:"rpc_url"`
	ChainID int64  `yaml:"chain_id"`
}

type ContractConfig struct {
	Address string `yaml:"address"`
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
