package lets

import (
	"os"
	"strings"
)

func makeDockerVars(
	envs []string, lookupEnv func(string) (string, bool),
) map[string]string {
	if lookupEnv == nil {
		lookupEnv = os.LookupEnv
	}

	m := make(map[string]string)

	for _, envVar := range envs {
		k, v, found := strings.Cut(envVar, "=")
		if found {
			m[k] = v
			continue
		}

		v, ok := lookupEnv(envVar)
		if ok {
			m[envVar] = v
		}
	}

	return m
}
