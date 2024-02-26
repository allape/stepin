package stepin

// https://smallstep.com/docs/step-cli/reference/certificate/create/#usage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"time"
)

var (
	ConfigFolder           = "/etc/stepin"
	RootCACertName         = "root_ca"
	IntermediateCACertName = "intermediate_ca"
	LeafCertFolder         = path.Join(ConfigFolder, "leaf")
	KeyFolder              = path.Join(ConfigFolder, "key")
)

var RootCaCrtPath string
var RootCaKeyPath string
var IntermediateCaKeyPath string
var IntermediateCaCrtPath string

func Setup() {
	LeafCertFolder = path.Join(ConfigFolder, "leaf")
	KeyFolder = path.Join(ConfigFolder, "key")
	RootCaCrtPath = path.Join(ConfigFolder, RootCACertName+".crt")
	RootCaKeyPath = path.Join(KeyFolder, RootCACertName+".key")
	IntermediateCaCrtPath = path.Join(ConfigFolder, IntermediateCACertName+".crt")
	IntermediateCaKeyPath = path.Join(KeyFolder, IntermediateCACertName+".key")
}

func init() {
	Setup()
}

func Exec(cmd string, args ...string) (string, error) {
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

type CAConfig struct {
	RootCaName             string
	RootCaPassword         string
	IntermediateCaName     string
	IntermediateCaPassword string
}

func createPasswordFile(filename, password string) (*os.File, func() error, error) {
	passwordFile, err := os.CreateTemp(os.TempDir(), filename)
	if err != nil {
		return nil, nil, err
	}

	_, err = io.WriteString(passwordFile, password)
	if err != nil {
		_ = os.Remove(passwordFile.Name())
		return nil, nil, err
	}

	err = passwordFile.Close()
	if err != nil {
		_ = os.Remove(passwordFile.Name())
		return nil, nil, err
	}

	return passwordFile,
		func() error {
			return os.Remove(passwordFile.Name())
		},
		nil
}

func fileExists(filename string) bool {
	stat, err := os.Stat(filename)
	return err == nil && !stat.IsDir()
}

func mkdirAll(path string, perm os.FileMode) error {
	_, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.MkdirAll(path, perm)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func Initialize(config CAConfig) (err error) {
	err = mkdirAll(LeafCertFolder, 0755)
	if err != nil {
		return err
	}
	err = mkdirAll(KeyFolder, 0755)
	if err != nil {
		return err
	}

	rootCaPasswordFile, rootCaCleanup, err := createPasswordFile("stepin_root_ca_password_*.txt", config.RootCaPassword)
	if err != nil {
		return err
	}
	defer func() {
		_ = rootCaCleanup()
	}()

	intermediateCaPasswordFile, intermediateCaCleanup, err := createPasswordFile("stepin_intermediate_ca_password_*.txt", config.IntermediateCaPassword)
	if err != nil {
		return err
	}
	defer func() {
		_ = intermediateCaCleanup()
	}()

	if fileExists(RootCaCrtPath) {
		err = os.Remove(RootCaCrtPath)
		if err != nil {
			return err
		}
	}
	if fileExists(RootCaKeyPath) {
		err = os.Remove(RootCaKeyPath)
		if err != nil {
			return err
		}
	}

	_, err = Exec(
		"step",
		"certificate",
		"create",
		config.RootCaName,
		RootCaCrtPath,
		RootCaKeyPath,
		"--profile",
		"root-ca",
		"--password-file",
		rootCaPasswordFile.Name(),
	)
	if err != nil {
		return err
	}

	if fileExists(IntermediateCaCrtPath) {
		err = os.Remove(IntermediateCaCrtPath)
		if err != nil {
			return err
		}
	}
	if fileExists(IntermediateCaKeyPath) {
		err = os.Remove(IntermediateCaKeyPath)
		if err != nil {
			return err
		}
	}
	_, err = Exec(
		"step",
		"certificate",
		"create",
		config.IntermediateCaName,
		IntermediateCaCrtPath,
		IntermediateCaKeyPath,
		"--profile",
		"intermediate-ca",
		"--ca",
		RootCaCrtPath,
		"--ca-key",
		RootCaKeyPath,
		"--ca-password-file",
		rootCaPasswordFile.Name(),
		"--password-file",
		intermediateCaPasswordFile.Name(),
	)
	if err != nil {
		return err
	}

	return nil
}

func SignCert(config CAConfig, filename, hostname string, keyType string, expireInHour int) error {
	crtPath := path.Join(LeafCertFolder, filename+".crt")
	keyPath := path.Join(LeafCertFolder, filename+".key")

	intermediateCaPasswordFile, intermediateCaCleanup, err := createPasswordFile("stepin_intermediate_ca_password_*.txt", config.IntermediateCaPassword)
	if err != nil {
		return err
	}
	defer func() {
		_ = intermediateCaCleanup()
	}()

	_, err = Exec(
		"step",
		"certificate",
		"create",
		hostname,
		crtPath,
		keyPath,
		"--profile",
		"leaf",
		"--ca",
		IntermediateCaCrtPath,
		"--ca-key",
		IntermediateCaKeyPath,
		"--ca-password-file",
		intermediateCaPasswordFile.Name(),
		"--bundle",
		"--insecure",
		"--no-password",
		"--kty",
		keyType,
		"--not-after",
		fmt.Sprintf("%dh", expireInHour),
	)
	return err
}

type Cert struct {
	Filename   string
	Inspection string
}

func CertList(withInspection bool) ([]Cert, error) {
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
			var inspection string
			if withInspection {
				inspection, err = InspectCert(name, true)
				if err != nil {
					return nil, err
				}
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
	return Exec(
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
	return fileExists(RootCaCrtPath) &&
		fileExists(RootCaKeyPath) &&
		fileExists(IntermediateCaCrtPath) &&
		fileExists(IntermediateCaKeyPath)
}
