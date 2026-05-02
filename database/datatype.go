package database

import "encoding/json"

const (
	List = "list"
)

type KeyValue interface {
	GetKey() string
	GetValue() interface{}
	GetType() string
}

type ListType struct {
	Key   string   `json:"key"`
	Value []string `json:"value"`
}

func (lt ListType) GetKey() string {
	return lt.Key
}

func (lt ListType) GetValue() interface{} {
	return lt.Value
}

func (lt ListType) GetType() string {
	return List
}

type KVWrapper struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}
