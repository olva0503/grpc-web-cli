package template

import (
	"fmt"
	"strings"
)

// Substitute replaces variables in the format {{key}} with values from the map.
func Substitute(input string, variables map[string]interface{}) string {
	if len(variables) == 0 {
		return input
	}

	result := input
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		valStr := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, valStr)
	}

	return result
}
