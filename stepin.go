package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/allape/stdhook"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

var (
	ConfigFolder         = "/etc/stepin"
	RootCACertName       = "root_ca"
	IntermediaCACertName = "intermedia_ca"
	LeafCertFolder       = path.Join(ConfigFolder, "leaf")
)

var rootCaCrtPath string
var rootCaKeyPath string
var intermediaCaKeyPath string
var intermediaCaCrtPath string

func init() {
	rootCaCrtPath = path.Join(ConfigFolder, RootCACertName+".crt")
	rootCaKeyPath = path.Join(ConfigFolder, RootCACertName+".key")
	intermediaCaKeyPath = path.Join(ConfigFolder, IntermediaCACertName+".crt")
	intermediaCaCrtPath = path.Join(ConfigFolder, IntermediaCACertName+".key")
}

func FlashExec(cmd string, args ...string) (string, error) {
	log.Println("Run command:", cmd, args)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, cmd, args...)
	output, err := command.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func Exec(onColon func(channel int, line string) string, cmd string, args ...string) error {
	log.Println("Exec:", cmd, args)
	config := &stdhook.Config{
		Timeout:               2 * time.Minute,
		TriggerWord:           "█",
		OnlyTriggerOnLastLine: true,
		OnTrigger:             onColon,
		OnOutput: func(channel int, content []byte) {
			if channel == 1 {
				_, _ = fmt.Fprint(os.Stdout, string(content))
			} else {
				_, _ = fmt.Fprint(os.Stderr, string(content))
			}
		},
	}
	return stdhook.Hook(config, cmd, args...)
}

type CAConfig struct {
	Force                bool
	RootCaName           string
	RootCaPassword       string
	IntermediaCaName     string
	IntermediaCaPassword string
}

func onTriggerBuilder(config CAConfig, crtPath, keyPath, password string) func(channel int, line string) string {
	return func(channel int, line string) string {
		if strings.Contains(line, fmt.Sprintf("Would you like to overwrite %s [y/n]:", crtPath)) ||
			strings.Contains(line, fmt.Sprintf("Would you like to overwrite %s [y/n]:", keyPath)) {
			if config.Force {
				return "y\n"
			}
			return "n\n"
		}
		if strings.Contains(line, fmt.Sprintf("Please enter the password to decrypt %s:", rootCaKeyPath)) {
			return config.RootCaPassword + "\n"
		}
		if strings.Contains(line, fmt.Sprintf("Please enter the password to decrypt %s:", intermediaCaKeyPath)) {
			return config.IntermediaCaPassword + "\n"
		}
		if strings.Contains(line, "Please enter the password to encrypt the private key:") {
			return password + "\n"
		}
		return ""
	}
}

func Initialize(config CAConfig) (err error) {
	_, err = os.Stat(LeafCertFolder)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.MkdirAll(LeafCertFolder, 0755)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	err = Exec(
		onTriggerBuilder(config, rootCaCrtPath, rootCaKeyPath, config.RootCaPassword),
		"step",
		"certificate",
		"create",
		config.RootCaName,
		rootCaCrtPath,
		rootCaKeyPath,
		"--profile",
		"root-ca",
	)
	if err != nil {
		return err
	}
	err = Exec(
		onTriggerBuilder(config, intermediaCaCrtPath, intermediaCaKeyPath, config.IntermediaCaPassword),
		"step",
		"certificate",
		"create",
		config.IntermediaCaName,
		intermediaCaCrtPath,
		intermediaCaKeyPath,
		"--profile",
		"intermediate-ca",
		"--ca",
		rootCaCrtPath,
		"--ca-key",
		rootCaKeyPath,
	)
	if err != nil {
		return err
	}

	return nil
}

func SignCert(config CAConfig, filename, hostname string, expireInHour int) error {
	crtPath := path.Join(LeafCertFolder, filename+".crt")
	keyPath := path.Join(LeafCertFolder, filename+".key")
	return Exec(
		onTriggerBuilder(config, crtPath, keyPath, ""),
		"step",
		"certificate",
		"create",
		hostname,
		crtPath,
		keyPath,
		"--profile",
		"leaf",
		"--ca",
		intermediaCaCrtPath,
		"--ca-key",
		intermediaCaKeyPath,
		"--insecure",
		"--no-password",
		"--not-after",
		fmt.Sprintf("%dh", expireInHour),
	)
}

type Cert struct {
	Filename   string
	Inspection string
}

func CertList() ([]Cert, error) {
	file, err := os.Open(LeafCertFolder)
	if err != nil {
		return nil, err
	}
	names, err := file.Readdirnames(0)
	if err != nil {
		return nil, err
	}

	var certs []Cert
	for _, name := range names {
		if path.Ext(name) == ".crt" {
			name = name[:len(name)-4]
			inspection, err := InspectCert(name, true)
			if err != nil {
				return nil, err
			}
			certs = append(certs, Cert{
				Filename:   name,
				Inspection: inspection,
			})
		}
	}

	return certs, nil
}

func InspectCert(filename string, short bool) (string, error) {
	crtPath := path.Join(LeafCertFolder, filename+".crt")
	args := []string{
		"certificate",
		"inspect",
		crtPath,
	}
	if short {
		args = append(args, "--short")
	}
	return FlashExec(
		"step",
		args...,
	)
}

func RemoveCert(filename string) error {
	crtPath := path.Join(LeafCertFolder, filename+".crt")
	err := os.Remove(crtPath)
	if err != nil {
		return err
	}
	keyPath := path.Join(LeafCertFolder, filename+".key")
	return os.Remove(keyPath)
}

func IsInitialized() bool {
	_, err := os.Stat(rootCaCrtPath)
	if err != nil {
		return false
	}
	_, err = os.Stat(rootCaKeyPath)
	if err != nil {
		return false
	}
	_, err = os.Stat(intermediaCaCrtPath)
	if err != nil {
		return false
	}
	_, err = os.Stat(intermediaCaKeyPath)
	if err != nil {
		return false
	}
	return true
}
