package main

import (
	"fmt"
	"github.com/allape/stepin/env"
	"github.com/allape/stepin/salt"
	"github.com/allape/stepin/stepin"
	"github.com/allape/stepin/stepin/create"
	"github.com/gin-gonic/gin"
	"github.com/nalgeon/redka"
	"net/http"
	"slices"
	"strings"
	"time"
)

func handleCertType(certType create.Profile) (create.Profile, error) {
	if certType != "" && !slices.Contains(create.AllProfiles, certType) {
		return "", fmt.Errorf("invalid certificate type")
	}
	return certType, nil
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

type GetCertBody struct {
	Key        string             `json:"key"`
	Name       create.SubjectName `json:"name"`
	Inspection stepin.Inspection  `json:"inspection"`
}

type PutCertBody struct {
	Name           create.SubjectName `json:"name"`
	Pass           create.Password    `json:"pass"`
	Years          int64              `json:"years"`
	RootCaName     string             `json:"rootCaName"`
	RootCaPassword create.Password    `json:"rootCaPassword"`
}

func GetCert(key string, db *redka.DB) (create.SubjectName, stepin.Inspection, create.Crt, create.Key, error) {
	value, err := db.Str().Get(key)
	if err != nil {
		return "", "", nil, nil, err
	}

	segments := strings.Split(value.String(), ",")
	if len(segments) != 3 {
		return "", "", nil, nil, fmt.Errorf("invalid content for %s", key)
	}

	name, err := salt.DecodeFromHexStringToString(salt.HexString(strings.Split(key, ":")[2]))
	if err != nil {
		return "", "", nil, nil, err
	}

	crtBytes, err := salt.DecodeFromHexString(salt.HexString(segments[0]))
	if err != nil {
		return "", "", nil, nil, err
	}
	keyBytes, err := salt.DecodeFromHexString(salt.HexString(segments[1]))
	if err != nil {
		return "", "", nil, nil, err
	}
	inspection, err := salt.DecodeFromHexStringToString(salt.HexString(segments[2]))
	if err != nil {
		return "", "", nil, nil, err
	}

	return create.SubjectName(name), stepin.Inspection(inspection), crtBytes, keyBytes, nil
}

func SetupStepinServer(server *gin.Engine, db *redka.DB) error {
	server.GET("/cert", func(c *gin.Context) {
		ct := c.Query("type")
		certType, err := handleCertType(create.Profile(ct))
		if err != nil {
			c.JSON(http.StatusBadRequest, R[any]{"-1", err.Error(), nil})
			return
		}

		var dbKeys []redka.Key

		if certType != "" {
			dbKeys, err = db.Key().Keys(fmt.Sprintf("cert:%s:*", certType))
		} else {
			dbKeys, err = db.Key().Keys("cert:*")
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, R[any]{"-1", err.Error(), nil})
			return
		}

		var certs []GetCertBody
		for _, key := range dbKeys {
			name, inspection, _, _, err := GetCert(key.Key, db)
			if err != nil {
				c.JSON(http.StatusInternalServerError, R[any]{"-1", err.Error(), nil})
				return
			}

			certs = append(certs, GetCertBody{
				Key:        key.Key,
				Name:       name,
				Inspection: inspection,
			})
		}
		c.JSON(http.StatusOK, R[[]GetCertBody]{"0", "OK", certs})
	})

	server.PUT("/cert/:type", func(c *gin.Context) {
		ct := c.Param("type")
		certType, err := handleCertType(create.Profile(ct))
		if err != nil {
			c.JSON(http.StatusBadRequest, R[any]{"-1", err.Error(), nil})
			return
		}
		if certType == "" {
			c.JSON(http.StatusBadRequest, R[any]{"-1", "invalid certificate type", nil})
			return
		}

		var body PutCertBody
		err = c.BindJSON(&body)
		if err != nil {
			c.JSON(http.StatusBadRequest, R[any]{"-1", err.Error(), nil})
			return
		}

		var (
			inspection stepin.Inspection
			crt        create.Crt
			key        create.Key
		)

		var options []create.CreationOption

		if body.Years > 0 {
			options = append(options, create.OptionNotAfter{
				NotAfter: time.Now().Add(time.Duration(body.Years*365*24) * time.Hour),
			})
		}

		switch certType {
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

		case create.IntermediateCA:
			body.Pass, err = handleIntermediateCAPassword(body.Pass)
			if err != nil {
				c.JSON(http.StatusBadRequest, R[any]{"-1", err.Error(), nil})
				return
			}
			_, _, rootCrt, rootKey, err := GetCert(body.RootCaName, db)
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
				RootCaCrt:    rootCrt,
				RootCaKey:    rootKey,
				RootPassword: rootPassword,
			}, options...)
		case create.Leaf:
			_, _, rootCrt, rootKey, err := GetCert(body.RootCaName, db)
			if err != nil {
				c.JSON(http.StatusInternalServerError, R[any]{"-1", err.Error(), nil})
				return
			}

			var rootPassword create.Password

			if strings.HasPrefix(body.RootCaName, fmt.Sprintf("cert:%s", create.RootCA)) {
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
				RootCaCrt:    rootCrt,
				RootCaKey:    rootKey,
				RootPassword: rootPassword,
			}, options...)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, R[any]{"-1", err.Error(), nil})
			return
		}

		saltyCrt, err := salt.EncodeToHexString(crt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, R[any]{"-1", err.Error(), nil})
			return
		}

		saltyKey, err := salt.EncodeToHexString(key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, R[any]{"-1", err.Error(), nil})
			return
		}

		saltyInspection, err := salt.EncodeToHexString([]byte(inspection))
		if err != nil {
			c.JSON(http.StatusInternalServerError, R[any]{"-1", err.Error(), nil})
			return
		}

		saltyName, err := salt.EncodeToHexString([]byte(body.Name))
		if err != nil {
			c.JSON(http.StatusInternalServerError, R[any]{"-1", err.Error(), nil})
			return
		}

		value := fmt.Sprintf(
			"%s,%s,%s",
			saltyCrt,
			saltyKey,
			saltyInspection,
		)

		keyName := fmt.Sprintf("cert:%s:%s", certType, saltyName)
		err = db.Str().Set(keyName, value)
		if err != nil {
			c.JSON(http.StatusInternalServerError, R[any]{"-1", err.Error(), nil})
			return
		}

		c.JSON(http.StatusOK, R[string]{"0", "OK", keyName})
	})

	return nil
}
