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

type CreationOption interface {
	stepin.CmdOption
}

type OptionSubject struct {
	CreationOption
	Subject SubjectName `json:"<subject>"`
	CrtFile CrtFile     `json:"<crt-file>"`
	KeyFile KeyFile     `json:"<key-file>"`
}

func (o OptionSubject) Apply(args stepin.CommandArguments) (stepin.CommandArguments, error) {
	return append([]string{
		string(o.Subject),
		string(o.CrtFile),
		string(o.KeyFile),
	}, args...), nil
}

type OptionKeyType struct {
	CreationOption
	KTY KeyType `json:"--kty"`
}

func (o OptionKeyType) Apply(args stepin.CommandArguments) (stepin.CommandArguments, error) {
	return append(args, "--kty", string(o.KTY)), nil
}

type OptionSize struct {
	CreationOption
	Size BitSize `json:"--size"`
}

func (o OptionSize) Apply(args stepin.CommandArguments) (stepin.CommandArguments, error) {
	return append(args, "--size", fmt.Sprintf("%d", o.Size)), nil
}

type OptionCurve struct {
	CreationOption
	Curve Curve `json:"--crv/--curve"`
}

func (o OptionCurve) Apply(args stepin.CommandArguments) (stepin.CommandArguments, error) {
	return append(args, "--crv", string(o.Curve)), nil
}

type OptionCSR struct {
	CreationOption
	CSR bool `json:"--csr"`
}

func (o OptionCSR) Apply(args stepin.CommandArguments) (stepin.CommandArguments, error) {
	if o.CSR {
		return append(args, "--csr"), nil
	}
	return args, nil
}

type OptionProfile struct {
	CreationOption
	Profile Profile `json:"--profile"`
}

func (o OptionProfile) Apply(args stepin.CommandArguments) (stepin.CommandArguments, error) {
	return append(args, "--profile", string(o.Profile)), nil
}

type OptionTemplate struct {
	CreationOption
	Template FilePath `json:"--template"`
}

func (o OptionTemplate) Apply(args stepin.CommandArguments) (stepin.CommandArguments, error) {
	if o.Template == "" {
		return args, nil
	}
	return append(args, "--template", string(o.Template)), nil
}

type OptionSet struct {
	CreationOption
	Set []Set `json:"--set"`
}

func (o OptionSet) Apply(args stepin.CommandArguments) (stepin.CommandArguments, error) {
	for _, set := range o.Set {
		args = append(args, "--set", fmt.Sprintf("%s=%s", set.Key, set.Value))
	}
	return args, nil
}

type OptionSetFile struct {
	CreationOption
	SetFile FilePath `json:"--set-file"`
}

func (o OptionSetFile) Apply(args stepin.CommandArguments) (stepin.CommandArguments, error) {
	if o.SetFile == "" {
		return args, nil
	}
	return append(args, "--set-file", string(o.SetFile)), nil
}

type OptionNotBefore struct {
	CreationOption
	NotBefore time.Time `json:"--not-before"`
}

func (o OptionNotBefore) Apply(args stepin.CommandArguments) (stepin.CommandArguments, error) {
	return append(args, "--not-before", o.NotBefore.Format(time.RFC3339)), nil
}

type OptionNotAfter struct {
	CreationOption
	NotAfter time.Time `json:"--not-after"`
}

func (o OptionNotAfter) Apply(args stepin.CommandArguments) (stepin.CommandArguments, error) {
	return append(args, "--not-after", o.NotAfter.Format(time.RFC3339)), nil
}

type OptionSAN struct {
	CreationOption
	SAN []SAN `json:"--san"`
}

func (o OptionSAN) Apply(args stepin.CommandArguments) (stepin.CommandArguments, error) {
	for _, record := range o.SAN {
		args = append(args, "--san", string(record))
	}
	return args, nil
}

type OptionCA struct {
	CreationOption
	CA CrtFile `json:"--ca"`
}

func (o OptionCA) Apply(args stepin.CommandArguments) (stepin.CommandArguments, error) {
	if o.CA == "" {
		return args, nil
	}
	return append(args, "--ca", string(o.CA)), nil
}

type OptionCAKMS struct {
	CreationOption
	CAKMS URI `json:"--ca-kms"`
}

func (o OptionCAKMS) Apply(args stepin.CommandArguments) (stepin.CommandArguments, error) {
	return append(args, "--ca-kms", string(o.CAKMS)), nil
}

type OptionCAKey struct {
	CreationOption
	CAKey KeyFile `json:"--ca-key"`
}

func (o OptionCAKey) Apply(args stepin.CommandArguments) (stepin.CommandArguments, error) {
	if o.CAKey == "" {
		return args, nil
	}
	return append(args, "--ca-key", string(o.CAKey)), nil
}

type OptionCAPasswordFile struct {
	CreationOption
	CAPasswordFile PasswordFile `json:"--ca-password-file"`
}

func (o OptionCAPasswordFile) Apply(args stepin.CommandArguments) (stepin.CommandArguments, error) {
	if o.CAPasswordFile == "" {
		return args, nil
	}
	return append(args, "--ca-password-file", string(o.CAPasswordFile)), nil
}

type OptionKMS struct {
	CreationOption
	KMS URI `json:"--kms"`
}

func (o OptionKMS) Apply(args stepin.CommandArguments) (stepin.CommandArguments, error) {
	return append(args, "--kms", string(o.KMS)), nil
}

type OptionKey struct {
	CreationOption
	Key KeyFile `json:"--key"`
}

func (o OptionKey) Apply(args stepin.CommandArguments) (stepin.CommandArguments, error) {
	if o.Key == "" {
		return args, nil
	}
	return append(args, "--key", string(o.Key)), nil
}

type OptionPasswordFile struct {
	CreationOption
	PasswordFile PasswordFile `json:"--password-file"`
}

func (o OptionPasswordFile) Apply(args stepin.CommandArguments) (stepin.CommandArguments, error) {
	if o.PasswordFile == "" {
		return args, nil
	}
	return append(args, "--password-file", string(o.PasswordFile)), nil
}

type OptionNoPassword struct {
	CreationOption
	NoPassword bool `json:"--no-password"`
}

func (o OptionNoPassword) Apply(args stepin.CommandArguments) (stepin.CommandArguments, error) {
	if o.NoPassword {
		return append(args, "--no-password", "--insecure"), nil
	}
	return args, nil
}

type OptionBundle struct {
	CreationOption
	Bundle bool `json:"--bundle"`
}

func (o OptionBundle) Apply(args stepin.CommandArguments) (stepin.CommandArguments, error) {
	if o.Bundle {
		return append(args, "--bundle"), nil
	}
	return args, nil
}

type OptionSkipCSRSignature struct {
	CreationOption
	SkipCSRSignature bool `json:"--skip-csr-signature"`
}

func (o OptionSkipCSRSignature) Apply(args stepin.CommandArguments) (stepin.CommandArguments, error) {
	if o.SkipCSRSignature {
		return append(args, "--skip-csr-signature"), nil
	}
	return args, nil
}

type OptionForce struct {
	CreationOption
	Force bool `json:"-f/--force"`
}

func (o OptionForce) Apply(args stepin.CommandArguments) (stepin.CommandArguments, error) {
	if o.Force {
		return append(args, "-f"), nil
	}
	return args, nil
}

type OptionSubtle struct {
	CreationOption
	Subtle bool `json:"--subtle"`
}

func (o OptionSubtle) Apply(args stepin.CommandArguments) (stepin.CommandArguments, error) {
	if o.Subtle {
		return append(args, "--subtle"), nil
	}
	return args, nil
}
