package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

var ErrInvalidValueType = errors.New("key's type does not match value's asserted type")

type SQLiteDB struct {
	db *sql.DB

	path string
}

func (s *SQLiteDB) withTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return errors.Join(err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

func NewSQLiteDB(p string) Database {
	return &SQLiteDB{
		path: p,
	}
}

// Init implements [Database].
func (s *SQLiteDB) Init(ctx context.Context) error {

	db, err := sql.Open("sqlite", s.path)
	if err != nil {
		return err
	}

	createNamespacesStmt := `
	CREATE TABLE IF NOT EXISTS namespaces (
	    id   INTEGER PRIMARY KEY,
	    name TEXT NOT NULL UNIQUE
	);
`

	createKeysStmt := `
	CREATE TABLE IF NOT EXISTS keys (
		id INTEGER PRIMARY KEY,
		namespace_id INTEGER NOT NULL REFERENCES namespaces(id),
		name TEXT NOT NULL UNIQUE
	);
`

	createValuesStmt := `
	CREATE TABLE IF NOT EXISTS value_store (
		id INTEGER PRIMARY KEY,
		key_id INTEGER REFERENCES keys(id),
		type TEXT NOT NULL,
		value TEXT NOT NULL
	);	
`

	s.db = db

	return s.withTx(ctx, func(tx *sql.Tx) error {
		_, err = tx.ExecContext(ctx, createNamespacesStmt)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, createKeysStmt)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, createValuesStmt)
		if err != nil {
			return err
		}

		return nil
	})

}

// AddObject implements [Database].
func (s *SQLiteDB) AddObject(ctx context.Context, kv KeyValue) error {
	addKey := `
		INSERT INTO keys (name, namespace_id)
		VALUES ($1, $2);
	`

	addValues := `
		INSERT INTO value_store (key_id, type, value)
		VALUES ($1, $2, $3)
	`

	createNamespace := `
		INSERT OR IGNORE INTO namespaces (name)
		VALUES ($1);
	`

	getNamespace := `
		SELECT id FROM namespaces
		WHERE name = $1;
	`

	err := s.withTx(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, createNamespace, kv.GetNamespace())
		if err != nil {
			return err
		}

		namespaceID := 0

		if r, err := res.RowsAffected(); err == nil && r > 0 {

			id, err := res.LastInsertId()
			if err != nil {
				return err
			}

			namespaceID = int(id)
		} else {
			row := tx.QueryRowContext(ctx, getNamespace, kv.GetNamespace())

			if row.Err() != nil {
				return row.Err()
			}

			if err := row.Scan(&namespaceID); err != nil {
				return err
			}

		}

		res, err = tx.ExecContext(ctx, addKey, kv.GetKey(), namespaceID)
		if err != nil {
			return err
		}

		keyID, err := res.LastInsertId()
		if err != nil {
			return err
		}

		stmt, err := tx.PrepareContext(ctx, addValues)
		if err != nil {
			return err
		}
		defer stmt.Close()

		switch kv.GetType() {
		case List:
			values, ok := kv.GetValue().([]string)
			if !ok {
				return errors.New("invalid value")
			}
			for _, v := range values {
				if _, err := stmt.ExecContext(ctx, keyID, kv.GetType(), v); err != nil {
					return err
				}
			}

		default:
			return errors.New("unsupported value type")
		}

		return nil
	})

	return err
}

// Close implements [Database].
func (s *SQLiteDB) Close(ctx context.Context) error {
	return s.db.Close()
}

// DeleteObject implements [Database].
func (s *SQLiteDB) DeleteObject(ctx context.Context, key string) {
	deleteValues := `
		DELETE FROM value_store as v
		WHERE key_id IN (
			SELECT id FROM keys
			WHERE name = $1	
		);
		`
	deleteKey := `
		DELETE FROM keys
		WHERE name = $1;
	`

	err := s.withTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, deleteValues, key); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, deleteKey, key); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.Println(err)
	}

	// TODO: return error
}

