package main

import (
	"errors"
	"fmt"
	"github.com/allape/stepin/stepin"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
)

var (
	Bind       = ":8080"
	AllowedIPs = []string{"::1", "127.0.0.1"}
	Config     = stepin.CAConfig{
		RootCaName:           "root_ca",
		RootCaPassword:       "123456",
		IntermediaCaName:     "intermedia_ca",
		IntermediaCaPassword: "123_456",
	}
)

var (
	InitializeAuthKey = uuid.New().String()
	Initialized       = false
)

func init() {
	StepinBind := os.Getenv("STEPIN_BIND")
	if StepinBind != "" {
		Bind = StepinBind
	}

	StepinAllowedIpFile := os.Getenv("STEPIN_ALLOWED_IP_FILE")
	if StepinAllowedIpFile != "" {
		ips, err := os.ReadFile(StepinAllowedIpFile)
		if err == nil {
			AllowedIPs = append(AllowedIPs, strings.Split(string(ips), "\n")...)
		}
	}

	var ips []string
	for _, ip := range AllowedIPs {
		ip = strings.TrimSpace(ip)
		if ip != "" && !slices.Contains(ips, ip) {
			ips = append(ips, ip)
		}
	}
	AllowedIPs = ips

	RootCaPassword := os.Getenv("STEPIN_ROOT_CA_PASSWORD")
	if RootCaPassword != "" {
		Config.RootCaPassword = RootCaPassword
	}
	IntermediaCaPassword := os.Getenv("STEPIN_INTERMEDIA_CA_PASSWORD")
	if IntermediaCaPassword != "" {
		Config.IntermediaCaPassword = IntermediaCaPassword
	}

	StepinConfigFolder := os.Getenv("STEPIN_CONFIG_FOLDER")
	if StepinConfigFolder != "" {
		stepin.ConfigFolder = StepinConfigFolder
		stepin.Setup()
	}

	log.Println("Use config folder:", stepin.ConfigFolder)
	log.Println("Use leaf cert folder:", stepin.LeafCertFolder)

	Initialized = stepin.IsInitialized()
}

func ErrorPage(ctx *gin.Context, code int, err ...error) {
	ctx.HTML(code, "error.html", gin.H{
		"Errors": err,
	})
}

func IndexPage(ctx *gin.Context, code int, errs ...error) {
	certs, err := stepin.CertList(true)
	if err != nil {
		errs = append(errs, err)
	}
	ctx.HTML(code, "index.html", gin.H{
		"Errors":            errs,
		"Certs":             certs,
		"InitializeAuthKey": InitializeAuthKey,
		"IsInitialized":     Initialized,
	})
}

type CertificateForm struct {
	Filename     string `form:"filename"`
	Hostname     string `form:"hostname"`
	KeyType      string `form:"keyType"`
	ExpireInHour int    `form:"expireInHour"`
}

