package pkg

import (
	"os"
	"strconv"
)

func GetPortFromEnv(envKey string, defaultPort int) int {
	if strVal, found := os.LookupEnv(envKey); found {
		if i, err := strconv.ParseInt(strVal, 10, 8); err == nil {
			return int(i)
		}
	}
	return defaultPort
}

func GetStringFromEnv(envKey, defaultVal string) string {
	if strVal, found := os.LookupEnv(envKey); found {
		return strVal
	}
	return defaultVal
}
