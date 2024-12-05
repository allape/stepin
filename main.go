package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/allape/gocrud"
	"github.com/allape/gogger"
	"github.com/allape/stepin/env"
	"github.com/allape/stepin/model"
	"github.com/allape/stepin/stepin"
	"github.com/allape/stepin/stepin/create"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"
	"unicode"
)

var l = gogger.New("main")

func main() {
	err := gogger.InitFromEnv()
	if err != nil {
		l.Error().Fatalf("failed to init logger: %v", err)
	}

	dl := logger.New(gogger.New("db").Debug(), logger.Config{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      logger.Info,
		Colorful:      true,
	})
	db, err := gorm.Open(sqlite.Open(env.DatabaseFilename), &gorm.Config{
		Logger: dl,
	})
	if err != nil {
		l.Error().Fatalf("failed to create database: %v", err)
	}

	err = db.AutoMigrate(&model.Cert{})
	if err != nil {
		l.Error().Fatalf("failed to auto migrate database: %v", err)
	}

	engine := gin.Default()

	if env.HttpCors {
		engine.Use(cors.Default())
	}

	apiGroup := engine.Group("api")

	apiGroup.PATCH("recovery", func(context *gin.Context) {
		file, err := os.Open("./cert.json")
		if err != nil {
			gocrud.MakeErrorResponse(context, gocrud.RestCoder.InternalServerError(), err)
			return
		}

		bs, err := io.ReadAll(file)
		if err != nil {
			gocrud.MakeErrorResponse(context, gocrud.RestCoder.InternalServerError(), err)
			return
		}

		var certs []model.Cert
		err = json.Unmarshal(bs, &certs)
		if err != nil {
			gocrud.MakeErrorResponse(context, gocrud.RestCoder.InternalServerError(), err)
			return
		}

		for i := 0; i < len(certs); i++ {
			err = certs[i].Encode()
			if err != nil {
				gocrud.MakeErrorResponse(context, gocrud.RestCoder.InternalServerError(), err)
				return
			}
		}

		err = db.Save(&certs).Error
		if err != nil {
			gocrud.MakeErrorResponse(context, gocrud.RestCoder.InternalServerError(), err)
			return
		}

		context.JSON(http.StatusOK, gocrud.R[any]{
			Code: gocrud.RestCoder.OK(),
		})
	})

	err = SetupCertController(apiGroup, db)
	if err != nil {
		l.Error().Fatalf("failed to setup cert controller: %v", err)
	}

	uiGroup := engine.Group("ui")
	err = gocrud.NewSingleHTMLServe(uiGroup, env.UIIndex, &gocrud.SingleHTMLServeConfig{
		AllowReplace: false,
	})
	if err != nil {
		l.Error().Fatalln("failed to setup ui controller", err)
	}

	err = engine.Run(env.HttpAddress)
	if err != nil {
		l.Error().Fatalln("failed to start server", err)
	}
}

type PutCertBody struct {
	Name             create.SubjectName `json:"name"`
	Pass             create.Password    `json:"pass"`
	Years            int64              `json:"years"`
	KeyType          create.KeyType     `json:"keyType"`
	ParentCaID       uint               `json:"parentCaID"`
	ParentCaPassword create.Password    `json:"parentCaPassword"`
}

type DownloadType string

var (
	DownloadCRT       DownloadType = "crt"
	DownloadKey       DownloadType = "key"
	DownloadableTypes              = []DownloadType{DownloadCRT, DownloadKey}
)