func main() {
	router := gin.Default()

	router.Use(func(ctx *gin.Context) {
		if !slices.Contains[[]string](AllowedIPs, ctx.ClientIP()) {
			ErrorPage(ctx, http.StatusUnauthorized, errors.New("permission denied"))
			ctx.Abort()
		}
	})

	router.SetFuncMap(template.FuncMap{
		"urlescaper":  template.URLQueryEscaper,
		"htmlescaper": template.HTMLEscapeString,
	})
	router.LoadHTMLGlob("templates/*")

	router.GET("/", func(ctx *gin.Context) {
		IndexPage(ctx, http.StatusOK)
	})
	router.GET("/index", func(ctx *gin.Context) {
		IndexPage(ctx, http.StatusOK)
	})

	router.GET("/download-root-ca", func(ctx *gin.Context) {
		_, err := os.Stat(stepin.RootCaCrtPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				ErrorPage(ctx, http.StatusNotFound, err)
				return
			}
			ErrorPage(ctx, http.StatusInternalServerError, err)
			return
		}
		ctx.FileAttachment(stepin.RootCaCrtPath, path.Base(stepin.RootCaCrtPath))
	})
	router.GET("/download-intermedia-ca", func(ctx *gin.Context) {
		_, err := os.Stat(stepin.IntermediaCaCrtPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				ErrorPage(ctx, http.StatusNotFound, err)
				return
			}
			ErrorPage(ctx, http.StatusInternalServerError, err)
			return
		}
		ctx.FileAttachment(stepin.IntermediaCaCrtPath, path.Base(stepin.IntermediaCaCrtPath))
	})

	router.GET("/download", func(ctx *gin.Context) {
		filename := strings.TrimSpace(ctx.Query("filename"))
		if filename == "" {
			ErrorPage(ctx, http.StatusBadRequest, errors.New("filename is required for downloading"))
			return
		}
		filename = path.Join(stepin.LeafCertFolder, filename)
		_, err := os.Stat(filename)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				ErrorPage(ctx, http.StatusNotFound, err)
				return
			}
			ErrorPage(ctx, http.StatusInternalServerError, err)
			return
		}
		ctx.FileAttachment(filename, filename)
	})

	router.GET("/initialize", func(ctx *gin.Context) {
		key := strings.TrimSpace(ctx.Query("key"))
		if key != InitializeAuthKey {
			ErrorPage(ctx, http.StatusBadRequest, errors.New("key is not valid"))
			return
		}
		err := stepin.Initialize(Config)
		if err != nil {
			ErrorPage(ctx, http.StatusInternalServerError, err)
			return
		}
		InitializeAuthKey = uuid.New().String()
		Initialized = stepin.IsInitialized()
		ctx.Redirect(http.StatusSeeOther, "/")
	})

	router.GET("/add", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "edit.html", gin.H{})
	})
	router.POST("/add.do", func(ctx *gin.Context) {
		form := CertificateForm{}
		err := ctx.Bind(&form)
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "edit.html", gin.H{
				"Errors":          []error{err},
				"CertificateForm": form,
			})
			return
		}

		form.Filename = strings.TrimSpace(form.Filename)
		if form.Filename == "" {
			ctx.HTML(http.StatusBadRequest, "edit.html", gin.H{
				"Errors":          []error{errors.New("filename must not be empty")},
				"CertificateForm": form,
			})
			return
		} else if ok, _ := regexp.MatchString("^[\\w.-]+$", form.Filename); !ok {
			ctx.HTML(http.StatusBadRequest, "edit.html", gin.H{
				"Errors":          []error{errors.New("filename is not valid")},
				"CertificateForm": form,
			})
			return
		}

		form.Hostname = strings.TrimSpace(form.Hostname)
		if form.Hostname == "" {
			ctx.HTML(http.StatusBadRequest, "edit.html", gin.H{
				"Errors":          []error{errors.New("hostname must not be empty")},
				"CertificateForm": form,
			})
			return
		}

		certs, err := stepin.CertList(false)
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "edit.html", gin.H{
				"Errors":          []error{err},
				"CertificateForm": form,
			})
			return
		}

		form.KeyType = strings.TrimSpace(form.KeyType)
		if form.KeyType == "" {
			form.KeyType = "EC"
		}

		for _, cert := range certs {
			if cert.Filename == form.Filename {
				ctx.HTML(http.StatusInternalServerError, "edit.html", gin.H{
					"Errors":          []error{fmt.Errorf("%s already exists", form.Filename)},
					"CertificateForm": form,
				})
				return
			}
		}

		if form.ExpireInHour <= 0 {
			form.ExpireInHour = 8760
		}

		err = stepin.SignCert(Config, form.Filename, form.Hostname, form.KeyType, form.ExpireInHour)
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "edit.html", gin.H{
				"Errors":          []error{err},
				"CertificateForm": form,
			})
			return
		}

		ctx.Redirect(http.StatusSeeOther, "/")
	})

	router.GET("/remove", func(ctx *gin.Context) {
		filename := strings.TrimSpace(ctx.Query("filename"))
		if filename == "" {
			IndexPage(ctx, http.StatusBadRequest, errors.New("filename can NOT be empty"))
			return
		}
		err := stepin.RemoveCert(filename)
		if err != nil {
			IndexPage(ctx, http.StatusInternalServerError, err)
			return
		}
		ctx.Redirect(http.StatusSeeOther, "/")
	})

	err := router.Run(Bind)
	if err != nil {
		panic(err)
	}
}
