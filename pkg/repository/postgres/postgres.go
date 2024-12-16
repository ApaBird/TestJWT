package postgres

import "database/sql"

type PostgresConnect struct {
	db *sql.DB
}

type PostgresConfig struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
	Sslmode  string
}

func NewPostgresConnect(conf *PostgresConfig) (*PostgresConnect, error) {
	db, err := sql.Open("postgres", prepareConnectionString(conf))
	if err != nil {
		return nil, err
	}

	return &PostgresConnect{db: db}, nil
}

func prepareConnectionString(config *PostgresConfig) string {
	return "host=" + config.Host +
		" port=" + config.Port +
		" user=" + config.User +
		" password=" + config.Password +
		" dbname=" + config.Database +
		" sslmode=" + config.Sslmode
}

func (p *PostgresConnect) Close() error {
	return p.db.Close()
}
