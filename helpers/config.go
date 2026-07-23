package helpers

import (
	"errors"
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	DataDir   string              `yaml:"data_dir"`
	Databases map[string]DBConfig `yaml:"databases"`
}

type DBConfig struct {
	URL         string      `yaml:"url"`
	Driver      string      `yaml:"driver"`
	Filename    string      `yaml:"filename"`
	Remote      bool        `yaml:"remote"`
	Primary     bool        `yaml:"primary"`
	Host        string      `yaml:"host"`
	Port        string      `yaml:"port"`
	Credentials Credentials `yaml:"credentials"`
}

type Credentials struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func ParseConfig(fp string) Config {

	f, err := os.OpenFile(fp, os.O_RDONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		panic(err)
	}

	return cfg

}

func ValidateConfig(cfg Config) error {
	atleastOnePrimary := false
	for _, db := range cfg.Databases {
		atleastOnePrimary = db.Primary || atleastOnePrimary

		switch db.Driver {
		case "sqlite", "file":
			if db.Remote && db.URL == "" {
				return fmt.Errorf("url is required for remote %s database", db.Driver)
			}
		case "redis":
			return fmt.Errorf("unsupported database driver: %s", db.Driver)
		default:
			return fmt.Errorf("unsupported database driver: %s", db.Driver)
		}
	}

	if !atleastOnePrimary {
		return errors.New("atleast one primary database is required")
	}

	return nil
}

func (c Config) PrimaryDB() DBConfig {
	for _, db := range c.Databases {
		if db.Primary {
			return db
		}
	}

	return DBConfig{}
}

func (c DBConfig) LoadDefault() DBConfig {
	cfg := c

	switch cfg.Driver {
	case "file":
		if cfg.Filename == "" {
			cfg.Filename = "mapil.json"
		}
	case "sqlite":
		if cfg.Filename == "" {
			cfg.Filename = "mapil.db"
		}

	case "redis":
		if cfg.URL == "" {
			url := "redis://"

			if cfg.Credentials.Username != "" {
				pw := ""
				if cfg.Credentials.Password != "" {
					pw = ":" + cfg.Credentials.Password
				}
				url += cfg.Credentials.Username + pw
			}

			cfg.URL = fmt.Sprintf("%s@%s:%s", url, cfg.Host, cfg.Port)
		}
	}

	return cfg
}
