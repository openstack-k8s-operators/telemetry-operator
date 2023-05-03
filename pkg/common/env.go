package common

import "os"

// GetEnvDefault - Get the value associated with key from environment variables, but use baseDefault as a value in the case of an empty string
func GetEnvDefault(key string, baseDefault string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return baseDefault
}
