package cmd

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var delAll *bool

func init() {
	delAll = delCmd.PersistentFlags().BoolP("all", "a", false, "delete all the data in the data store.")
}

var delCmd = &cobra.Command{
	Use:   "del",
	Short: "Delete an object",
	Long:  `Delete an object from the list`,
	Run: func(cmd *cobra.Command, args []string) {
		delObj(args)
	},
}

func delObj(args []string) {
	var key string
	var err error

	if len(args) > 0 {
		key = args[0]
	}

	if *delAll {
		dataStore.DeleteAll()
		fmt.Printf("Deleted all data.\n")
		return
	}
	keys := dataStore.GetKeys()
	if len(keys) == 0 {
		fmt.Println("Data store empty.")
		return
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "> {{ . | green | underline }}",
		Inactive: "  {{ . | cyan }}",
		Selected: "{{ . | red | cyan }}",
	}

	keyPrompt := promptui.Select{
		Label:     "Select the key you want to delete.",
		Items:     keys,
		Templates: templates,
	}

	if len(key) == 0 {
		_, key, err = keyPrompt.Run()
		if err != nil {
			fmt.Println("Prompt cancelled")
			return
		}
	}

	err = dataStore.DeleteValue(key)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("'%s' deleted.\n", key)
}
