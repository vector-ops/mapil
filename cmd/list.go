package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	Underline = "\033[4m"
	Reset     = "\033[0m"
	DarkGrey  = "\033[90m"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all objects",
	Long:  `All objects stored are listed`,
	Run: func(cmd *cobra.Command, args []string) {
		data := DataStore.GetAllData()
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
