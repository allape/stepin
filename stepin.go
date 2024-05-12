package main

import (
	"encoding/json"
	"fmt"
	"github.com/allape/stepin/env"
	"github.com/allape/stepin/salt"
	"github.com/allape/stepin/stepin"
	"github.com/allape/stepin/stepin/create"
	"github.com/gin-gonic/gin"
	"github.com/nalgeon/redka"
	"net/http"
	"net/url"
	"slices"
	"sort"
	"strings"
	"time"
	"unicode"
)

var (
	CrtSalt = []byte("_crt_salt")
	KeySalt = []byte("_key_salt")
	TxtSalt = []byte("_txt_salt")
)

type (
	DBKey  string
	CertID DBKey

	Cert struct {
		ID         CertID             `json:"id"`
		Profile    create.Profile     `json:"profile"`
		Name       create.SubjectName `json:"name"`
		Crt        create.Crt         `json:"crt"`
		Key        create.Key         `json:"key"`
		Inspection stepin.Inspection  `json:"inspection"`
	}

	SaltyCert struct {
		ID         CertID         `json:"id"`
		Profile    create.Profile `json:"profile"`
		Name       salt.HexString `json:"name"`
		Crt        salt.HexString `json:"crt"`
		Key        salt.HexString `json:"key"`
		Inspection salt.HexString `json:"inspection"`
	}
)

func BuildDBKey(profile create.Profile, name create.SubjectName) DBKey {
	return DBKey(fmt.Sprintf("cert:%s:%s", profile, salt.Sha256ToHexStringFromString(string(name))))
}

func BuildProfileDBKeyPattern(profile create.Profile) string {
	return fmt.Sprintf("cert:%s:*", profile)
}

func BuildAllCertDBKeyPattern() string {
	return "cert:*"
}

