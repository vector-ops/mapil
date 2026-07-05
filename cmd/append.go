package cmd

import (
	"context"
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/vector-ops/mapil/helpers"
)

var apdCmd = &cobra.Command{
	Use:   "apd",
	Short: "Append to an object",
	Long:  `Append one or more values to the list`,
	Run: func(cmd *cobra.Command, args []string) {
		apdObj(cmd.Context(), args)
	},
}

func apdObj(ctx context.Context, args []string) {

	var key string
	var err error

	if len(args) > 0 {
		key = args[0]
	}

	keys := dataStore.GetKeys(ctx)
	if len(keys) == 0 {
		fmt.Println("Data store empty.")
		return
	}

	selectTemplates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "> {{ . | green | underline }}",
		Inactive: "  {{ . | cyan }}",
		Selected: "{{ . | red | cyan }}",
	}

	selectPrompt := promptui.Select{
		Label:     "? Choose a key to append to:",
		Items:     keys,
		Templates: selectTemplates,
	}

	if len(key) == 0 {
		_, key, err = selectPrompt.Run()
		if err != nil {
			fmt.Printf("Prompt cancelled %s\n", err)
			return
		}
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

	valuePrompt := promptui.Prompt{
		Label:     "? Enter the values to append: ",
		Templates: templates,
		Validate:  validate,
	}

	value, err := valuePrompt.Run()
	if err != nil {
		fmt.Printf("Prompt cancelled %s\n", err)
		return
	}

	vals := helpers.CleanInput(value)

	err = dataStore.AppendList(ctx, key, vals, true)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("'%s' updated.\n", key)
}
