package main

import (
	"github.com/vector-ops/mapil/cmd"
	"github.com/vector-ops/mapil/store"
)

var devMode string

func main() {
	dev := devMode == "true"

	store := store.NewStore(dev)
	store.Init()

	cmd.Execute(store)
}
