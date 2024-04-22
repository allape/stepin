package main

import (
	"github.com/allape/stepin/env"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nalgeon/redka"
	"log"
)

const Tag = "[stepin]"

type R[T any] struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

var Production = env.Get(env.StepinMode, "") == "release"
var IndexHTML = env.Get(env.StepinAssetsFolder, "assets/index.html")

func init() {
	if !Production {
		IndexHTML = "ui/dist/index.html"
	}
}

func main() {
	db, err := redka.Open(env.Get(env.StepinDatabaseFilename, "data.db"), nil)
	if err != nil {
		log.Fatal(Tag, err)
	}
	defer func() {
		_ = db.Close()
	}()

	server := gin.Default()

	if !Production {
		log.Println(Tag, "[WARNING] CORS is enabled!")
		server.Use(cors.Default())
	}

	err = SetupStepinServer(server, db)
	if err != nil {
		log.Fatalln(Tag, err)
	}

	server.StaticFile("/", IndexHTML)
	server.StaticFile("/index", IndexHTML)
	server.StaticFile("/index.html", IndexHTML)

	err = server.Run(env.Get(env.StepinListen, ":8080"))
	if err != nil {
		log.Fatalln(Tag, err)
	}
}
