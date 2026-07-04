// Package store provides an abstraction over the database and manages data persistence
package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/vector-ops/mapil/database"
)

const defaultNamespace = "default"
const namespacesKey = "namespaces"

var ErrUnsupportedValue = errors.New("unsupported value")
var ErrReservedKeyMutation = errors.New("mutating reserved key is not allowed")
var ErrReservedNamespaceMutation = errors.New("mutating reserved namespace is not allowed")
var ErrDuplicateValue = errors.New("object has duplicate value(s)")

type Store struct {
	data               *database.Database
	allowReservedKeyOp bool
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
	if namespace == namespacesKey && !s.allowReservedKeyOp {
		return ErrReservedNamespaceMutation
	}

	if key == namespacesKey && !s.allowReservedKeyOp {
		return ErrReservedKeyMutation
	}

	if namespace == "" {
		namespace = defaultNamespace
	}

	// kind of a mutex? allowing only one operation on reserved keys at a time
	if !s.allowReservedKeyOp {
		s.allowReservedKeyOp = true
		if err := s.AppendList(namespacesKey, []string{namespace}, false); err != nil {
			if errors.Is(err, database.ErrKeyDoesNotExist) {
				if addErr := s.AddList(namespacesKey, []string{namespace}, ""); addErr != nil {
					return fmt.Errorf("database error")
				}
			} else if !errors.Is(err, ErrDuplicateValue) {
				return err
			}
		}
		s.allowReservedKeyOp = false
	}

	return s.data.AddObject(database.ListType{Key: key, Value: value, Namespace: namespace})
}

func (s *Store) UpdateList(key string, value []string, namespace string) error {
	if key == namespacesKey && !s.allowReservedKeyOp {
		return ErrReservedNamespaceMutation
	}

	if namespace == "" {
		namespace = defaultNamespace
	}

	return s.data.UpdateObject(database.ListType{Key: key, Value: value, Namespace: namespace})
}

func (s *Store) AppendList(key string, values []string, allowDuplicates bool) error {

	if key == namespacesKey && !s.allowReservedKeyOp {
		return ErrReservedKeyMutation
	}

	existingValues, err := s.GetValue(key)
	if err != nil {
		return err
	}

	if !allowDuplicates {
		sortedValues := existingValues
		slices.Sort(sortedValues)
		for _, v := range values {
			if _, found := slices.BinarySearch(sortedValues, v); found {
				return ErrDuplicateValue
			}
		}
	}

	ns, err := s.GetNamespace(key)
	if err != nil {
		return err
	}

	existingValues = append(existingValues, values...)
	return s.UpdateList(key, existingValues, ns)
}

func (s *Store) DeleteObject(key string) error {
	if key == namespacesKey && !s.allowReservedKeyOp {
		return ErrReservedKeyMutation
	}
	s.data.DeleteObject(key)
	return nil
}

func (s *Store) DeleteAll() {
	keys := s.GetKeys()
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
	keys := s.data.GetAllKeys()
	return slices.DeleteFunc(keys, func(k string) bool {
		return k == namespacesKey && !s.allowReservedKeyOp
	})
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
			if kv.GetKey() == namespacesKey && !s.allowReservedKeyOp {
				continue
			}
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
			if kv.GetKey() == namespacesKey && !s.allowReservedKeyOp {
				continue
			}
			do = append(do, database.ListType{
				Key:       kv.GetKey(),
				Value:     kv.GetValue().([]string),
				Namespace: kv.GetNamespace(),
			})
		}
	}
	return do
}
