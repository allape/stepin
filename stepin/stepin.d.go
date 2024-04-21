package stepin

type Inspection string

type CommandArguments []string
type DisposeFunc func() error

type CmdOption interface {
	Apply(args CommandArguments) (CommandArguments, error)
}
