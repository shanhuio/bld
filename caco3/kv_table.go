package caco3

import (
	"errors"
)

type kvTable struct {
}

func openKVTable(f string) (*kvTable, error) {
	// TODO
	return &kvTable{}, nil
}

func (t *kvTable) replace(k string, v any) error {
	return nil
}

var errKeyNotFound = errors.New("key not found")

func (t *kvTable) get(k string, v any) error {
	return errKeyNotFound
}

func (t *kvTable) remove(k string) error {
	return errKeyNotFound
}
