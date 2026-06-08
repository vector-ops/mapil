package database

import (
	"errors"
)

var (
	ErrKeyDoesNotExist = errors.New("key does not exist")
	ErrConflictingKeys = errors.New("key already exists")
)

type Database struct {
	List map[string]KeyValue `json:"list,omitempty"`
}

func NewDatabase(fp string) *Database {

	file := NewFileObjectWithFile(fp)

	return &Database{
		List: make(map[string]KeyValue),
		file: file,
	}

}

func (d *Database) Init() error {
	if err := d.file.Init(); err != nil {
		return err
	}

	return d.LoadData()
}

func (d *Database) Close() error {
	return d.persist()
}

func (d *Database) AddObject(kv KeyValue) error {
	if _, ok := d.List[kv.GetKey()]; ok {
		return ErrConflictingKeys
	}
	d.List[kv.GetKey()] = kv
	return nil
}

func (d *Database) UpdateObject(kv KeyValue) error {
	if _, ok := d.List[kv.GetKey()]; ok {
		d.List[kv.GetKey()] = kv
		return nil
	}
	return ErrKeyDoesNotExist
}

func (d *Database) GetObject(key string) (KeyValue, error) {
	if kv, ok := d.List[key]; ok {
		return kv, nil
	}
	return nil, ErrKeyDoesNotExist
}

func (d *Database) GetValue(key string) (interface{}, error) {
	if kv, ok := d.List[key]; ok {
		return kv.GetValue(), nil
	}
	return nil, ErrKeyDoesNotExist
}

func (d *Database) GetAllObjects() []KeyValue {
	var objs []KeyValue
	for _, kv := range d.List {
		objs = append(objs, kv)
	}
	return objs
}

func (d *Database) GetAllKeys() []string {
	keys := make([]string, 0, len(d.List))
	for k := range d.List {
		keys = append(keys, k)
	}
	return keys
}

func (d *Database) DeleteObject(key string) {
	delete(d.List, key)
}

func (d *Database) LoadData() error {

	data, err := d.file.LoadFile()
	if err != nil {
		return fmt.Errorf("failed to load data file: %v", err)
	}

	for _, v := range data {
		d.AddObject(v)
	}

	return nil
}

func (d *Database) persist() error {
	err := d.file.SaveFile(d.GetAllObjects())
	if err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}
	return nil
}
