// Package store provides an abstraction over the database and manages data persistence
package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/vector-ops/mapil/database"
	"github.com/vector-ops/mapil/helpers"
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

	return fn()
}

type Store struct {
	data database.Database

	reservedOp *reservedOp
}

func NewStore(dev bool, cfg helpers.Config) *Store {
	dbCfg := cfg.PrimaryDB().LoadDefault()

	var db database.Database

	switch dbCfg.Driver {
	case "file":
		fp := ""
		if dev {
			curDir, err := os.Getwd()
			if err != nil {
				curDir = "."
			}

			fp = filepath.Join(curDir, ".mapil", dbCfg.Filename)
		}
		db = database.NewLocalFileDB(fp)
	case "sqlite":
		fp := ""
		if dev {
			curDir, err := os.Getwd()
			if err != nil {
				curDir = "."
			}

			fp = filepath.Join(curDir, ".mapil", dbCfg.Filename)
		}
		db = database.NewSQLiteDB(fp)
	default:
		fp := ""
		if dev {
			curDir, err := os.Getwd()
			if err != nil {
				curDir = "."
			}

			fp = filepath.Join(curDir, ".mapil", dbCfg.Filename)
		}
		db = database.NewLocalFileDB(fp)
	}

	return &Store{
		data: db,
		reservedOp: &reservedOp{
			m:        sync.Mutex{},
			reserved: atomic.Bool{},
		},
	}
}

func (s *Store) Init(ctx context.Context) error {
	return s.data.Init(ctx)
}

func (s *Store) Close(ctx context.Context) error {
	return s.data.Close(ctx)
}

func (s *Store) AddList(ctx context.Context, key string, value []string, namespace string) error {
	if (namespace == namespacesNS || namespace == namespacesKey) && !s.reservedOp.IsLocked() {
		return ErrReservedNamespaceMutation
	}

	if (key == namespacesKey || key == namespacesNS) && !s.reservedOp.IsLocked() {
		return ErrReservedKeyMutation
	}

	if namespace == "" {
		namespace = defaultNamespace
	}

	s.reservedOp.IfUnlocked(
		func() error {
			if err := s.AppendList(ctx, namespacesKey, []string{namespace}, false); err != nil {
				if errors.Is(err, database.ErrKeyDoesNotExist) {
					if addErr := s.AddList(ctx, namespacesKey, []string{namespace}, namespacesNS); addErr != nil {
						return fmt.Errorf("database error")
					}
				} else if !errors.Is(err, ErrDuplicateValue) {
					return err
				}
			}

			return nil
		})

	return s.data.AddObject(ctx, database.ListType{Key: key, Value: value, Namespace: namespace})
}

func (s *Store) UpdateList(ctx context.Context, key string, value []string, namespace string) error {
	if (key == namespacesKey || key == namespacesNS) && !s.reservedOp.IsLocked() {
		return ErrReservedNamespaceMutation
	}

	if namespace == "" {
		namespace = defaultNamespace
	}

	return s.data.UpdateObject(ctx, database.ListType{Key: key, Value: value, Namespace: namespace})
}

func (s *Store) AppendList(ctx context.Context, key string, values []string, allowDuplicates bool) error {

	if (key == namespacesKey || key == namespacesNS) && !s.reservedOp.IsLocked() {
		return ErrReservedKeyMutation
	}

	existingValues, err := s.GetValue(ctx, key)
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

	ns, err := s.GetNamespace(ctx, key)
	if err != nil {
		return err
	}

	existingValues = append(existingValues, values...)
	return s.UpdateList(ctx, key, existingValues, ns)
}

func (s *Store) DeleteObject(ctx context.Context, key string) error {
	if (key == namespacesKey || key == namespacesNS) && !s.reservedOp.IsLocked() {
		return ErrReservedKeyMutation
	}

	ns, err := s.GetNamespace(ctx, key)
	if err != nil {
		return err
	}

	s.reservedOp.ReservedFunc(func() error {
		objects := s.GetNamespaceObjects(ctx, ns)
		if len(objects) == 1 {
			values, err := s.GetValue(ctx, namespacesKey)
			if err != nil {
				return err
			}

			values = slices.DeleteFunc(values, func(v string) bool {
				return v == ns
			})

			return s.UpdateList(ctx, namespacesKey, values, namespacesNS)
		}

		return nil
	})

	s.data.DeleteObject(ctx, key)
	return nil
}

func (s *Store) DeleteAll(ctx context.Context) {
	keys := s.GetKeys(ctx)

	for _, k := range keys {
		s.data.DeleteObject(ctx, k)
	}

	s.reservedOp.ReservedFunc(func() error {
		return s.UpdateList(ctx, namespacesKey, []string{defaultNamespace}, namespacesNS)
	})
}

func (s *Store) GetValue(ctx context.Context, key string) ([]string, error) {
	keyval, err := s.data.GetObject(ctx, key)
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

func (s *Store) GetKeys(ctx context.Context) []string {
	keys := s.data.GetAllKeys(ctx)
	return slices.DeleteFunc(keys, func(k string) bool {
		return k == namespacesKey && !s.reservedOp.IsLocked()
	})
}

func (s *Store) GetNamespace(ctx context.Context, key string) (string, error) {
	ns, err := s.data.GetNamespace(ctx, key)
	if err != nil {
		return "", err
	}

	return ns, nil
}

func (s *Store) GetNamespaceObjects(ctx context.Context, ns string) []database.ListType {
	data := s.data.GetNamespaceObjects(ctx, ns)
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

func (s *Store) GetAllData(ctx context.Context) []database.ListType {
	data := s.data.GetAllObjects(ctx)
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
