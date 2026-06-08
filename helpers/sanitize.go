package helpers

import "strings"

func CleanInput(value string) []string {
	vals := strings.Split(value, ",")
	for i := 0; i < len(vals); i++ {
		vals[i] = strings.TrimSpace(vals[i])
	}

	return vals
}
