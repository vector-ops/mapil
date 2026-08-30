package helpers

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v4"
)

const DefaultDataDir = ".mapil"

type Config struct {
	DataDir   string              `yaml:"data_dir"`
	Databases map[string]DBConfig `yaml:"databases"`
	WriteBack bool                `yaml:"-"`
}

type DBConfig struct {
	URL         string      `yaml:"url,omitempty"`
	Driver      string      `yaml:"driver"`
	Filename    string      `yaml:"filename,omitempty"`
	Remote      bool        `yaml:"remote,omitempty"`
	Primary     bool        `yaml:"primary"`
	Host        string      `yaml:"host,omitempty"`
	Port        string      `yaml:"port,omitempty"`
	Credentials Credentials `yaml:"credentials,omitempty"`
}

type Credentials struct {
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

func ParseConfig(fp string) Config {

	f, err := os.OpenFile(fp, os.O_RDONLY, os.ModePerm)
	if err != nil {
		fmt.Println("failed to open config file")
		os.Exit(1)
	}
	defer f.Close()

	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		if errors.Is(err, io.EOF) {
			return cfg
		}
		fmt.Println("failed to decode config")
		os.Exit(1)
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

	if !atleastOnePrimary && len(cfg.Databases) > 0 {
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

	return DBConfig{Driver: "file", Primary: true}
}

func (c Config) LoadDefault() Config {
	if c.DataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}

		c.DataDir = filepath.Join(home, DefaultDataDir)
		c.WriteBack = true
	}

	if len(c.Databases) == 0 {
		c.Databases = make(map[string]DBConfig)
		c.Databases["file"] = c.PrimaryDB().LoadDefault()
		c.WriteBack = true
	}

	return c
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
