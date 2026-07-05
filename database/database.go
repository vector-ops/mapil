package database

import "context"

type Database interface {
	Init(context.Context) error
	Close(context.Context) error
	AddObject(context.Context, KeyValue) error
	UpdateObject(context.Context, KeyValue) error
	GetObject(context.Context, string) (KeyValue, error)
	GetValue(context.Context, string) (any, error)
	GetNamespace(context.Context, string) (string, error)
	GetAllObjects(context.Context) []KeyValue
	GetAllKeys(context.Context) []string
	GetNamespaceObjects(context.Context, string) []KeyValue
	DeleteObject(context.Context, string)
}
