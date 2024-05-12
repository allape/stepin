package create

// https://smallstep.com/docs/step-cli/reference/certificate/create/#usage

import (
	"fmt"
	"github.com/allape/stepin/stepin"
	"time"
)

type Password string

type Crt []byte // Certificate content
type Key []byte // Private key content

type Profile string

const (
	RootCA         Profile = "root-ca"
	IntermediateCA Profile = "intermediate-ca"
	Leaf           Profile = "leaf"
	SelfSigned     Profile = "self-signed"
)

var AllProfiles = []Profile{
	RootCA,
	IntermediateCA,
	Leaf,
	SelfSigned,
}

type KeyType string

const (
	EC  KeyType = "EC"
	OKP KeyType = "OKP"
	RSA KeyType = "RSA"
)

var AllKeyTypes = []KeyType{
	EC,
	OKP,
	RSA,
}

type Curve string

const (
	P256    Curve = "P-256"
	P384    Curve = "P-384"
	P521    Curve = "P-521"
	Ed25519 Curve = "Ed25519"
)

type Set struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type (
	SubjectName string
	BitSize     uint
	SAN         string
	URI         string
)

type (
	FilePath     string
	CrtFile      FilePath
	KeyFile      FilePath
	PasswordFile FilePath
)

type OptionSubject struct {
	stepin.CommandOption
	Subject SubjectName `json:"<subject>"`
	CrtFile CrtFile     `json:"<crt-file>"`
	KeyFile KeyFile     `json:"<key-file>"`
}

func (o OptionSubject) Apply(commander *stepin.Commander) (*stepin.Commander, error) {
	commander.Arguments = append([]string{
		string(o.Subject),
		string(o.CrtFile),
		string(o.KeyFile),
	}, commander.Arguments...)

	return commander, nil
}

type OptionKeyType struct {
	stepin.CommandOption
	KTY KeyType `json:"--kty"`
}

func (o OptionKeyType) Apply(commander *stepin.Commander) (*stepin.Commander, error) {
	commander.Arguments = append(commander.Arguments, "--kty", string(o.KTY))
	return commander, nil
}

type OptionSize struct {
	stepin.CommandOption
	Size BitSize `json:"--size"`
}

func (o OptionSize) Apply(commander *stepin.Commander) (*stepin.Commander, error) {
	commander.Arguments = append(commander.Arguments, "--size", fmt.Sprintf("%d", o.Size))
	return commander, nil
}

type OptionCurve struct {
	stepin.CommandOption
	Curve Curve `json:"--crv/--curve"`
}

func (o OptionCurve) Apply(commander *stepin.Commander) (*stepin.Commander, error) {
	commander.Arguments = append(commander.Arguments, "--crv", string(o.Curve))
	return commander, nil
}

type OptionCSR struct {
	stepin.CommandOption
	CSR bool `json:"--csr"`
}

func (o OptionCSR) Apply(commander *stepin.Commander) (*stepin.Commander, error) {
	if o.CSR {
		commander.Arguments = append(commander.Arguments, "--csr")
	}
	return commander, nil
}

type OptionProfile struct {
	stepin.CommandOption
	Profile Profile `json:"--profile"`
}

func (o OptionProfile) Apply(commander *stepin.Commander) (*stepin.Commander, error) {
	commander.Arguments = append(commander.Arguments, "--profile", string(o.Profile))
	return commander, nil
}

type OptionTemplate struct {
	stepin.CommandOption
	Template FilePath `json:"--template"`
}

func (o OptionTemplate) Apply(commander *stepin.Commander) (*stepin.Commander, error) {
	if o.Template != "" {
		commander.Arguments = append(commander.Arguments, "--template", string(o.Template))
	}
	return commander, nil
}

type OptionSet struct {
	stepin.CommandOption
	Set []Set `json:"--set"`
}

func (o OptionSet) Apply(commander *stepin.Commander) (*stepin.Commander, error) {
	for _, set := range o.Set {
		commander.Arguments = append(commander.Arguments, "--set", fmt.Sprintf("%s=%s", set.Key, set.Value))
	}
	return commander, nil
}

type OptionSetFile struct {
	stepin.CommandOption
	SetFile FilePath `json:"--set-file"`
}

func (o OptionSetFile) Apply(commander *stepin.Commander) (*stepin.Commander, error) {
	if o.SetFile != "" {
		commander.Arguments = append(commander.Arguments, "--set-file", string(o.SetFile))
	}
	return commander, nil
}

type OptionNotBefore struct {
	stepin.CommandOption
	NotBefore time.Time `json:"--not-before"`
}

func (o OptionNotBefore) Apply(commander *stepin.Commander) (*stepin.Commander, error) {
	commander.Arguments = append(commander.Arguments, "--not-before", o.NotBefore.Format(time.RFC3339))
	return commander, nil
}

type OptionNotAfter struct {
	stepin.CommandOption
	NotAfter time.Time `json:"--not-after"`
}

