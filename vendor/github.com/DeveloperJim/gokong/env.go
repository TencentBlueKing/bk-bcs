package gokong

import "os"

func GetEnvVarOrDefault(key string, defaultValue string) string {
	result := os.Getenv(key)

	if result == "" {
		return defaultValue
	}

	return result
}
