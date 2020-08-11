package gomigo

import "flag"

type Args struct {
	Command string
	Version int
	Name    string
	Module  string
}

func Parse(args *Args) {
	var version = flag.Int("version", -1, "version to upgrade/downgrade")
	var name = flag.String("name", "", "migration name")
	var module = flag.String("module", "", "module name")

	flag.Parse()

	command := flag.Arg(0)

	args.Command = command
	args.Version = *version
	args.Name = *name
	args.Module = *module
}
