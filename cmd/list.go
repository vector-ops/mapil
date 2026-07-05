package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vector-ops/mapil/database"
)

const (
	Underline = "\033[4m"
	Reset     = "\033[0m"
	DarkGrey  = "\033[90m"
)

var lsns *string

func init() {
	lsns = listCmd.PersistentFlags().StringP("namespace", "s", "", "list objects from namespace.\n if empty lists all objects")

}

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all objects",
	Long:    `All objects stored are listed`,
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		data := []database.ListType{}

		if *lsns != "" {
			data = dataStore.GetNamespaceObjects(cmd.Context(), *lsns)
		} else {
			data = dataStore.GetAllData(cmd.Context())
		}

		if len(data) == 0 {
			fmt.Println("Data store empty.")
		} else {
			for i, do := range data {
				fmt.Printf("  %s%s%s %s[%d]%s\n", Underline, do.Key, Reset, DarkGrey, len(do.Value), Reset)
				for i, v := range do.Value {
					fmt.Printf("   %s%d.%s %s\n", DarkGrey, i+1, Reset, v)
				}

				if i < len(data)-1 {
					fmt.Println()
				}
			}
		}
	},
}
