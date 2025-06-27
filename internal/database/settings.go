package database

import (
	"context"
	"database/sql"
	"potatossh/internal/theme"
)

const SETTINGS_TABLE = `CREATE TABLE IF NOT EXISTS settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			theme TEXT NOT NULL, 
			fontSize INTEGER NOT NULL, 
			openInNewWindow INTEGER NOT NULL
			)`

type Settings struct {
	Theme           *theme.Theme
	FontSize        uint
	OpenInNewWindow bool
}

var settings_id int64

func (db *Database) UpdateSettings(s *Settings) (int64, error) {
	result, err := db.conn.ExecContext(
		context.Background(),
		`UPDATE settings SET theme = ?, fontSize = ?, openInNewWindow = ? WHERE id = ?`, s.Theme.Name, s.FontSize, s.OpenInNewWindow, settings_id)
	if err != nil {
		return -1, err
	}
	return result.RowsAffected()
}

func (db *Database) GetSettings(defaultSettings Settings, themes []theme.Theme) (Settings, error) {
	var theme_name string
	settings := defaultSettings
	row := db.conn.QueryRow("SELECT id, theme, fontSize, openInNewWindow FROM settings")
	err := row.Scan(&settings_id, &theme_name, &settings.FontSize, &settings.OpenInNewWindow)
	if err == sql.ErrNoRows {
		res, err := db.conn.ExecContext(
			context.Background(),
			`INSERT INTO settings (theme, fontSize, openInNewWindow) VALUES (?,?,?)`, settings.Theme.Name, settings.FontSize, settings.OpenInNewWindow)

		if err != nil {
			return settings, err
		}
		settings_id, err = res.LastInsertId()
		if err != nil {
			return settings, err
		}
		return settings, nil
	} else if err != nil {
		return settings, err
	}
	for i, t := range themes {
		if t.Name == theme_name {
			settings.Theme = &themes[i]
		}
	}

	return settings, nil
}
