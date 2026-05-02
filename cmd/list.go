package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
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
				fmt.Println(do.Key)
				for i, v := range do.Value {
					fmt.Printf("  %d. %s\n", i+1, v)
				}

				if i < len(data)-1 {
					fmt.Println()
				}
			}
		}
	},
}
