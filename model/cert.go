package model

import (
	"encoding/base64"
	censored "github.com/allape/gocensored"
	"github.com/allape/gogger"
	"github.com/allape/stepin/env"
	"github.com/allape/stepin/stepin"
	"github.com/allape/stepin/stepin/create"
	"time"
)

var (
	CrtSalt = []byte("_crt_salt")
	KeySalt = []byte("_key_salt")
)

type Base struct {
	ID      uint       `gorm:"primaryKey" json:"id"`
	Created time.Time  `gorm:"autoCreateTime" json:"created"`
	Updated time.Time  `gorm:"autoUpdateTime" json:"updated"`
	Deleted *time.Time `json:"deleted"`
}

var l = gogger.New("model.item")
var (
	CrtCensor *censored.Censor
	KeyCensor *censored.Censor
)

func init() {
	var err error
	CrtCensor, err = censored.NewDefaultCensor(&censored.Config{
		TagName:  "crtcensored",
		Password: append([]byte(env.DatabasePassword), CrtSalt...),
	})
	if err != nil {
		l.Error().Fatalf("failed to create censor: %v", err)
	}

	KeyCensor, err = censored.NewDefaultCensor(&censored.Config{
		TagName:  "keycensored",
		Password: append([]byte(env.DatabasePassword), KeySalt...),
	})
	if err != nil {
		l.Error().Fatalf("failed to create censor: %v", err)
	}
}

type CensoredField string

func (s CensoredField) ToBytes() []byte {
	decoded, _ := base64.StdEncoding.DecodeString(string(s))
	return decoded
}

type Cert struct {
	Base
	Profile    create.Profile     `json:"profile"`
	Name       create.SubjectName `json:"name"`
	Crt        CensoredField      `json:"crt" crtcensored:"saltyaes.base64"`
	Key        CensoredField      `json:"key" keycensored:"saltyaes.base64"`
	Inspection stepin.Inspection  `json:"inspection"`
}

func (c *Cert) Encode() error {
	err := CrtCensor.Encencor(c)
	if err != nil {
		return err
	}

	err = KeyCensor.Encencor(c)
	if err != nil {
		return err
	}

	return nil
}

func (c *Cert) Decode() error {
	err := CrtCensor.Decensor(c)
	if err != nil {
		return err
	}

	err = KeyCensor.Decensor(c)
	if err != nil {
		return err
	}

	return nil
}

func (c *Cert) Strip() *Cert {
	c.Crt = ""
	c.Key = ""
	return c
}
