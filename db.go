package gomigo

import (
	"github.com/jackc/pgx"
	"github.com/sirupsen/logrus"
)

type logAdapter struct {
}

func (la *logAdapter) Log(level pgx.LogLevel, _ string, data map[string]interface{}) {
	if sql, ok := data["sql"]; ok {
		switch level {
		case pgx.LogLevelInfo:
			logrus.Infof("Executing SQL: %s", sql)
		case pgx.LogLevelError:
			if err, ok := data["err"]; ok {
				logrus.Errorf("Error executing %s %v", sql, err)
			}
		}
	}
}

func connect(connStr string) (*pgx.Conn, error) {
	config, err := pgx.ParseConnectionString(connStr)

	if err != nil {
		return nil, err
	}

	config.Logger = &logAdapter{}

	conn, err := pgx.Connect(config)

	if err != nil {
		return nil, err
	}

	return conn, nil
}
