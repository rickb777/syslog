package env

import (
	"os"
	"strconv"
)

func GetString(key, def string) string {
	v, exists := os.LookupEnv(key)
	if !exists {
		return def
	}
	return v
}

func GetBool(key string, def bool) (bool, error) {
	v, exists := os.LookupEnv(key)
	if !exists {
		return def, nil
	}
	return strconv.ParseBool(v)
}

func GetInt(key string, def int) (int, error) {
	v, exists := os.LookupEnv(key)
	if !exists {
		return def, nil
	}
	return strconv.Atoi(v)
}
