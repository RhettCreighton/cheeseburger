package repositories

import (
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger/v4"
)

const (
	// Key prefixes for different entity types
	PostKeyPrefix    = "post:"
	CommentKeyPrefix = "comment:"

	// Sequence keys for auto-incrementing IDs
	PostSeqKey    = "seq:post"
	CommentSeqKey = "seq:comment"
)

// getNextID gets the next available ID for a given sequence key
func getNextID(txn *badger.Txn, seqKey string) (int, error) {
	var id int
	item, err := txn.Get([]byte(seqKey))
	if err == badger.ErrKeyNotFound {
		id = 1
	} else if err != nil {
		return 0, err
	} else {
		err = item.Value(func(val []byte) error {
			id = int(val[0])<<24 | int(val[1])<<16 | int(val[2])<<8 | int(val[3])
			return nil
		})
		if err != nil {
			return 0, err
		}
		id++
	}

	// Store new ID
	idBytes := []byte{byte(id >> 24), byte(id >> 16), byte(id >> 8), byte(id)}
	if err := txn.Set([]byte(seqKey), idBytes); err != nil {
		return 0, err
	}

	return id, nil
}

// marshalEntity marshals an entity to JSON
func marshalEntity(entity interface{}) ([]byte, error) {
	data, err := json.Marshal(entity)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal entity: %v", err)
	}
	return data, nil
}

// unmarshalEntity unmarshals JSON data into an entity
func unmarshalEntity(data []byte, entity interface{}) error {
	if err := json.Unmarshal(data, entity); err != nil {
		return fmt.Errorf("failed to unmarshal entity: %v", err)
	}
	return nil
}
