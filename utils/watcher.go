package utils

import (
	"strings"
)

func EnvToMap(env []string) map[string]string {
	kv := make(map[string]string)
	for _, value := range env {
		tempArray := strings.Split(value, "=")
		kv[tempArray[0]] = tempArray[1]
	}
	return kv
}
