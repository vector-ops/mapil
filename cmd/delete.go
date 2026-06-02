package cmd

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

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
		DataStore.DeleteAll()
		DataStore.Persist()
		fmt.Printf("Deleted all data.\n")
		return
	}
	keys := DataStore.GetKeys()
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

	DataStore.DeleteValue(key)
	err = DataStore.Persist()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("'%s' deleted.\n", key)
}
