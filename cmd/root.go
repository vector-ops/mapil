package cmd

import (
	"fmt"
	"os"
	"path"
	"runtime/debug"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vector-ops/mapil/store"
)

var (
	DataStore *store.Store
	info      debug.BuildInfo
	Version   string

	delAll *bool

	rootCmd = &cobra.Command{
		Use:     "mapil",
		Aliases: []string{"mpl"},
		Short:   "Mapil is used to store and access lists from CLI.",
		Long:    `Mapil is a CLI based tool to store and view lists on the command line. It allows you to create different lists on the command line and store api keys, bookmarks, todo lists etc.`,
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
)

func Execute(st *store.Store) {
	DataStore = st
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	setVersion()
	cobra.OnInitialize(initConfig)
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

	// Delete Flags
	delAll = delCmd.PersistentFlags().BoolP("all", "a", false, "delete all the data in the data store.")
}

func initConfig() {
	home, err := os.UserConfigDir()
	cobra.CheckErr(err)
	cfgFile := path.Join(home, "mapil", "config.yaml")
	cfgDir := path.Join(home, "mapil")
	createCfgFile(cfgDir)
	viper.SetConfigFile(cfgFile)

	viper.AddConfigPath(cfgDir)
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Error reading config file")
	}
}

func createCfgFile(cfgDir string) {
	f, err := os.OpenFile(path.Join(cfgDir, "config.yaml"), os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer f.Close()
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