// GetAllKeys implements [Database].
func (s *SQLiteDB) GetAllKeys(ctx context.Context) []string {
	res, err := s.db.QueryContext(ctx, "SELECT name FROM keys;")
	if err != nil {
		return nil
	}

	// TODO: return error
	keys := []string{}

	for res.Next() {
		var key string
		if err := res.Scan(&key); err != nil {
			continue
		}

		keys = append(keys, key)
	}

	return keys
}

// GetAllObjects implements [Database].
func (s *SQLiteDB) GetAllObjects(ctx context.Context) []KeyValue {
	// TODO: return error

	query := `
		SELECT k.name, n.name, v.value, v.type FROM keys as k
		JOIN value_store AS v ON k.id = v.key_id 
		JOIN namespaces as n ON k.namespace_id = n.id
		ORDER BY k.id ASC, v.id ASC;
	`

	res, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil
	}

	kvs := make([]KeyValue, 0)
	kvMap := make(map[string]KeyValue)

	for res.Next() {

		var k string
		var ns string
		var v string
		var t string

		if err := res.Scan(&k, &ns, &v, &t); err != nil {
			continue
		}

		switch t {
		case List:
			var vals []string
			if kv, seen := kvMap[k]; seen {
				var ok bool
				vals, ok = kv.GetValue().([]string)
				if !ok {
					continue
				}

			}

			vals = append(vals, v)
			kvMap[k] = ListType{
				Key:       k,
				Value:     vals,
				Namespace: ns,
			}
		default:
			continue
		}

	}

	for _, kv := range kvMap {
		kvs = append(kvs, kv)
	}

	return kvs

}

// GetNamespace implements [Database].
func (s *SQLiteDB) GetNamespace(ctx context.Context, key string) (string, error) {
	query := `
		SELECT n.name FROM namespaces AS n
		JOIN keys AS k ON k.namespace_id = n.id
		WHERE k.name = $1;
	`
	row := s.db.QueryRowContext(ctx, query, key)

	var ns string
	if err := row.Scan(&ns); err != nil {
		return "", err
	}

	return ns, nil
}

// GetNamespaceObjects implements [Database].
func (s *SQLiteDB) GetNamespaceObjects(ctx context.Context, ns string) []KeyValue {
	query := `
		SELECT k.name, v.value, v.type FROM keys as k
		JOIN namespaces as n ON k.namespace_id = n.id
		JOIN value_store AS v ON k.id = v.key_id 
		WHERE n.name = $1
		ORDER BY k.id ASC, v.id ASC;
	`

	res, err := s.db.QueryContext(ctx, query, ns)
	if err != nil {
		return nil
	}

	kvs := make([]KeyValue, 0)
	kvMap := make(map[string]KeyValue)

	for res.Next() {

		var k string
		var v string
		var t string

		if err := res.Scan(&k, &v, &t); err != nil {
			continue
		}

		switch t {
		case List:
			var vals []string
			if kv, seen := kvMap[k]; seen {
				var ok bool
				vals, ok = kv.GetValue().([]string)
				if !ok {
					continue
				}
			}

			vals = append(vals, v)
			kvMap[k] = ListType{
				Key:       k,
				Value:     vals,
				Namespace: ns,
			}
		default:
			continue
		}

	}

	for _, kv := range kvMap {
		kvs = append(kvs, kv)
	}

	return kvs

}

