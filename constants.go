package gomigo

const (
	CheckSchema     = "SELECT EXISTS(SELECT schema_name FROM information_schema.schemata WHERE schema_name = '__migrations');"
	Add             = "SELECT * FROM __migrations.add($1);"
	Remove          = "SELECT * FROM __migrations.remove($1);"
	Up              = "SELECT * FROM __migrations.up($1);"
	Down            = "SELECT * FROM __migrations.down($1);"
	DiffUp          = "SELECT * FROM __migrations.get_diff_up($1, $2);"
	DiffDown        = "SELECT * FROM __migrations.get_diff_down($1, $2);"
	CurrentVersion  = "SELECT * FROM __migrations.get_current_version();"
	Layout          = "20060102150405"
	MigrateTemplate = `package m%[1]s

import (
	"github.com/gaarutyunov/gomigo"
	"github.com/markbates/pkger"
	log "github.com/sirupsen/logrus"
)

func Up(conn *gomigo.Migrator) {
	f, err := pkger.Open("/migrations/%[1]s/up.sql")

	if err != nil {
		log.Fatalln(err)
	}

	v, err := conn.Up(&gomigo.Migration{Name: "%[1]s"}, f)
	
	if err != nil {
		log.Fatalln(err)
	}

	log.Infof("upgraded to version: %%d", v)
}

func Down(conn *gomigo.Migrator) {
	f, err := pkger.Open("/migrations/%[1]s/down.sql")

	if err != nil {
		log.Fatalln(err)
	}

	v, err := conn.Down(&gomigo.Migration{Name: "%[1]s"}, f)

	if err != nil {
		log.Fatalln(err)
	}

	log.Infof("downgraded to version: %%d", v)
}`
)
