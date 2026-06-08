package database

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/vector-ops/mapil/helpers"
)

const (
	dir      = ".mapil"
	fileName = "mapil.json"
)

type File struct {
	filePath string
}

func NewFileObject() *File {
	return &File{}
}

func NewFileObjectWithFile(filePath string) *File {
	return &File{
		filePath: filePath,
	}
}

func (f *File) Init() error {

	var dirPath string

	if f.filePath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to create data file\n%s", err)
		}

		dirPath = path.Join(home, dir)

	} else {
		f.filePath = path.Join(dirPath, fileName)

		dirPath, _ = strings.CutSuffix(f.filePath, "."+fileName)
	}

	if !helpers.PathExists(dirPath) {
		if err := helpers.CreateDir(dirPath); err != nil {
			return fmt.Errorf("failed to create data file\n%s", err)
		}
	}

	if err := f.createFile(); err != nil {
		return fmt.Errorf("failed to create data file\n%s", err)
	}

	return nil
}

func (f *File) createFile() error {
	file, err := os.OpenFile(f.filePath, os.O_CREATE, os.ModePerm)
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

	if err := helpers.WriteToFile(b, f.filePath); err != nil {
		return err
	}
	return nil
}

func (f *File) LoadFile() ([]KeyValue, error) {
	var data []KeyValue
	file, err := os.Open(f.filePath)
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
		return nil, fmt.Errorf("no data in file")
	}
	return unwrap(data)
}
