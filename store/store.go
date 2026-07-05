// Package store provides an abstraction over the database and manages data persistence
package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/vector-ops/mapil/database"
)

const defaultNamespace = "default"
const namespacesKey = "namespaces"
const namespacesNS = "namespace"

var ErrUnsupportedValue = errors.New("unsupported value")
var ErrReservedKeyMutation = errors.New("mutating reserved key is not allowed")
var ErrReservedNamespaceMutation = errors.New("mutating reserved namespace is not allowed")
var ErrDuplicateValue = errors.New("object has duplicate value(s)")

type reservedOp struct {
	m        sync.Mutex
	reserved atomic.Bool
}

func (r *reservedOp) Lock() {
	r.m.Lock()
	r.reserved.Store(true)
}

func (r *reservedOp) Unlock() {
	r.reserved.Store(false)
	r.m.Unlock()
}

func (r *reservedOp) IsLocked() bool {
	return r.reserved.Load()
}

func (r *reservedOp) ReservedFunc(fn func() error) error {
	r.Lock()
	defer r.Unlock()
	return fn()
}

func (r *reservedOp) IfUnlocked(fn func() error) error {
	if !r.m.TryLock() {
		return nil
	}

	defer func() {
		r.reserved.Store(false)
		r.m.Unlock()
	}()

	return r.ReservedFunc(fn)
}

type Store struct {
	data *database.Database

	reservedOp *reservedOp
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
		reservedOp: &reservedOp{
			m:        sync.Mutex{},
			reserved: atomic.Bool{},
		},
	}
}

func (s *Store) Init() error {
	return s.data.Init()
}

func (s *Store) Close() error {
	return s.data.Close()
}

func (s *Store) AddList(key string, value []string, namespace string) error {
	if (namespace == namespacesNS || namespace == namespacesKey) && !s.reservedOp.IsLocked() {
		return ErrReservedNamespaceMutation
	}

	if (key == namespacesKey || key == namespacesNS) && !s.reservedOp.IsLocked() {
		return ErrReservedKeyMutation
	}

	if namespace == "" {
		namespace = defaultNamespace
	}

	// if !s.reservedOp.IsLocked() {
	// s.reservedOp.Lock()

	s.reservedOp.IfUnlocked(
		func() error {
			if err := s.AppendList(namespacesKey, []string{namespace}, false); err != nil {
				if errors.Is(err, database.ErrKeyDoesNotExist) {
					if addErr := s.AddList(namespacesKey, []string{namespace}, namespacesNS); addErr != nil {
						return fmt.Errorf("database error")
					}
				} else if !errors.Is(err, ErrDuplicateValue) {
					return err
				}
			}

			return nil
		})
	// s.reservedOp.Unlock()
	// }

	return s.data.AddObject(database.ListType{Key: key, Value: value, Namespace: namespace})
}

func (s *Store) UpdateList(key string, value []string, namespace string) error {
	if (key == namespacesKey || key == namespacesNS) && !s.reservedOp.IsLocked() {
		return ErrReservedNamespaceMutation
	}

	if namespace == "" {
		namespace = defaultNamespace
	}

	return s.data.UpdateObject(database.ListType{Key: key, Value: value, Namespace: namespace})
}

func (s *Store) AppendList(key string, values []string, allowDuplicates bool) error {

	if (key == namespacesKey || key == namespacesNS) && !s.reservedOp.IsLocked() {
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
	if (key == namespacesKey || key == namespacesNS) && !s.reservedOp.IsLocked() {
		return ErrReservedKeyMutation
	}

	ns, err := s.GetNamespace(key)
	if err != nil {
		return err
	}

	s.reservedOp.ReservedFunc(func() error {
		objects := s.GetNamespaceObjects(ns)
		if len(objects) == 1 {
			values, err := s.GetValue(namespacesKey)
			if err != nil {
				return err
			}

			values = slices.DeleteFunc(values, func(v string) bool {
				return v == ns
			})

			return s.UpdateList(namespacesKey, values, namespacesNS)
		}

		return nil
	})

	s.data.DeleteObject(key)
	return nil
}

func (s *Store) DeleteAll() {
	keys := s.GetKeys()

	for _, k := range keys {
		s.data.DeleteObject(k)
	}

	s.reservedOp.ReservedFunc(func() error {
		return s.UpdateList(namespacesKey, []string{defaultNamespace}, namespacesNS)
	})
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
		return k == namespacesKey && !s.reservedOp.IsLocked()
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
			if (kv.GetKey() == namespacesKey || kv.GetKey() == namespacesNS) && !s.reservedOp.IsLocked() {
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
			if (kv.GetKey() == namespacesKey || kv.GetKey() == namespacesNS) && !s.reservedOp.IsLocked() {
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