func (o OptionNotAfter) Apply(commander *stepin.Commander) (*stepin.Commander, error) {
	commander.Arguments = append(commander.Arguments, "--not-after", o.NotAfter.Format(time.RFC3339))
	return commander, nil
}

type OptionSAN struct {
	stepin.CommandOption
	SAN []SAN `json:"--san"`
}

func (o OptionSAN) Apply(commander *stepin.Commander) (*stepin.Commander, error) {
	for _, record := range o.SAN {
		commander.Arguments = append(commander.Arguments, "--san", string(record))
	}
	return commander, nil
}

type OptionCA struct {
	stepin.CommandOption
	CA CrtFile `json:"--ca"`
}

func (o OptionCA) Apply(commander *stepin.Commander) (*stepin.Commander, error) {
	if o.CA != "" {
		commander.Arguments = append(commander.Arguments, "--ca", string(o.CA))
	}
	return commander, nil
}

type OptionCAKMS struct {
	stepin.CommandOption
	CAKMS URI `json:"--ca-kms"`
}

func (o OptionCAKMS) Apply(commander *stepin.Commander) (*stepin.Commander, error) {
	commander.Arguments = append(commander.Arguments, "--ca-kms", string(o.CAKMS))
	return commander, nil
}

type OptionCAKey struct {
	stepin.CommandOption
	CAKey KeyFile `json:"--ca-key"`
}

func (o OptionCAKey) Apply(commander *stepin.Commander) (*stepin.Commander, error) {
	if o.CAKey != "" {
		commander.Arguments = append(commander.Arguments, "--ca-key", string(o.CAKey))
	}
	return commander, nil
}

type OptionCAPasswordFile struct {
	stepin.CommandOption
	CAPasswordFile PasswordFile `json:"--ca-password-file"`
}

func (o OptionCAPasswordFile) Apply(commander *stepin.Commander) (*stepin.Commander, error) {
	if o.CAPasswordFile != "" {
		commander.Arguments = append(commander.Arguments, "--ca-password-file", string(o.CAPasswordFile))
	}
	return commander, nil
}

type OptionKMS struct {
	stepin.CommandOption
	KMS URI `json:"--kms"`
}

func (o OptionKMS) Apply(commander *stepin.Commander) (*stepin.Commander, error) {
	commander.Arguments = append(commander.Arguments, "--kms", string(o.KMS))
	return commander, nil
}

type OptionKey struct {
	stepin.CommandOption
	Key KeyFile `json:"--key"`
}

func (o OptionKey) Apply(commander *stepin.Commander) (*stepin.Commander, error) {
	if o.Key != "" {
		commander.Arguments = append(commander.Arguments, "--key", string(o.Key))
	}
	return commander, nil
}

type OptionPasswordFile struct {
	stepin.CommandOption
	PasswordFile PasswordFile `json:"--password-file"`
}

func (o OptionPasswordFile) Apply(commander *stepin.Commander) (*stepin.Commander, error) {
	if o.PasswordFile != "" {
		commander.Arguments = append(commander.Arguments, "--password-file", string(o.PasswordFile))
	}
	return commander, nil
}

type OptionNoPassword struct {
	stepin.CommandOption
	NoPassword bool `json:"--no-password"`
}

func (o OptionNoPassword) Apply(commander *stepin.Commander) (*stepin.Commander, error) {
	if o.NoPassword {
		commander.Arguments = append(commander.Arguments, "--no-password", "--insecure")
	}
	return commander, nil
}

type OptionBundle struct {
	stepin.CommandOption
	Bundle bool `json:"--bundle"`
}

func (o OptionBundle) Apply(commander *stepin.Commander) (*stepin.Commander, error) {
	if o.Bundle {
		commander.Arguments = append(commander.Arguments, "--bundle")
	}
	return commander, nil
}

type OptionSkipCSRSignature struct {
	stepin.CommandOption
	SkipCSRSignature bool `json:"--skip-csr-signature"`
}

func (o OptionSkipCSRSignature) Apply(commander *stepin.Commander) (*stepin.Commander, error) {
	if o.SkipCSRSignature {
		commander.Arguments = append(commander.Arguments, "--skip-csr-signature")
	}
	return commander, nil
}

type OptionForce struct {
	stepin.CommandOption
	Force bool `json:"-f/--force"`
}

func (o OptionForce) Apply(commander *stepin.Commander) (*stepin.Commander, error) {
	if o.Force {
		commander.Arguments = append(commander.Arguments, "-f")
	}
	return commander, nil
}

type OptionSubtle struct {
	stepin.CommandOption
	Subtle bool `json:"--subtle"`
}

func (o OptionSubtle) Apply(commander *stepin.Commander) (*stepin.Commander, error) {
	if o.Subtle {
		commander.Arguments = append(commander.Arguments, "--subtle")
	}
	return commander, nil
}
