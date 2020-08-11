package gomigo

import (
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx"
	"github.com/markbates/pkger"
	"github.com/markbates/pkger/pkging"
	log "github.com/sirupsen/logrus"
	"golang.org/x/tools/imports"
	"os"
	"path/filepath"
	"strings"
)

type MigratorConfig struct {
	Module  string
	ConnStr string
}

type Migrator struct {
	conn    *pgx.Conn
	config  *MigratorConfig
}

func Connect(config *MigratorConfig) (*Migrator, error) {
	conn, err := connect(config.ConnStr)

	if err != nil {
		return nil, err
	}

	return &Migrator{conn: conn, config: config}, nil
}

func (m *Migrator) WithOptions(options *MigratorConfig) {
	m.config = options
}

func (m *Migrator) Close() error {
	return m.conn.Close()
}

func (m *Migrator) CheckSchema() (bool, error) {
	var exists bool

	if err := m.conn.QueryRow(CheckSchema).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (m *Migrator) Init() error {
	exists, err := m.CheckSchema()

	if err != nil {
		return err
	}

	if !exists {
		f, err := pkger.Open("/sql/init.sql")

		if err != nil {
			return fmt.Errorf("error initializing migrations: %v", err)
		}

		err = ExecFile(f, m.conn)

		if err != nil {
			return err
		}

		log.Info("initialized migrations")
	} else {
		log.Info("migrations already initialized")
	}

	return nil
}

func (m *Migrator) Clean() error {
	exists, err := m.CheckSchema()

	if err != nil {
		return err
	}

	if exists {
		f, err := pkger.Open("/sql/clean.sql")

		if err != nil {
			return fmt.Errorf("error cleaning up migrations: %v", err)
		}

		err = ExecFile(f, m.conn)

		if err != nil {
			return err
		}

		log.Info("cleaned migrations")
	} else {
		log.Info("migrations already cleaned up")
	}

	return nil
}

func (m *Migrator) Update() error {

	return nil
}

func (m *Migrator) Add(name string, dir string) error {
	migration := &Migration{Name: ConstructName(name)}

	if name == "" {
		return fmt.Errorf("migration name cannot be empty")
	}

	if dir == "" {
		temp, err := os.Getwd()

		if err != nil {
			return fmt.Errorf("error determining current dir: %v", err)
		}

		dir = temp
	}

	tx, err := m.conn.Begin()

	if err != nil {
		return err
	}

	if err = doAdd(migration, dir, tx); err != nil {
		_ = tx.Rollback()

		return err
	}

	return nil
}

func (m *Migrator) Remove(name string, dir string) error {
	if _, err := m.conn.Exec(Remove, name); err != nil {
		return err
	}

	dirName := filepath.Join(dir, name)

	return os.RemoveAll(dirName)
}

func (m *Migrator) UpV(version int) (int, error) {
	var oldV int
	var diff []string
	var newV int
	tx, err := m.conn.Begin()

	if err != nil {
		return -1, err
	}

	if err := tx.QueryRow(CurrentVersion).Scan(&oldV); err != nil {
		return -1, err
	}

	if err := tx.QueryRow(Diff, oldV, version).Scan(&diff); err != nil {
		return oldV, err
	}

	f, err := os.Create("main.go")

	if err != nil {
		return oldV, err
	}

	g := &generator{
		ConnStr:    m.config.ConnStr,
		Module:     m.config.Module,
		Migrations: diff,
		Function:   "Up",
		File:       f,
	}

	g.Generate()

	out, err := g.Build()

	if err != nil {
		log.Errorln(string(out))

		return oldV, err
	}

	log.Debugln(out)

	return newV, nil
}

func (m *Migrator) Up(migration *Migration, file pkging.File) (int, error) {
	var oldV int
	var v int
	tx, err := m.conn.Begin()

	if err != nil {
		return -1, err
	}

	if err := tx.QueryRow(CurrentVersion).Scan(&oldV); err != nil {
		return -1, err
	}

	if err := tx.QueryRow(Up, migration.Name).Scan(&v); err != nil {
		return oldV, fmt.Errorf("error upgrading to version %d: %v", v, err)
	}

	if err := ExecFileTx(file, tx); err != nil {
		_ = tx.Rollback()

		return oldV, fmt.Errorf("error upgrading to version %d: %v", v, err)
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()

		return oldV, err
	}

	log.Infof("new version: %d", v)

	return v, nil
}

func (m *Migrator) Down(migration *Migration, file pkging.File) (int, error) {
	var oldV int
	var v int

	tx, err := m.conn.Begin()

	if err != nil {
		return -1, err
	}

	if err := tx.QueryRow(CurrentVersion).Scan(&oldV); err != nil {
		return -1, err
	}

	if err := tx.QueryRow(Down, migration.Name).Scan(&v); err != nil {
		return oldV, fmt.Errorf("error downgrading to version %d: %v", v, err)
	}

	if err := ExecFileTx(file, tx); err != nil {
		err = tx.Rollback()

		return oldV, fmt.Errorf("error downgrading to version %d: %v", v, err)
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()

		return oldV, err
	}

	log.Infof("new version: %d", v)

	return v, nil
}

func doAdd(migration *Migration, dir string, tx *pgx.Tx) error {
	var str string

	if err := tx.QueryRow(Add, migration.Name).Scan(&str); err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(str), migration); err != nil {
		return err
	}

	dirName := filepath.Join(dir, migration.Name)

	d, err := os.Open(dir)

	if err != nil {
		return err
	}

	names, err := d.Readdirnames(-1)

	for _, name := range names {
		if strings.Split(name, "_")[1] == strings.Split(migration.Name, "_")[1] {
			return fmt.Errorf("migration %s already exists", name)
		}
	}

	if err = d.Close(); err != nil {
		return err
	}

	if err := os.MkdirAll(dirName, 0755); err != nil {
		return err
	}

	migrate := filepath.Join(dirName, "migrate.go")

	f, err := os.Create(migrate)

	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(f, MigrateTemplate, migration.Name); err != nil {
		return err
	}

	if _, err := imports.Process(f.Name(), nil, nil); err != nil {
		return err
	}

	if _, err = os.Create(dirName + "/up.sql"); err != nil {
		return err
	}

	if _, err = os.Create(dirName + "/down.sql"); err != nil {
		return err
	}

	return tx.Commit()
}