// GetObject implements [Database].
func (s *SQLiteDB) GetObject(ctx context.Context, key string) (KeyValue, error) {
	query := `
		SELECT n.name, v.value, v.type FROM keys as k
		JOIN value_store AS v ON k.id = v.key_id 
		JOIN namespaces as n ON k.namespace_id = n.id
		WHERE k.name = $1
		ORDER BY k.id ASC, v.id ASC;
	`

	res, err := s.db.QueryContext(ctx, query, key)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	var kv KeyValue
	for res.Next() {

		var ns string
		var v string
		var t string

		if err := res.Scan(&ns, &v, &t); err != nil {
			continue
		}

		switch t {
		case List:
			vals := []string{}
			if kv != nil && kv.GetKey() != "" {
				vals = kv.GetValue().([]string)
			}

			vals = append(vals, v)

			kv = ListType{
				Key:       key,
				Value:     vals,
				Namespace: ns,
			}
		default:
			return nil, ErrInvalidValueType
		}

	}

	if kv == nil || kv.GetKey() == "" {
		return nil, ErrKeyDoesNotExist
	}

	return kv, nil
}

// GetValue implements [Database].
func (s *SQLiteDB) GetValue(ctx context.Context, key string) (any, error) {
	query := `
		SELECT v.value, v.type FROM keys as k
		JOIN value_store AS v ON k.id = v.key_id 
		WHERE k.name = $1
		ORDER BY v.id ASC;
	`

	res, err := s.db.QueryContext(ctx, query, key)
	if err != nil {
		return nil, err
	}

	var kv KeyValue
	for res.Next() {
		var v string
		var t string

		if err := res.Scan(&v, &t); err != nil {
			continue
		}

		switch t {
		case List:
			vals := []string{v}
			if kv.GetKey() != "" {
				vals = kv.GetValue().([]string)
			}

			kv = ListType{
				Key:   key,
				Value: vals,
			}
		default:
			return nil, ErrInvalidValueType
		}

	}

	return kv.GetValue(), nil
}

// UpdateObject implements [Database].
func (s *SQLiteDB) UpdateObject(ctx context.Context, kv KeyValue) error {
	createNamespace := `
		INSERT OR IGNORE INTO namespaces (name)
		VALUES ($1);
	`

	getNamespace := `
		SELECT id FROM namespaces
		WHERE name = $1;
	`

	getKey := `
		SELECT id, namespace_id FROM keys
		WHERE name = $1;
	`

	delVals := `
		DELETE FROM value_store 
		WHERE key_id IN (
			SELECT id FROM keys
			WHERE name = $1
		);
	`

	addValues := `
		INSERT INTO value_store (key_id, type, value)
		VALUES ($1, $2, $3)
	`

	updateNamespace := `
		UPDATE keys
		SET namespace_id = $1
		WHERE id = $2;
	`

	return s.withTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, delVals, kv.GetKey()); err != nil {
			return err
		}

		namespaceID := 0
		row := tx.QueryRowContext(ctx, getNamespace, kv.GetNamespace())

		if err := row.Scan(&namespaceID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				res, err := tx.ExecContext(ctx, createNamespace, kv.GetNamespace())
				if err != nil {
					return err
				}

				id, err := res.LastInsertId()
				if err != nil {
					return err
				}

				namespaceID = int(id)
			} else {
				return err
			}
		}

		var keyID int
		var existingNamespaceID int
		row = tx.QueryRowContext(ctx, getKey, kv.GetKey())
		if err := row.Scan(&keyID, &existingNamespaceID); err != nil {
			return fmt.Errorf("get key: %w", err)
		}

		if existingNamespaceID != namespaceID {
			if _, err := tx.ExecContext(ctx, updateNamespace, namespaceID, keyID); err != nil {
				return err
			}
		}

		stmt, err := tx.PrepareContext(ctx, addValues)
		if err != nil {
			return err
		}
		defer stmt.Close()

		switch kv.GetType() {
		case List:
			values, ok := kv.GetValue().([]string)
			if !ok {
				return errors.New("invalid value")
			}
			for _, v := range values {
				if _, err := stmt.ExecContext(ctx, keyID, List, v); err != nil {
					return err
				}
			}

		default:
			return errors.New("unsupported value type")
		}

		return nil
	})
}
