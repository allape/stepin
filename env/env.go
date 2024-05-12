package env

import "os"

type EnvarName string

const (
	StepinMode         EnvarName = "STEPIN_MODE"
	StepinListen       EnvarName = "STEPIN_LISTEN"
	StepinBin          EnvarName = "STEPIN_BIN"
	StepinAssetsFolder EnvarName = "STEPIN_ASSETS_FOLDER"

	StepinDatabaseFilename EnvarName = "STEPIN_DATABASE_FILENAME"
	StepinDatabaseSalt     EnvarName = "STEPIN_DATABASE_SALT"
	StepinDatabasePassword EnvarName = "STEPIN_DATABASE_PASSWORD"

	StepinRootCAPassword         EnvarName = "STEPIN_ROOT_CA_PASSWORD"
	StepinIntermediateCAPassword EnvarName = "STEPIN_INTERMEDIATE_CA_PASSWORD"
)

func Get(envar EnvarName, defaultValue string) string {
	env := os.Getenv(string(envar))
	if env == "" {
		return defaultValue
	}
	return env
}