func SetupCertController(group *gin.RouterGroup, db *gorm.DB) error {
	group = group.Group("cert")
	certCrudy := gocrud.CRUD[model.Cert]{
		EnableGetAll:  true,
		DisablePage:   true,
		DisableCount:  false,
		DisableDelete: true,
		DisableSave:   true,
		DisableGetOne: true,
		DidGetAll: func(record []model.Cert, ctx *gin.Context, repo *gorm.DB) error {
			for i := range record {
				record[i].Strip()
			}
			return nil
		},
	}
	err := certCrudy.Setup(group, db)
	if err != nil {
		return err
	}

	group.PUT(":profile", func(context *gin.Context) {
		profile, err := handleCertProfile(create.Profile(context.Param("profile")))
		if err != nil {
			gocrud.MakeErrorResponse(context, gocrud.RestCoder.BadRequest(), err)
			return
		}

		var body PutCertBody
		err = context.BindJSON(&body)
		if err != nil {
			gocrud.MakeErrorResponse(context, gocrud.RestCoder.BadRequest(), err)
			return
		}

		if body.Name == "" {
			gocrud.MakeErrorResponse(context, gocrud.RestCoder.BadRequest(), fmt.Errorf("name is required"))
		}

		commandBinOption := stepin.OptionCommandBin{
			CommandBin: "step",
		}
		commandBin := env.Bin
		if commandBin != "" {
			commandBinOption.CommandBin = commandBin
		}

		options := []stepin.CommandOption{
			commandBinOption,
		}

		if body.KeyType != "" {
			if slices.Contains(create.AllKeyTypes, body.KeyType) {
				options = append(options, create.OptionKeyType{
					KTY: body.KeyType,
				})
			} else {
				gocrud.MakeErrorResponse(context, gocrud.RestCoder.BadRequest(), fmt.Errorf("invalid key type"))
				return
			}
		}

		if body.Years > 0 {
			options = append(options, create.OptionNotAfter{
				NotAfter: time.Now().Add(time.Duration(body.Years*365*24) * time.Hour),
			})
		}

		var (
			inspection stepin.Inspection
			crt        create.Crt
			key        create.Key
		)

		switch profile {
		case create.RootCA:
			body.Pass, err = handleRootCAPassword(body.Pass)
			if err != nil {
				gocrud.MakeErrorResponse(context, gocrud.RestCoder.BadRequest(), err)
				return
			}

			inspection, crt, key, err = create.NewRootCA(create.RootOptions{
				PrimaryOptions: create.PrimaryOptions{
					Subject:  body.Name,
					Password: body.Pass,
				},
			}, options...)
			if err != nil {
				gocrud.MakeErrorResponse(context, gocrud.RestCoder.InternalServerError(), err)
				return
			}
		case create.IntermediateCA:
			body.Pass, err = handleIntermediateCAPassword(body.Pass)
			if err != nil {
				gocrud.MakeErrorResponse(context, gocrud.RestCoder.BadRequest(), err)
				return
			}
			if body.ParentCaID == 0 {
				gocrud.MakeErrorResponse(context, gocrud.RestCoder.BadRequest(), fmt.Errorf("parent ca is required for intermediate cert"))
				return
			}

			var parentCa model.Cert
			err = db.First(&parentCa, body.ParentCaID).Error
			if err != nil {
				gocrud.MakeErrorResponse(context, gocrud.RestCoder.NotFound(), err)
				return
			}

			err = parentCa.Decode()
			if err != nil {
				gocrud.MakeErrorResponse(context, gocrud.RestCoder.InternalServerError(), err)
				return
			}

			rootPassword, err := handleRootCAPassword(body.ParentCaPassword)
			if err != nil {
				gocrud.MakeErrorResponse(context, gocrud.RestCoder.BadRequest(), err)
				return
			}

			inspection, crt, key, err = create.NewIntermediateCA(create.RootlessOptions{
				PrimaryOptions: create.PrimaryOptions{
					Subject:  body.Name,
					Password: body.Pass,
				},
				RootCaCrt:    parentCa.Crt.ToBytes(),
				RootCaKey:    parentCa.Key.ToBytes(),
				RootPassword: rootPassword,
			}, options...)
			if err != nil {
				gocrud.MakeErrorResponse(context, gocrud.RestCoder.InternalServerError(), err)
				return
			}
		case create.Leaf:
			if body.ParentCaID == 0 {
				gocrud.MakeErrorResponse(context, gocrud.RestCoder.BadRequest(), fmt.Errorf("parent ca is required for leaf cert"))
				return
			}

			var parentCa model.Cert
			err = db.First(&parentCa, body.ParentCaID).Error
			if err != nil {
				gocrud.MakeErrorResponse(context, gocrud.RestCoder.NotFound(), err)
				return
			}

			err = parentCa.Decode()
			if err != nil {
				gocrud.MakeErrorResponse(context, gocrud.RestCoder.InternalServerError(), err)
				return
			}

			var parentPassword create.Password
			if parentCa.Profile == create.RootCA {
				parentPassword, err = handleRootCAPassword(body.ParentCaPassword)
			} else {
				parentPassword, err = handleIntermediateCAPassword(body.ParentCaPassword)
			}
			if err != nil {
				gocrud.MakeErrorResponse(context, gocrud.RestCoder.BadRequest(), err)
				return
			}

			inspection, crt, key, err = create.NewTLS(create.RootlessOptions{
				PrimaryOptions: create.PrimaryOptions{
					Subject: body.Name,
					// no password on leaf
					//Password: body.Pass,
				},
				RootCaCrt:    parentCa.Crt.ToBytes(),
				RootCaKey:    parentCa.Key.ToBytes(),
				RootPassword: parentPassword,
			}, options...)
			if err != nil {
				gocrud.MakeErrorResponse(context, gocrud.RestCoder.InternalServerError(), err)
				return
			}
		}

		cert := &model.Cert{
			Profile:    profile,
			Name:       body.Name,
			Crt:        model.CensoredField(base64.StdEncoding.EncodeToString(crt)),
			Key:        model.CensoredField(base64.StdEncoding.EncodeToString(key)),
			Inspection: inspection,
		}

		err = cert.Encode()
		if err != nil {
			gocrud.MakeErrorResponse(context, gocrud.RestCoder.InternalServerError(), err)
			return
		}

		err = db.Create(cert).Error
		if err != nil {
			gocrud.MakeErrorResponse(context, gocrud.RestCoder.InternalServerError(), err)
			return
		}

		context.JSON(http.StatusOK, gocrud.R[*model.Cert]{
			Code: gocrud.RestCoder.OK(),
			Data: cert.Strip(),
		})
	})

	group.GET(":type/:id", func(context *gin.Context) {
		certType := context.Param("type")
		if !slices.Contains(DownloadableTypes, DownloadType(certType)) {
			gocrud.MakeErrorResponse(context, gocrud.RestCoder.BadRequest(), fmt.Errorf("invalid download type"))
			return
		}

		id := context.Param("id")

		var cert model.Cert
		err := db.First(&cert, id).Error
		if err != nil {
			gocrud.MakeErrorResponse(context, gocrud.RestCoder.NotFound(), err)
			return
		}

		err = cert.Decode()
		if err != nil {
			gocrud.MakeErrorResponse(context, gocrud.RestCoder.InternalServerError(), err)
			return
		}

		var data []byte
		switch DownloadType(certType) {
		case DownloadCRT:
			data = cert.Crt.ToBytes()
		case DownloadKey:
			if cert.Profile == create.RootCA || cert.Profile == create.IntermediateCA {
				gocrud.MakeErrorResponse(context, gocrud.RestCoder.BadRequest(), fmt.Errorf("root/intermediate ca key is not downloadable"))
				return
			}
			data = cert.Key.ToBytes()
		}

		dataAttachment(context, data, fmt.Sprintf("%s.%s", cert.Name, certType))
	})

	return nil
}

