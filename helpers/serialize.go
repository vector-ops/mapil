package helpers

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/vector-ops/mapil/database"
)

func Serialize(data []database.KeyValue) ([]byte, error) {
	var wrappedItems []database.KVWrapper

	for _, kv := range data {
		switch kv.(type) {
		case database.ListType:
			lbuf, err := json.Marshal(kv)
			if err != nil {
				return nil, err
			}
			wrappedItems = append(wrappedItems, database.KVWrapper{
				Type: database.List,
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

func Deserialize(data []byte) ([]database.KeyValue, error) {
	var wrappedItems []database.KVWrapper
	err := json.Unmarshal(data, &wrappedItems)
	if err != nil {
		return nil, err
	}

	var result []database.KeyValue

	for _, item := range wrappedItems {
		var obj database.KeyValue
		switch item.Type {
		case database.List:
			var lt database.ListType
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

func DeserializeFile(file *os.File) ([]database.KeyValue, error) {
	var data []byte
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("no data in file")
	}
	return Deserialize(data)
}
