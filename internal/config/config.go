package config

import (
	"gopkg.in/yaml.v2"
	"os"
	"time"
)

type Config struct {
	DBConnect  string        `yaml:"DB_CONNECT"`
	Cancel     time.Duration `yaml:"cancel"`
	UserToken  string        `yaml:"userToken"`
	AdminToken string        `yaml:"adminToken"`
	Port       string        `yaml:"port"`
	Redis      string        `yaml:"redisConn"`
}

func LoadConfig(path string) (*Config, error) {

	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	err = yaml.Unmarshal(yamlFile, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, err

}
