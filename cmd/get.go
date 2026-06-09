package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var getns *string

func init() {
	getns = getCmd.PersistentFlags().StringP("namespace", "s", "", "get object from namespace, if empty gets from the default namespace")
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a key",
	Long:  `Use key to get the value`,
	Run: func(cmd *cobra.Command, args []string) {
		key := ""
		if len(args) > 0 {
			key = args[0]
		} else {
			fmt.Println("get command requires a key")
			return
		}

		var values []string
		var err error

		if *getns == "" {

			values, err = dataStore.GetValue(key)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		} else {
			objs := dataStore.GetNamespaceObjects(*getns)
			var ok bool
			for _, obj := range objs {
				if obj.GetKey() == key {
					values, ok = obj.GetValue().([]string)
					if !ok {
						fmt.Println("failed to assert type of value")
						return
					}
					break
				}
			}

		}

		if len(values) == 0 {
			if *getns != "" {
				fmt.Printf("Key not found in namespace '%s'\n", *getns)
				return
			}
			fmt.Println("Key not found.")
		} else {
			fmt.Printf("  %s%s%s %s[%d]%s\n", Underline, key, Reset, DarkGrey, len(values), Reset)
			for i, v := range values {
				fmt.Printf("   %s%d.%s %s\n", DarkGrey, i+1, Reset, v)
			}
		}
	},
}
