package gomigo

import (
	"github.com/jackc/pgx"
	"github.com/markbates/pkger/pkging"
)

func ExecFileTx(file pkging.File, tx *pgx.Tx) error {
	stat, err := file.Stat()

	if err != nil {
		return err
	}

	bytes := make([]byte, stat.Size())

	_, err = file.Read(bytes)

	if err != nil {
		return err
	}

	if _, err := tx.Exec(string(bytes)); err != nil {
		return err
	}

	return nil
}

func ExecFile(file pkging.File, conn *pgx.Conn) error {
	stat, err := file.Stat()

	if err != nil {
		return err
	}

	bytes := make([]byte, stat.Size())

	_, err = file.Read(bytes)

	if err != nil {
		return err
	}

	if _, err := conn.Exec(string(bytes)); err != nil {
		return err
	}

	return nil
}
