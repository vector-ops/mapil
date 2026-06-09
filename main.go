package main

import (
	"fmt"

	"github.com/vector-ops/mapil/cmd"
	"github.com/vector-ops/mapil/store"
)

var devMode string

func main() {
	dev := devMode == "true"

	store := store.NewStore(dev)
	if err := store.Init(); err != nil {
		fmt.Println(err)
		return
	}

	cmd.Execute(store)
}
