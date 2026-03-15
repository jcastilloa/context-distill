package main

import "strings"

func shouldUseSetupConfigLoader(args []string) bool {
	for _, arg := range args {
		if arg == "--config-ui" || strings.HasPrefix(arg, "--config-ui=") {
			return true
		}
	}

	return false
}
