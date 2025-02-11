package scripts

import (
	"sync"
)

type DataStore struct {
	data map[string]interface{}
	mu   sync.Mutex
}

func (ds *DataStore) SetData(key string, val interface{}) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.data[key] = val
}

func (ds *DataStore) GetData(key string) (interface{}, bool) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	val, ok := ds.data[key]
	return val, ok
}
