package postgres

import "database/sql"

func ManuelMigrator(con *sql.DB) error {
	// Create the ticket_options table
	_, err := con.Exec(`CREATE TABLE IF NOT EXISTS ticket_options (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT NOT NULL,
		allocation INT NOT NULL
	)`)
	if err != nil {
		return err
	}

	// Create the purchases table
	_, err = con.Exec(`CREATE TABLE IF NOT EXISTS purchases (
		id SERIAL PRIMARY KEY,
		quantity INT NOT NULL,
		user_id TEXT NOT NULL,
		ticket_option_id INT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return err
	}
	return nil
}
