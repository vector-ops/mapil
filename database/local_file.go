// Package database provides methods to interact with the data
package database

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrKeyDoesNotExist = errors.New("key does not exist")
	ErrConflictingKeys = errors.New("key already exists")
)

type LocalFile struct {
	List map[string]KeyValue `json:"list,omitempty"`
	file *File
}

func NewLocalFileDB(fp string) Database {

	file := NewFileObjectWithFile(fp)

	return &LocalFile{
		List: make(map[string]KeyValue),
		file: file,
	}

}

func (d *LocalFile) Init(ctx context.Context) error {
	if err := d.file.Init(); err != nil {
		return err
	}

	return d.loadData(ctx)
}

func (d *LocalFile) Close(ctx context.Context) error {
	return d.persist(ctx)
}

func (d *LocalFile) AddObject(_ context.Context, kv KeyValue) error {
	if _, ok := d.List[kv.GetKey()]; ok {
		return ErrConflictingKeys
	}

	d.List[kv.GetKey()] = kv
	return nil
}

func (d *LocalFile) UpdateObject(_ context.Context, kv KeyValue) error {
	if _, ok := d.List[kv.GetKey()]; ok {
		d.List[kv.GetKey()] = kv
		return nil
	}
	return ErrKeyDoesNotExist
}

func (d *LocalFile) GetObject(_ context.Context, key string) (KeyValue, error) {
	if kv, ok := d.List[key]; ok {
		return kv, nil
	}
	return nil, ErrKeyDoesNotExist
}

func (d *LocalFile) GetValue(_ context.Context, key string) (any, error) {
	if kv, ok := d.List[key]; ok {
		return kv.GetValue(), nil
	}
	return nil, ErrKeyDoesNotExist
}

func (d *LocalFile) GetNamespace(_ context.Context, key string) (string, error) {

	if kv, ok := d.List[key]; ok {
		return kv.GetNamespace(), nil
	}

	return "", ErrKeyDoesNotExist
}

func (d *LocalFile) GetAllObjects(_ context.Context) []KeyValue {
	var objs []KeyValue
	for _, kv := range d.List {
		objs = append(objs, kv)
	}
	return objs
}

func (d *LocalFile) GetAllKeys(_ context.Context) []string {
	keys := make([]string, 0, len(d.List))
	for k := range d.List {
		keys = append(keys, k)
	}
	return keys
}

func (d *LocalFile) GetNamespaceObjects(_ context.Context, ns string) []KeyValue {
	var objs []KeyValue
	for _, kv := range d.List {
		if kv.GetNamespace() == ns {
			objs = append(objs, kv)
		}
	}

	return objs
}

func (d *LocalFile) DeleteObject(_ context.Context, key string) {
	delete(d.List, key)
}

func (d *LocalFile) loadData(ctx context.Context) error {

	data, err := d.file.LoadFile()
	if err != nil {
		return fmt.Errorf("failed to load data file: %v", err)
	}

	for _, v := range data {
		d.AddObject(ctx, v)
	}

	return nil
}

func (d *LocalFile) persist(ctx context.Context) error {
	err := d.file.SaveFile(d.GetAllObjects(ctx))
	if err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}
	return nil
}
