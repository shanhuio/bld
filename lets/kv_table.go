package lets

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	_ "modernc.org/sqlite" // sqlite db driver
)

type kvTable struct {
	db *sql.DB
}

const kvTableSchema = `CREATE TABLE IF NOT EXISTS kv (
	k TEXT PRIMARY KEY,
	v BLOB NOT NULL
)`

func openKVTable(f string) (*kvTable, error) {
	db, err := sql.Open("sqlite", f)
	if err != nil {
		return nil, fmt.Errorf("open sqlite %q: %w", f, err)
	}
	if _, err := db.Exec(kvTableSchema); err != nil {
		db.Close()
		return nil, fmt.Errorf("create kv table: %w", err)
	}
	return &kvTable{db: db}, nil
}

func (t *kvTable) replace(k string, v any) error {
	bs, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal value: %w", err)
	}
	_, err = t.db.Exec(
		`INSERT INTO kv (k, v) VALUES (?, ?)
		 ON CONFLICT (k) DO UPDATE SET v=excluded.v`, k, bs,
	)
	if err != nil {
		return fmt.Errorf("write kv: %w", err)
	}
	return nil
}

var errKeyNotFound = errors.New("key not found")

func (t *kvTable) get(k string, v any) error {
	var bs []byte
	row := t.db.QueryRow(`SELECT v FROM kv WHERE k=?`, k)
	switch err := row.Scan(&bs); {
	case errors.Is(err, sql.ErrNoRows):
		return errKeyNotFound
	case err != nil:
		return fmt.Errorf("read kv: %w", err)
	}
	if err := json.Unmarshal(bs, v); err != nil {
		return fmt.Errorf("unmarshal kv: %w", err)
	}
	return nil
}

func (t *kvTable) remove(k string) error {
	res, err := t.db.Exec(`DELETE FROM kv WHERE k=?`, k)
	if err != nil {
		return fmt.Errorf("delete kv: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return errKeyNotFound
	}
	return nil
}
