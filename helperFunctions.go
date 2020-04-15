package basicGo

import (
	"os"
	"strconv"
	"time"
)

func GetIntEnv(key string) int {
	ret, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		return 0
	}
	return ret
}

func GetDurationEnv(key string) time.Duration {
return time.Second
}

func GetBoolEnv(key string) bool {
return false
}

func GetStringEnv(key string) string {
	return os.Getenv(key)
}