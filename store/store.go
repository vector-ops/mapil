// Package store provides an abstraction over the database and manages data persistence
package store

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/vector-ops/mapil/database"
)

const defaultNamespace = "default"

var ErrUnsupportedValue = errors.New("unsupported value")

type Store struct {
	data *database.Database
}

func NewStore(devMode bool) *Store {
	fp := ""
	if devMode {
		curDir, err := os.Getwd()
		if err != nil {
			curDir = "."
		}

		fp = filepath.Join(curDir, ".mapil", "mapil.json")
	}

	return &Store{
		data: database.NewDatabase(fp),
	}
}

func (s *Store) Init() error {
	return s.data.Init()
}

func (s *Store) Close() error {
	return s.data.Close()
}

func (s *Store) AddList(key string, value []string, namespace string) error {
	if namespace == "" {
		namespace = defaultNamespace
	}
	return s.data.AddObject(database.ListType{Key: key, Value: value, Namespace: namespace})
}

func (s *Store) UpdateList(key string, value []string, namespace string) error {
	if namespace == "" {
		namespace = defaultNamespace
	}

	return s.data.UpdateObject(database.ListType{Key: key, Value: value, Namespace: namespace})
}

func (s *Store) AppendList(key string, value []string) error {
	existingValues, err := s.GetValue(key)
	if err != nil {
		return err
	}

	ns, err := s.GetNamespace(key)
	if err != nil {
		return err
	}

	existingValues = append(existingValues, value...)
	return s.UpdateList(key, existingValues, ns)
}

func (s *Store) DeleteValue(key string) error {
	s.data.DeleteObject(key)
	return nil
}

func (s *Store) DeleteAll() {
	keys := s.data.GetAllKeys()
	for _, k := range keys {
		s.data.DeleteObject(k)
	}
}

func (s *Store) GetValue(key string) ([]string, error) {
	keyval, err := s.data.GetObject(key)
	if err != nil {
		return nil, err
	}

	switch keyval.GetType() {
	case database.List:
		return keyval.GetValue().([]string), nil
	default:
		return nil, ErrUnsupportedValue
	}
}

func (s *Store) GetKeys() []string {
	return s.data.GetAllKeys()
}

func (s *Store) GetNamespace(key string) (string, error) {
	ns, err := s.data.GetNamespace(key)
	if err != nil {
		return "", err
	}

	return ns, nil
}

func (s *Store) GetNamespaceObjects(ns string) []database.ListType {
	data := s.data.GetNamespaceObjects(ns)
	var do []database.ListType
	for _, kv := range data {
		switch kv.(type) {
		case database.ListType:
			do = append(do, database.ListType{
				Key:       kv.GetKey(),
				Value:     kv.GetValue().([]string),
				Namespace: kv.GetNamespace(),
			})
		}
	}
	return do
}

func (s *Store) GetAllData() []database.ListType {
	data := s.data.GetAllObjects()
	var do []database.ListType
	for _, kv := range data {
		switch kv.(type) {
		case database.ListType:
			do = append(do, database.ListType{
				Key:       kv.GetKey(),
				Value:     kv.GetValue().([]string),
				Namespace: kv.GetNamespace(),
			})
		}
	}
	return do
}
