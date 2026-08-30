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
)

var devMode string

const CfgFile = "config.yaml"

func main() {
	dev := devMode == "true"

	ctx, cancel := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
	defer cancel()

	cfgDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("could not open user config directory")
		return
	}

	if dev {
		cfgDir = os.TempDir()
		fmt.Printf("Using temp directory for development, path: %s\n", filepath.Join(cfgDir, CfgFile))
	}

	cfgPath := filepath.Join(cfgDir, CfgFile)

	createConfig(cfgPath)
	cfg := helpers.ParseConfig(cfgPath)
	if err := helpers.ValidateConfig(cfg); err != nil {
		fmt.Println(err)
		return
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
