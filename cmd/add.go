package cmd

import (
	"context"
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/vector-ops/mapil/helpers"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "List all objects",
	Long:  `All objects stored are listed`,
	Run: func(cmd *cobra.Command, args []string) {
		addObj(cmd.Context(), args)
	},
}

var addns *string

func init() {
	addns = addCmd.PersistentFlags().StringP("namespace", "s", "", "add object to namespace")
}

func addObj(ctx context.Context, args []string) {
	var key string
	var err error

	if len(args) > 0 {
		key = args[0]
	}

	validate := func(input string) error {
		if input == "" {
			return fmt.Errorf("name should not be empty")
		}
		return nil
	}
	templates := &promptui.PromptTemplates{
		Prompt:  "{{ . }} ",
		Valid:   "{{ . | bold }} ",
		Invalid: "{{ . | bold }} ",
		Success: "{{ . | green }} ",
	}

	keyPrompt := promptui.Prompt{
		Label:     "? Enter a name for the key:",
		Templates: templates,
		Validate:  validate,
	}

	valuePrompt := promptui.Prompt{
		Label:     "? Enter the value(s):",
		Templates: templates,
		Validate:  validate,
	}

	if len(key) == 0 {
		key, err = keyPrompt.Run()
		if err != nil {
			fmt.Printf("Prompt cancelled %s\n", err)
			return
		}
	}

	value, err := valuePrompt.Run()
	if err != nil {
		fmt.Printf("Prompt cancelled %s\n", err)
		return
	}

	vals := helpers.CleanInput(value)

	err = dataStore.AddList(ctx, key, vals, *addns)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("'%s' successfully added to Mapil keyring.\n", key)
}
