package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/vector-ops/mapil/helpers"
)

const (
	MapilDataDir = ".mapil"
	fileName     = "mapil.json"
)

var ErrUnsupportedFileExt = errors.New("unsupported data file")

type File struct {
	path string
}

func NewFileObject() *File {
	return &File{}
}

func NewFileObjectWithFile(filePath string) *File {
	return &File{
		path: filePath,
	}
}

func (f *File) Init() error {

	fmt.Println("file path", f.path)

	if filepath.Ext(f.path) != ".json" {
		return ErrUnsupportedFileExt
	}

	dirPath := filepath.Dir(f.path)

	if !helpers.PathExists(dirPath) {
		if err := helpers.CreateDir(dirPath); err != nil {
			return fmt.Errorf("failed to create data directory\n%s", err)
		}
	}

	if err := f.createFile(); err != nil {
		return fmt.Errorf("failed to create data file\n%s", err)
	}

	return nil
}

func (f *File) createFile() error {
	file, err := os.OpenFile(f.path, os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	return nil
}

func (f *File) SaveFile(data []KeyValue) error {

	b, err := serialize(data)
	if err != nil {
		return err
	}

	if err := helpers.WriteToFile(b, f.path); err != nil {
		return err
	}
	return nil
}

func (f *File) LoadFile() ([]KeyValue, error) {
	var data []KeyValue
	file, err := os.Open(f.path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err = deserialize(file)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func serialize(data []KeyValue) ([]byte, error) {
	var wrappedItems []KVWrapper

	for _, kv := range data {
		switch kv.(type) {
		case ListType:
			lbuf, err := json.Marshal(kv)
			if err != nil {
				return nil, err
			}
			wrappedItems = append(wrappedItems, KVWrapper{
				Type: List,
				Data: lbuf,
			})
		}
	}

	buf, err := json.Marshal(wrappedItems)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func unwrap(data []byte) ([]KeyValue, error) {
	var wrappedItems []KVWrapper
	err := json.Unmarshal(data, &wrappedItems)
	if err != nil {
		return nil, err
	}

	var result []KeyValue

	for _, item := range wrappedItems {
		var obj KeyValue
		switch item.Type {
		case List:
			var lt ListType
			err = json.Unmarshal(item.Data, &lt)
			if err != nil {
				return nil, err
			}
			obj = lt
		default:
			if item.Type == "" {
				return nil, fmt.Errorf("missing type field")
			}
			return nil, fmt.Errorf("unknown type: %s", item.Type)
		}

		result = append(result, obj)
	}

	return result, nil
}

func deserialize(file *os.File) ([]KeyValue, error) {
	var data []byte
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	return unwrap(data)
}
