package create

// https://smallstep.com/docs/step-cli/reference/certificate/create/#usage

import (
	"github.com/allape/stepin/stepin"
	"github.com/allape/stepin/stepin/inspect"
	"os"
)

func New(subject OptionSubject, options ...stepin.CommandOption) (stepin.Inspection, error) {
	commander, err := subject.Apply(&stepin.Commander{
		Executable: "step",
		Arguments:  nil,
	})
	if err != nil {
		return "", err
	}

	for _, option := range options {
		commander, err = option.Apply(commander)
		if err != nil {
			return "", err
		}
	}

	_, err = stepin.Exec(
		commander.Executable,
		append([]string{
			"certificate",
			"create",
		}, commander.Arguments...)...,
	)
	if err != nil {
		return "", err
	}

	return inspect.Inspect(string(subject.CrtFile), false, stepin.OptionCommandBin{
		CommandBin: commander.Executable,
	})
}

type PrimaryOptions struct {
	Subject  SubjectName `json:"subject"`
	Password Password    `json:"password"`
}

func NewRaw(opt PrimaryOptions, options ...stepin.CommandOption) (stepin.Inspection, Crt, Key, error) {
	subject := opt.Subject
	password := opt.Password
	passFilePath := PasswordFile("")
	if password != "" {
		passFile, disposePassFile, err := stepin.NewTmpFile("stepin_password_*.txt", []byte(password))
		if err != nil {
			return "", nil, nil, err
		}
		defer func() {
			_ = disposePassFile()
		}()
		_ = passFile.Close()
		passFilePath = PasswordFile(passFile.Name())
	}

	certFile, disposeCertFile, err := stepin.NewTmpFile("stepin_cert_*.crt", nil)
	if err != nil {
		return "", nil, nil, err
	}
	defer func() {
		_ = disposeCertFile()
	}()
	_ = certFile.Close()

	keyFile, disposeKeyFile, err := stepin.NewTmpFile("stepin_key_*.key", nil)
	if err != nil {
		return "", nil, nil, err
	}
	defer func() {
		_ = disposeKeyFile()
	}()
	_ = keyFile.Close()

	inspection, err := New(OptionSubject{
		Subject: subject,
		CrtFile: CrtFile(certFile.Name()),
		KeyFile: KeyFile(keyFile.Name()),
	}, append(
		options,
		OptionPasswordFile{PasswordFile: passFilePath},
		OptionForce{Force: true},
	)...)

	if err != nil {
		return inspection, nil, nil, err
	}

	crt, err := os.ReadFile(certFile.Name())
	if err != nil {
		return inspection, nil, nil, err
	}
	key, err := os.ReadFile(keyFile.Name())
	if err != nil {
		return inspection, nil, nil, err
	}
	return inspection, crt, key, nil
}

type RootOptions struct {
	PrimaryOptions
}

func NewRootCA(
	opt RootOptions,
	options ...stepin.CommandOption,
) (stepin.Inspection, Crt, Key, error) {
	return NewRaw(opt.PrimaryOptions, append(options, OptionProfile{Profile: RootCA})...)
}

type RootlessOptions struct {
	PrimaryOptions
	RootCaCrt    Crt      `json:"rootCaCrt"`
	RootCaKey    Key      `json:"rootCaKey"`
	RootPassword Password `json:"rootPassword"`
}

func NewRootless(
	opt RootlessOptions,
	options ...stepin.CommandOption,
) (stepin.Inspection, Crt, Key, error) {
	rootCaCrtFile, disRCCF, err := stepin.NewTmpFile("stepin_root_ca_crt_*.txt", opt.RootCaCrt)
	if err != nil {
		return "", nil, nil, err
	}
	defer func() {
		_ = disRCCF()
	}()
	_ = rootCaCrtFile.Close()

	rootCaKeyFile, disRCKF, err := stepin.NewTmpFile("stepin_root_ca_key_*.txt", opt.RootCaKey)
	if err != nil {
		return "", nil, nil, err
	}
	defer func() {
		_ = disRCKF()
	}()
	_ = rootCaKeyFile.Close()

	rootCaPasswordFile, disRCPF, err := stepin.NewTmpFile("stepin_root_ca_password_*.txt", []byte(opt.RootPassword))
	if err != nil {
		return "", nil, nil, err
	}
	defer func() {
		_ = disRCPF()
	}()
	_ = rootCaPasswordFile.Close()

	return NewRaw(
		opt.PrimaryOptions,
		append(
			options,
			OptionCA{CA: CrtFile(rootCaCrtFile.Name())},
			OptionCAKey{CAKey: KeyFile(rootCaKeyFile.Name())},
			OptionCAPasswordFile{CAPasswordFile: PasswordFile(rootCaPasswordFile.Name())},
		)...,
	)
}

func NewIntermediateCA(
	opt RootlessOptions,
	options ...stepin.CommandOption,
) (stepin.Inspection, Crt, Key, error) {
	return NewRootless(opt, append(options, OptionProfile{Profile: IntermediateCA})...)
}

func NewLeaf(
	opt RootlessOptions,
	options ...stepin.CommandOption,
) (stepin.Inspection, Crt, Key, error) {
	return NewRootless(opt, append(options, OptionProfile{Profile: Leaf})...)
}

// NewTLS
// Create a new leaf certificate and key suitable for use in a TLS server.
// Example:
//
//		NewTLS(
//			RootlessOptions{
//				PrimaryOptions: PrimaryOptions{
//					Subject: SubjectName("SOME HOSTNAME")
//				},
//				RootCaCrt:    []byte("ROOT OR INTERMEDIATE CA CERTIFICATE"),
//				RootCaKey:    []byte("ROOT OR INTERMEDIATE CA PRIVATE KEY"),
//				RootPassword: Password("ROOT OR INTERMEDIATE CA PASSWORD"),
//			},
//	     OptionNotAfter{NotAfter: time.Now().AddDate(1, 0, 0)},
//	     OptionKeyType{KeyType: RSA},
//		)
func NewTLS(
	opt RootlessOptions,
	options ...stepin.CommandOption,
) (stepin.Inspection, Crt, Key, error) {
	return NewLeaf(
		opt,
		append(
			options,
			OptionBundle{Bundle: true},
			OptionNoPassword{NoPassword: true},
		)...,
	)
}
