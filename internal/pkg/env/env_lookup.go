package env

import "os"

func TrySetFromEnv(envName string, val *string) {
	if envVal, found := os.LookupEnv(envName); found {
		*val = envVal
	}
}
