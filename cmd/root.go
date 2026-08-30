package cmd

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
	"github.com/vector-ops/mapil/store"
)

var (
	dataStore *store.Store
	info      debug.BuildInfo
	Version   string

	rootCmd = &cobra.Command{
		Use:   "mapil",
		Short: "Mapil is used to store and access lists from CLI.",
		Long:  `Mapil is a CLI based tool to store and view lists on the command line. It allows you to create different lists on the command line and store api keys, bookmarks, todo lists etc.`,
		Run: func(cmd *cobra.Command, args []string) {
		},

		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			err := dataStore.Close(cmd.Context())
			if err != nil {
				fmt.Println(err)
				return
			}
		},
	}
)

func Execute(ctx context.Context, st *store.Store) {
	dataStore = st
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	setVersion()
	rootCmd.Version = info.Main.Version
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(delCmd)
	rootCmd.AddCommand(updCmd)
	rootCmd.AddCommand(apdCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(popCmd)
	rootCmd.AddCommand(rmCmd)

}

func setVersion() {
	if i, ok := debug.ReadBuildInfo(); ok {
		if i.Main.Version != "(devel)" {
			info = *i
		} else {
			info = debug.BuildInfo{
				Main: debug.Module{
					Version: Version,
				},
			}
		}
	}
}
