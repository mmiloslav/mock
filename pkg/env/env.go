package env

import (
	"fmt"
	"os"
)

const (
	debugVarName = "DEBUG"
	trueStr      = "true"
)

// GetVar gets var from env
func GetVar(name string) (string, error) {
	val, ok := os.LookupEnv(name)
	if !ok {
		return "", fmt.Errorf("[%s] not found in env", name)
	}

	return val, nil
}

func Debug() bool {
	return os.Getenv(debugVarName) == trueStr
}
