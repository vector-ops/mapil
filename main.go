package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/vector-ops/mapil/cmd"
	"github.com/vector-ops/mapil/helpers"
	"github.com/vector-ops/mapil/store"
	"go.yaml.in/yaml/v4"
)

var devMode string

const CfgFile = "config.yaml"
const MplCfgDir = "mapil"

func main() {
	dev := devMode == "true"

	ctx, cancel := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
	defer cancel()

	cfgPath, dataDir, err := resolvePaths(dev)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if err := createConfigFile(cfgPath); err != nil {
		fmt.Println(err.Error())
		return
	}

	cfg := helpers.ParseConfig(cfgPath)
	if err := helpers.ValidateConfig(cfg); err != nil {
		fmt.Println(err)
		return
	}
	cfg = cfg.LoadDefault()

	if dev {
		cfg.DataDir = dataDir
	}

	if cfg.WriteBack {
		writeBackConfig(cfg, cfgPath)
	}

	store := store.NewStore(dev, cfg)
	if err := store.Init(ctx); err != nil {
		fmt.Println(err)
		return
	}

	cmd.Execute(ctx, store)
}

// createConfigFile creates the config file at path p
// if it does not exist already
func createConfigFile(p string) error {
	if helpers.PathExists(p) {
		return nil
	}

	return helpers.CreateFile(p)
}

// writeBackConfig writes the updated config back to the file.
// it returns a bool if it failed to write.
func writeBackConfig(cfg helpers.Config, fp string) bool {
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return false
	}

	return helpers.WriteToFile(b, fp) == nil
}

// resolvePaths returns config directory and data directory based on environment.
// If dev is true it assumes development environment else production environment.
// In case of production environment dataDir must be resolved from config.
//
// This function does not verify the existence of the paths.
func resolvePaths(dev bool) (cfgPath string, dataDir string, err error) {

	devDir := filepath.Join(os.TempDir(), "mapil")

	if dev {
		return filepath.Join(devDir, CfgFile), filepath.Join(devDir, "data"), nil
	}

	userCfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", "", fmt.Errorf("could not find user config directory")
	}

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("could not find user home directory")
	}

	cfgPath = filepath.Join(userCfgDir, MplCfgDir, CfgFile)
	dataDir = filepath.Join(userHomeDir, ".mapil")

	return
}
