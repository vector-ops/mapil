package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/vector-ops/mapil/database"
	"github.com/vector-ops/mapil/helpers"
)

var ErrUnsupportedValue = errors.New("unsupported value")

type Store struct {
	data *database.Database
	file *helpers.File
}

func NewStore(devMode bool) *Store {
	var file *helpers.File

	if devMode {
		curDir, err := os.Getwd()
		if err != nil {
			curDir = "."
		}

		filePath := filepath.Join(curDir, "mapil.json")

		file = helpers.NewFileObjectWithFile(filePath)
	} else {
		file = helpers.NewFileObject()
	}

	return &Store{
		data: database.NewDatabase(),
		file: file,
	}
}

func (s *Store) Init() error {
	s.file.Init()
	return s.LoadData()
}

func (s *Store) AddList(key string, value []string) error {
	return s.data.AddObject(database.ListType{Key: key, Value: value})
}

func (s *Store) UpdateList(key string, value []string) error {
	return s.data.UpdateObject(database.ListType{Key: key, Value: value})
}

func (s *Store) AppendList(key string, value []string) error {
	existingValues, err := s.GetValue(key)
	if err != nil {
		return err
	}
	existingValues = append(existingValues, value...)
	return s.UpdateList(key, existingValues)
}

func (s *Store) DeleteValue(key string) {
	s.data.DeleteObject(key)
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

type DataObject struct {
	Key   string
	Value []string
}

func (s *Store) GetAllData() []DataObject {
	data := s.data.GetAllObjects()
	var do []DataObject
	for _, kv := range data {
		switch kv.(type) {
		case database.ListType:
			do = append(do, DataObject{
				Key:   kv.GetKey(),
				Value: kv.GetValue().([]string),
			})
		}
	}
	return do
}

func (s *Store) LoadData() error {
	data, err := s.file.LoadFile()
	if err != nil {
		return fmt.Errorf("failed to load data file")
	}
	for _, v := range data {
		s.data.AddObject(v)
	}
	return nil
}

func (s *Store) Persist() error {
	err := s.file.SaveFile(s.data.GetAllObjects())
	if err != nil {
		return fmt.Errorf("failed to save file")
	}
	return nil
}
