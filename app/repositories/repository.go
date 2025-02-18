package repositories

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/dgraph-io/badger/v4"
)

// KeyPrefix defines the key prefixes for different entity types
const (
	PostKeyPrefix    = "post:"
	CommentKeyPrefix = "comment:"
	// For sequences (auto-incrementing IDs)
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
		return 0, fmt.Errorf("failed to get sequence: %v", err)
	} else {
		err = item.Value(func(val []byte) error {
			id, err = strconv.Atoi(string(val))
			if err != nil {
				return fmt.Errorf("failed to parse sequence: %v", err)
			}
			id++
			return nil
		})
		if err != nil {
			return 0, err
		}
	}

	// Update the sequence
	err = txn.Set([]byte(seqKey), []byte(strconv.Itoa(id)))
	if err != nil {
		return 0, fmt.Errorf("failed to update sequence: %v", err)
	}

	return id, nil
}

// marshalEntity marshals an entity to JSON
func marshalEntity(v interface{}) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal entity: %v", err)
	}
	return data, nil
}

// unmarshalEntity unmarshals JSON data into an entity
func unmarshalEntity(data []byte, v interface{}) error {
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal entity: %v", err)
	}
	return nil
}
