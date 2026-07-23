package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/vector-ops/mapil/cmd"
	"github.com/vector-ops/mapil/helpers"
	"github.com/vector-ops/mapil/store"
)

var devMode string

func main() {
	dev := devMode == "true"

	ctx, cancel := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
	defer cancel()

	cfg := helpers.ParseConfig("config.yaml")
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
