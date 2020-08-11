package gomigo

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
)

type generator struct {
	ConnStr    string
	Module     string
	Migrations []string
	Function   string
	File       *os.File
}

func (g *generator) Generate() {
	f := g.File

	_, _ = fmt.Fprintln(f, "package main")
	_, _ = fmt.Fprintln(f)
	_, _ = fmt.Fprintln(f, "import (")
	_, _ = fmt.Fprintln(f, "\t\"github.com/gaarutyunov/gomigo\"")
	_, _ = fmt.Fprintln(f, "\tlog \"github.com/sirupsen/logrus\"")

	for _, module := range g.Migrations {
		_, _ = fmt.Fprintf(f, "\tm%[1]s \"%s/migrations/%[1]s\"\n", module, g.Module)
	}

	_, _ = fmt.Fprintln(f, ")")
	_, _ = fmt.Fprintln(f)
	_, _ = fmt.Fprintln(f, "func main() {")
	_, _ = fmt.Fprintf(f, `
	conn, err := gomigo.Connect(&gomigo.MigratorConfig{
		Module:  "%s",
		ConnStr: "%s",
	})
`, g.Module, g.ConnStr)
	_, _ = fmt.Fprintln(f)
	_, _ = fmt.Fprint(f, `
	if err != nil {
		log.Fatalln(err)
	}

`)

	for _, module := range g.Migrations {
		_, _ = fmt.Fprintf(f, "\tm%s.%s(conn)\n", module, g.Function)
	}
	_, _ = fmt.Fprintln(f, "}")
}

func (g *generator) Build() ([]byte, error) {
	cmd := exec.Command("pkger")

	if out, err := cmd.CombinedOutput(); err != nil {
		log.Errorln(string(out))

		return nil, err
	}

	cmd = exec.Command("go", "build", g.File.Name())

	if out, err := cmd.CombinedOutput(); err != nil {
		log.Errorln(string(out))

		return nil, err
	}

	cmd = exec.Command("./main")

	return cmd.CombinedOutput()
}