func handleCertProfile(profile create.Profile) (create.Profile, error) {
	if profile == "" || !slices.Contains(create.AllProfiles, profile) {
		return "", fmt.Errorf("invalid certificate profile")
	}
	return profile, nil
}

func handleRootCAPassword(password create.Password) (create.Password, error) {
	if password != "" {
		return password, nil
	}
	password = create.Password(env.RootCAPassword)
	if password == "" {
		return "", fmt.Errorf("root ca password is empty")
	}
	return password, nil
}

func handleIntermediateCAPassword(password create.Password) (create.Password, error) {
	if password != "" {
		return password, nil
	}
	password = create.Password(env.IntermediateCAPassword)
	if password == "" {
		return "", fmt.Errorf("intermediate ca password is empty")
	}
	return password, nil
}

// vendor/go/pkg/mod/github.com/gin-gonic/gin@v1.9.1/context.go:1055
var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

// vendor/go/pkg/mod/github.com/gin-gonic/gin@v1.9.1/context.go:1063
func dataAttachment(c *gin.Context, data []byte, filename string) {
	if isASCII(filename) {
		c.Writer.Header().Set("Content-Disposition", `attachment; filename="`+escapeQuotes(filename)+`"`)
	} else {
		c.Writer.Header().Set("Content-Disposition", `attachment; filename*=UTF-8''`+url.QueryEscape(filename))
	}
	c.Data(http.StatusOK, "application/octet-stream", data)
}

// vendor/go/pkg/mod/github.com/gin-gonic/gin@v1.9.1/utils.go:157
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}
