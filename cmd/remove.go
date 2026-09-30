package cmd

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove an object",
	Long:  `Remove an object from the list`,
	Run: func(cmd *cobra.Command, args []string) {
		rmObj(cmd.Context(), args)
	},
}

func rmObj(ctx context.Context, args []string) {
	var key string
	var id int
	var err error

	if len(args) > 0 {
		key = args[0]
	}

	if len(args) > 1 {
		idS := args[1]
		id, err = strconv.Atoi(idS)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	keys := dataStore.GetKeys(ctx)
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
		Label:     "Select the key for the object you want to remove.",
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

	values, err := dataStore.GetValue(ctx, key)
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(values) == 0 {
		fmt.Printf("key '%s' has no values\n", key)
		return
	}

	if id > len(values) {
		fmt.Printf("id %d does not exist in the key '%s'\n", id, key)
		return
	}

	valuePrompt := promptui.Select{
		Label:        "Select the object you want to remove.",
		Items:        values,
		Templates:    templates,
		HideSelected: true,
	}

	var found bool
	if id <= 0 {
		_, val, err := valuePrompt.Run()
		if err != nil {
			fmt.Println("Prompt cancelled")
			return
		}

		slices.Sort(values)

		id, found = slices.BinarySearch(values, val)
		if !found {
			fmt.Printf("value not found in key '%s'\n", key)
			return
		}

		// id = idx+1
		id += 1
	}

	if id == len(values) {
		values = values[:id-1]
	} else {
		values = append(values[:id-1], values[id:]...)
	}

	ns, err := dataStore.GetNamespace(ctx, key)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = dataStore.UpdateList(ctx, key, values, ns)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Removed item %d from key '%s'.\n", id, key)
}
