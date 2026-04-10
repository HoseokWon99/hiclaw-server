package storage

import (
	"database/sql"
	"fmt"

	"hiclaw-server/core"

	_ "modernc.org/sqlite"
)

type SQLiteDeviceStore struct {
	db *sql.DB
}

func NewSQLiteDeviceStore(path string) (*SQLiteDeviceStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS devices (
		ip TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		is_agent BOOLEAN NOT NULL DEFAULT FALSE
	)`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create table: %w", err)
	}
	return &SQLiteDeviceStore{db: db}, nil
}

func (s *SQLiteDeviceStore) Register(d *core.Device) error {
	_, err := s.db.Exec(
		`INSERT INTO devices (ip, name, is_agent) VALUES (?, ?, ?)
		 ON CONFLICT(ip) DO UPDATE SET name = excluded.name, is_agent = excluded.is_agent`,
		d.IP, d.Name, d.IsAgent,
	)
	return err
}

func (s *SQLiteDeviceStore) Remove(ip string) error {
	_, err := s.db.Exec(`DELETE FROM devices WHERE ip = ?`, ip)
	return err
}

func (s *SQLiteDeviceStore) ListAll() ([]*core.Device, error) {
	rows, err := s.db.Query(`SELECT ip, name, is_agent FROM devices`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []*core.Device
	for rows.Next() {
		d := &core.Device{}
		if err := rows.Scan(&d.IP, &d.Name, &d.IsAgent); err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

func (s *SQLiteDeviceStore) GetAgent() (*core.Device, error) {
	d := &core.Device{}
	err := s.db.QueryRow(`SELECT ip, name, is_agent FROM devices WHERE is_agent = TRUE LIMIT 1`).
		Scan(&d.IP, &d.Name, &d.IsAgent)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (s *SQLiteDeviceStore) Close() error {
	return s.db.Close()
}
