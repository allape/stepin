package env

import "os"

type EnvarName string

const (
	StepinMode             EnvarName = "STEPIN_MODE"
	StepinListen           EnvarName = "STEPIN_LISTEN"
	StepinAssetsFolder     EnvarName = "STEPIN_ASSETS_FOLDER"
	StepinDatabaseFilename EnvarName = "STEPIN_DATABASE_FILENAME"
)

func Get(envar EnvarName, defaultValue string) string {
	env := os.Getenv(string(envar))
	if env == "" {
		return defaultValue
	}
	return env
}
