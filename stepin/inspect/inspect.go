package inspect

import "github.com/allape/stepin/stepin"

func Inspect(filename string, short bool) (stepin.Inspection, error) {
	args := []string{
		"certificate",
		"inspect",
		filename,
	}
	if short {
		args = append(args, "--short")
	}
	output, err := stepin.Exec(
		"step",
		args...,
	)
	return stepin.Inspection(output), err
}
