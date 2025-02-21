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

	doneCh chan struct{}
	doneWG sync.WaitGroup
}

func NewDataStore(groolpDir string) (*DataStore, error) {
	ds := &DataStore{
		data:     make(map[string]interface{}),
		dataPath: filepath.Join(groolpDir, "data.json"),

		persistCh: make(chan struct{}, 1),
		doneCh:    make(chan struct{}),
	}

	if _, err := os.Stat(ds.dataPath); err == nil {
		if err := ds.load(); err != nil {
			return nil, fmt.Errorf("failed to load persistent data: %w", err)
		}
	}

	ds.doneWG.Add(1)
	go func() {
		ds.persistenseWorker()
		ds.doneWG.Done()
	}()
	return ds, nil
}

func (ds *DataStore) SetData(key string, val interface{}) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.data[key] = val

	select {
	case ds.persistCh <- struct{}{}:
	default:
	}
}

func (ds *DataStore) GetData(key string) (interface{}, bool) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	val, ok := ds.data[key]
	return val, ok
}

func (ds *DataStore) load() error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	f, err := os.Open(ds.dataPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	return decoder.Decode(&ds.data)
}

func (ds *DataStore) persist() error {
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
		var timerCh <-chan time.Time
		if timer != nil {
			timerCh = timer.C
		} else {
			timerCh = nil
		}

		select {
		case <-ds.persistCh:
			if timer != nil {
				timer.Stop()
			}
			timer = time.NewTimer(debounceDuration)
		case <-timerCh:
			if err := ds.persist(); err != nil {
				fmt.Printf("Error persisting data store: %v\n", err)
			}
			timer = nil
		case <-ds.doneCh:
			if timer != nil {
				timer.Stop()
			}
			if err := ds.persist(); err != nil {
				fmt.Printf("Error persisting data on shutdown: %v\n", err)
			}
			return
		}
	}
}

func (ds *DataStore) Close() {
	close(ds.doneCh)
	ds.doneWG.Wait()
}
