package scripts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DataStore defines a global data store for the scripts to use
type DataStore struct {
	data map[string]interface{}
	mu   sync.Mutex

	dataPath  string
	persistCh chan struct{}
	done      chan struct{}
}

func NewDataStore(groolpDir string) *DataStore {
	ds := &DataStore{
		data:     make(map[string]interface{}),
		dataPath: filepath.Join(groolpDir, "data.json"),
	}
	go ds.persistenseWorker()
	return ds
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

func (ds *DataStore) Persist() error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	f, err := os.Create(ds.dataPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(ds.data)
}

func (ds *DataStore) persistenseWorker() {
	const debounceDuration = 500 * time.Millisecond
	var timer *time.Timer

	for {
		select {
		case <-ds.persistCh:
			if timer != nil {
				timer.Stop()
			}
			timer = time.NewTimer(debounceDuration)
		case <-func() <-chan time.Time {
			if timer != nil {
				return timer.C
			}
			ch := make(chan time.Time)
			return ch
		}():
			if err := ds.Persist(); err != nil {
				fmt.Printf("Error persisting data store: %v\n", err)
			}
			timer = nil
		case <-ds.done:
			return
		}
	}
}
