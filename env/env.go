package env

import (
	"github.com/allape/goenv"
)

const (
	stepinHttpAddress = "STEPIN_HTTP_ADDRESS"
	stepinHttpCors    = "STEPIN_HTTP_CORS"
	stepinUIIndex     = "STEPIN_UI_INDEX"

	stepinBin = "STEPIN_BIN"

	stepinDatabaseFilename      = "STEPIN_DATABASE_FILENAME"
	stepinDatabaseFieldPassword = "STEPIN_DATABASE_FIELD_PASSWORD"

	stepinRootCAPassword         = "STEPIN_ROOT_CA_PASSWORD"
	stepinIntermediateCAPassword = "STEPIN_INTERMEDIATE_CA_PASSWORD"
)

var (
	HttpAddress = goenv.Getenv(stepinHttpAddress, ":8080")
	HttpCors    = goenv.Getenv(stepinHttpCors, true)
	UIIndex     = goenv.Getenv(stepinUIIndex, "ui/dist/index.html")

	Bin = goenv.Getenv(stepinBin, "stepin")

	DatabaseFilename = goenv.Getenv(stepinDatabaseFilename, "database/data.db")
	DatabasePassword = goenv.Getenv(stepinDatabaseFieldPassword, "123456")

	RootCAPassword         = goenv.Getenv(stepinRootCAPassword, "123456")
	IntermediateCAPassword = goenv.Getenv(stepinIntermediateCAPassword, "456789")
)
