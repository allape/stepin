package inspect

import "github.com/allape/stepin/stepin"

func Inspect(filename string, short bool, options ...stepin.CommandOption) (stepin.Inspection, error) {
	args := []string{
		"certificate",
		"inspect",
		filename,
	}
	if short {
		args = append(args, "--short")
	}

	var err error
	commander := &stepin.Commander{
		Executable: "step",
		Arguments:  args,
	}

	for _, option := range options {
		commander, err = option.Apply(commander)
		if err != nil {
			return "", err
		}
	}

	output, err := stepin.Exec(
		commander.Executable,
		commander.Arguments...,
	)
	return stepin.Inspection(output), err
}
