package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/vector-ops/mapil/cmd"
	"github.com/vector-ops/mapil/store"
)

var devMode string

func main() {
	dev := devMode == "true"

	ctx, cancel := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
	defer cancel()

	store := store.NewStore(dev)
	if err := store.Init(ctx); err != nil {
		fmt.Println(err)
		return
	}

	cmd.Execute(ctx, store)
}