func GetCert(key DBKey, db *redka.DB) (*Cert, error) {
	value, err := db.Str().Get(string(key))
	if err != nil {
		return nil, err
	}

	if !value.Exists() {
		return nil, fmt.Errorf("cert not found")
	}

	var saltyCert SaltyCert
	err = json.Unmarshal([]byte(value.String()), &saltyCert)
	if err != nil {
		return nil, err
	}

	cert := &Cert{
		ID:      saltyCert.ID,
		Profile: saltyCert.Profile,
	}

	name, err := salt.DecodeFromHexStringToString(saltyCert.Name, TxtSalt)
	if err != nil {
		return nil, err
	}
	cert.Name = create.SubjectName(name)

	inspection, err := salt.DecodeFromHexStringToString(saltyCert.Inspection, TxtSalt)
	if err != nil {
		return nil, err
	}
	cert.Inspection = stepin.Inspection(inspection)

	cert.Crt, err = salt.DecodeFromHexString(saltyCert.Crt, CrtSalt)
	if err != nil {
		return nil, err
	}

	cert.Key, err = salt.DecodeFromHexString(saltyCert.Key, KeySalt)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func SetCert(key DBKey, cert *Cert, db *redka.DB) (*SaltyCert, error) {
	saltyName, err := salt.EncodeToHexString([]byte(cert.Name), TxtSalt)
	if err != nil {
		return nil, err
	}

	saltyCrt, err := salt.EncodeToHexString(cert.Crt, CrtSalt)
	if err != nil {
		return nil, err
	}

	saltyKey, err := salt.EncodeToHexString(cert.Key, KeySalt)
	if err != nil {
		return nil, err
	}

	saltyInspection, err := salt.EncodeToHexString([]byte(cert.Inspection), TxtSalt)
	if err != nil {
		return nil, err
	}

	cert.ID = CertID(key)
	saltyCert := SaltyCert{
		ID:         cert.ID,
		Profile:    cert.Profile,
		Name:       saltyName,
		Crt:        saltyCrt,
		Key:        saltyKey,
		Inspection: saltyInspection,
	}

	value, err := json.Marshal(saltyCert)
	if err != nil {
		return nil, err
	}

	err = db.Str().Set(string(key), value)
	if err != nil {
		return nil, err
	}

	return &saltyCert, nil
}

type DownloadType string

var (
	DownloadCRT       DownloadType = "crt"
	DownloadKey       DownloadType = "key"
	DownloadableTypes              = []DownloadType{DownloadCRT, DownloadKey}
)

type (
	GetCertBody struct {
		ID         CertID             `json:"id"`
		Profile    create.Profile     `json:"profile"`
		Name       create.SubjectName `json:"name"`
		Inspection stepin.Inspection  `json:"inspection"`
	}

	PutCertBody struct {
		Name           create.SubjectName `json:"name"`
		Pass           create.Password    `json:"pass"`
		Years          int64              `json:"years"`
		KeyType        create.KeyType     `json:"keyType"`
		RootCaID       CertID             `json:"rootCaID"`
		RootCaPassword create.Password    `json:"rootCaPassword"`
	}
)

func handleCertProfile(profile create.Profile) (create.Profile, error) {
	if profile != "" && !slices.Contains(create.AllProfiles, profile) {
		return "", fmt.Errorf("invalid certificate profile")
	}
	return profile, nil
}

func handleRootCAPassword(password create.Password) (create.Password, error) {
	if password != "" {
		return password, nil
	}
	password = create.Password(env.Get(env.StepinRootCAPassword, string(password)))
	if password == "" {
		return "", fmt.Errorf("root ca password is empty")
	}
	return password, nil
}

func handleIntermediateCAPassword(password create.Password) (create.Password, error) {
	if password != "" {
		return password, nil
	}
	password = create.Password(env.Get(env.StepinIntermediateCAPassword, string(password)))
	if password == "" {
		return "", fmt.Errorf("intermediate ca password is empty")
	}
	return password, nil
}

func SetupStepinServer(server *gin.Engine, db *redka.DB) error {
	server.GET("/cert", func(c *gin.Context) {
		ct := c.Query("profile")
		profile, err := handleCertProfile(create.Profile(ct))
		if err != nil {
			c.JSON(http.StatusBadRequest, R[any]{"-1", err.Error(), nil})
			return
		}

		var dbKeys []redka.Key

		if profile != "" {
			dbKeys, err = db.Key().Keys(BuildProfileDBKeyPattern(profile))
		} else {
			dbKeys, err = db.Key().Keys(BuildAllCertDBKeyPattern())
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, R[any]{"-1", err.Error(), nil})
			return
		}

		var certs []GetCertBody
		for _, key := range dbKeys {
			cert, err := GetCert(DBKey(key.Key), db)
			if err != nil {
				c.JSON(http.StatusInternalServerError, R[any]{"-1", err.Error(), nil})
				return
			}

			certs = append(certs, GetCertBody{
				ID:         cert.ID,
				Profile:    cert.Profile,
				Name:       cert.Name,
				Inspection: cert.Inspection,
			})
		}

		sort.Slice(certs, func(i, j int) bool {
			return certs[i].Name < certs[j].Name
		})

		c.JSON(http.StatusOK, R[[]GetCertBody]{"0", "OK", certs})
	})

	server.GET("/cert/:crtOrKey/:id", func(c *gin.Context) {
		crtOrKey := c.Param("crtOrKey")
		if !slices.Contains(DownloadableTypes, DownloadType(crtOrKey)) {
			c.JSON(http.StatusBadRequest, R[any]{"-1", "invalid download type", nil})
			return
		}

		id := c.Param("id")
		cert, err := GetCert(DBKey(id), db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, R[any]{"-1", err.Error(), nil})
			return
		}

		var data []byte
		switch DownloadType(crtOrKey) {
		case DownloadCRT:
			data = cert.Crt
		case DownloadKey:
			if cert.Profile == create.RootCA || cert.Profile == create.IntermediateCA {
				c.JSON(http.StatusBadRequest, R[any]{"-1", "root/intermediate ca key is not downloadable", nil})
				return
			}
			data = cert.Key
		}

		dataAttachment(
			c,
			http.StatusOK,
			"application/octet-stream",
			data,
			fmt.Sprintf("%s.%s", cert.Name, crtOrKey),
		)
	})

	server.PUT("/cert/:profile", func(c *gin.Context) {
		ct := c.Param("profile")
		profile, err := handleCertProfile(create.Profile(ct))
		if err != nil {
			c.JSON(http.StatusBadRequest, R[any]{"-1", err.Error(), nil})
			return
		}
		if profile == "" {
			c.JSON(http.StatusBadRequest, R[any]{"-1", "invalid certificate profile", nil})
			return
		}

		var body PutCertBody
		err = c.BindJSON(&body)
		if err != nil {
			c.JSON(http.StatusBadRequest, R[any]{"-1", err.Error(), nil})
			return
		}

		if body.Name == "" {
			c.JSON(http.StatusBadRequest, R[any]{"-1", "name is required", nil})
			return
		}

		dbKey := BuildDBKey(profile, body.Name)
		_, err = GetCert(dbKey, db)
		if err == nil {
			c.JSON(http.StatusBadRequest, R[any]{
				"-1",
				"certificate already exists, deletion only can be done in database manually",
				nil,
			})
			return
		}

		commandBinOption := stepin.OptionCommandBin{
			CommandBin: "step",
		}
		commandBin := env.Get(env.StepinBin, "")
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
				c.JSON(http.StatusBadRequest, R[any]{"-1", "invalid key type", nil})
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
				c.JSON(http.StatusBadRequest, R[any]{"-1", err.Error(), nil})
				return
			}

			inspection, crt, key, err = create.NewRootCA(create.RootOptions{
				PrimaryOptions: create.PrimaryOptions{
					Subject:  body.Name,
					Password: body.Pass,
				},
			}, options...)
			if err != nil {
				c.JSON(http.StatusInternalServerError, R[any]{"-1", err.Error(), nil})
				return
			}
		case create.IntermediateCA:
			body.Pass, err = handleIntermediateCAPassword(body.Pass)
			if err != nil {
				c.JSON(http.StatusBadRequest, R[any]{"-1", err.Error(), nil})
				return
			}
			if body.RootCaID == "" {
				c.JSON(http.StatusBadRequest, R[any]{"-1", "parent ca is required for intermediate cert", nil})
				return
			}
			cert, err := GetCert(DBKey(body.RootCaID), db)
			if err != nil {
				c.JSON(http.StatusInternalServerError, R[any]{"-1", err.Error(), nil})
				return
			}
			rootPassword, err := handleRootCAPassword(body.RootCaPassword)
			if err != nil {
				c.JSON(http.StatusBadRequest, R[any]{"-1", err.Error(), nil})
				return
			}

			inspection, crt, key, err = create.NewIntermediateCA(create.RootlessOptions{
				PrimaryOptions: create.PrimaryOptions{
					Subject:  body.Name,
					Password: body.Pass,
				},
				RootCaCrt:    cert.Crt,
				RootCaKey:    cert.Key,
				RootPassword: rootPassword,
			}, options...)
			if err != nil {
				c.JSON(http.StatusInternalServerError, R[any]{"-1", err.Error(), nil})
				return
			}
		case create.Leaf:
			if body.RootCaID == "" {
				c.JSON(http.StatusBadRequest, R[any]{"-1", "parent ca is required for leaf cert", nil})
				return
			}

			cert, err := GetCert(DBKey(body.RootCaID), db)
			if err != nil {
				c.JSON(http.StatusInternalServerError, R[any]{"-1", err.Error(), nil})
				return
			}

			var rootPassword create.Password
			if cert.Profile == create.RootCA {
				rootPassword, err = handleRootCAPassword(body.RootCaPassword)
			} else {
				rootPassword, err = handleIntermediateCAPassword(body.RootCaPassword)
			}
			if err != nil {
				c.JSON(http.StatusBadRequest, R[any]{"-1", err.Error(), nil})
				return
			}

			inspection, crt, key, err = create.NewTLS(create.RootlessOptions{
				PrimaryOptions: create.PrimaryOptions{
					Subject: body.Name,
					// no password on leaf
					//Password: body.Pass,
				},
				RootCaCrt:    cert.Crt,
				RootCaKey:    cert.Key,
				RootPassword: rootPassword,
			}, options...)
			if err != nil {
				c.JSON(http.StatusInternalServerError, R[any]{"-1", err.Error(), nil})
				return
			}
		}

		cert := &Cert{
			Profile:    profile,
			Name:       body.Name,
			Crt:        crt,
			Key:        key,
			Inspection: inspection,
		}
		_, err = SetCert(
			BuildDBKey(profile, body.Name),
			cert,
			db,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, R[any]{"-1", err.Error(), nil})
			return
		}

		c.JSON(http.StatusOK, R[CertID]{"0", "OK", cert.ID})
	})

	return nil
}

// vendor/go/pkg/mod/github.com/gin-gonic/gin@v1.9.1/context.go:1055
var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

// vendor/go/pkg/mod/github.com/gin-gonic/gin@v1.9.1/context.go:1063
func dataAttachment(c *gin.Context, status int, contentType string, data []byte, filename string) {
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
