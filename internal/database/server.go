package database

import (
	"context"
	"database/sql"
	"strings"
)

const SERVER_TABLE = `CREATE TABLE IF NOT EXISTS server (
			id INTEGER PRIMARY KEY AUTOINCREMENT, 
			address TEXT NOT NULL, 
			port INTEGER NOT NULL, 
			user TEXT NOT NULL,
			password TEXT NOT NULL,  
			name TEXT NOT NULL
			)`

type Server struct {
	Address  string
	Port     uint16
	User     string
	Password string
	Name     string
}

type ServerDbRow struct {
	ID int
	Server
}

type ServerOrDir struct {
	Name   string
	Server ServerDbRow
	Dir    bool
	Childs []*ServerOrDir
}

func (db *Database) AddServer(s *Server) (int64, error) {

	result, err := db.conn.ExecContext(
		context.Background(),
		`INSERT INTO server (address, port, user, password, name) VALUES (?,?,?,?,?);`, s.Address, s.Port, s.User, s.Password, s.Name,
	)

	if err != nil {
		return -1, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return -1, err
	}

	return id, nil
}

func (db *Database) DeleteServer(ID int) error {

	_, err := db.conn.ExecContext(
		context.Background(),
		`DELETE FROM server WHERE id == ?;`, ID,
	)

	if err != nil {
		return err
	}

	return nil
}

func (db *Database) ServerList() ([]ServerDbRow, error) {
	var servers []ServerDbRow

	rows, err := db.conn.QueryContext(
		context.Background(),
		`SELECT * FROM server;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var server ServerDbRow

		if err := rows.Scan(
			&server.ID, &server.Address, &server.Port, &server.User, &server.Password, &server.Name,
		); err != nil {
			return nil, err
		}
		servers = append(servers, server)
	}

	return servers, nil
}

func (db *Database) ServerListWithDirs() ([]*ServerOrDir, error) {
	root := []*ServerOrDir{}
	dirMap := map[string]*ServerOrDir{}
	rows, err := db.conn.QueryContext(
		context.Background(),
		`SELECT * FROM server ORDER by name ASC;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var server ServerDbRow
		if err := rows.Scan(
			&server.ID, &server.Address, &server.Port, &server.User, &server.Password, &server.Name,
		); err != nil {
			return nil, err
		}
		name_s := strings.Split(server.Name, "/")
		if len(name_s) == 1 {
			root = append(root, &ServerOrDir{Name: server.Name, Server: server, Dir: false, Childs: nil})
		} else {
			dir_level := len(name_s) - 1
			var child *ServerOrDir = nil
			for i := range dir_level {
				dir_path := strings.Join(name_s[:dir_level-i], "/")
				dir, ok := dirMap[dir_path]

				if !ok {
					dir = &ServerOrDir{Name: name_s[dir_level-i-1], Dir: true, Childs: nil}
					dirMap[dir_path] = dir
				}

				if child != nil {
					dir.Childs = append(dir.Childs, child)
				} else {
					dir.Childs = append(dir.Childs, &ServerOrDir{Name: name_s[dir_level], Server: server, Dir: false, Childs: nil})
				}

				if !ok {
					child = dir
				} else {
					child = nil
					break
				}
			}
			if child != nil {
				root = append(root, child)
			}
		}
	}
	return root, nil
}

func (db *Database) GetServer(ID int) (ServerDbRow, error) {
	row := db.conn.QueryRow("SELECT * FROM server WHERE id = ?", ID)
	var server ServerDbRow
	err := row.Scan(&server.ID, &server.Address, &server.Port, &server.User, &server.Password, &server.Name)
	if err != nil {
		return ServerDbRow{}, err
	}
	return server, nil
}

func (db *Database) IsNameUnique(name string) (bool, error) {
	row := db.conn.QueryRow("SELECT id FROM server WHERE name = ?", name)
	var id int
	err := row.Scan(&id)
	if err == sql.ErrNoRows {
		return true, nil
	} else if err != nil {
		return false, err
	}
	return false, nil
}
