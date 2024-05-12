package stepin

type Inspection string

type DisposeFunc func() error

type Commander struct {
	Executable string
	Arguments  []string
}

type CommandOption interface {
	Apply(commander *Commander) (*Commander, error)
}

// region options not in the official documentation

type OptionCommandBin struct {
	CommandOption
	CommandBin string `json:"step"`
}

func (o OptionCommandBin) Apply(commander *Commander) (*Commander, error) {
	commander.Executable = o.CommandBin
	return commander, nil
}

// endregion options not in the official documentation
