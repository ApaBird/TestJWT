package postgres

import "fmt"

func (conn *PostgresConnect) SaveToken(token string, guid string) error {
	query := `UPDATE users SET refresh_token = $1 WHERE id = $2 RETURNING id;`
	rows, err := conn.db.Query(query, token, guid)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		return fmt.Errorf("no users with guid %s", guid)
	}
	return nil
}

func (conn *PostgresConnect) GetToken(guid string) (string, error) {
	query := `SELECT refresh_token FROM users WHERE id = $1;`
	var token string
	err := conn.db.QueryRow(query, guid).Scan(&token)
	if err != nil {
		return "", err
	}
	return token, nil
}
