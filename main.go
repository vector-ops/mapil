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

	userCfgDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("could not open user config directory")
		return
	}

	if dev {
		userCfgDir = os.TempDir()
		fmt.Printf("Using temp directory for development, path: %s\n", filepath.Join(userCfgDir, MplCfgDir, CfgFile))
	}

	cfgPath := filepath.Join(userCfgDir, MplCfgDir, CfgFile)

	createConfig(cfgPath)
	cfg := helpers.ParseConfig(cfgPath)
	if err := helpers.ValidateConfig(cfg); err != nil {
		fmt.Println(err)
		return
	}
	cfg = cfg.LoadDefault()

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

func createConfig(p string) error {
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
