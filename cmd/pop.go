package cmd

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var popCmd = &cobra.Command{
	Use:   "pop",
	Short: "Pop last object",
	Long:  `Delete last object in the values`,
	Run: func(cmd *cobra.Command, args []string) {
		popObj(args)
	},
}

func popObj(args []string) {

	var key string
	var err error

	if len(args) > 0 {
		key = args[0]
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
		Label:     "Select the key you want to pop.",
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

	values, err := dataStore.GetValue(key)
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(values) == 0 {
		fmt.Printf("key '%s' has no values", key)
		return
	}

	values = values[:len(values)-1]

	err = dataStore.UpdateList(key, values, "")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Item at %d for key '%s' deleted.\n", len(values)+1, key)
}
