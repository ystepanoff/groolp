package scripts

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewDataStore_FileDoesNotExist(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "datastore_test_")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	ds, err := NewDataStore(tempDir)
	require.NoError(t, err)
	defer ds.Close()

	require.Empty(t, ds.data)
}

func TestNewDataStore_FileExists_ValidJSON(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "datastore_test_")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	dataPath := filepath.Join(tempDir, "data.json")
	originalData := map[string]interface{}{
		"foo": "bar",
		"num": 42,
	}
	dataBytes, err := json.Marshal(originalData)
	require.NoError(t, err)
	err = os.WriteFile(dataPath, dataBytes, 0644)
	require.NoError(t, err)

	ds, err := NewDataStore(tempDir)
	require.NoError(t, err)
	defer ds.Close()

	val1, ok1 := ds.GetData("foo")
	require.True(t, ok1)
	require.Equal(t, "bar", val1)

	val2, ok2 := ds.GetData("num")
	require.True(t, ok2)
	require.Equal(t, float64(42), val2)
}

func TestNewDataStore_FileExists_InvalidJSON(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "datastore_test_")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	dataPath := filepath.Join(tempDir, "data.json")
	err = os.WriteFile(dataPath, []byte(`{ "foo": `), 0644)
	require.NoError(t, err)

	_, err = NewDataStore(tempDir)
	require.Error(t, err)
}

func TestSetDataAndGetData(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "datastore_test_")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	ds, err := NewDataStore(tempDir)
	require.NoError(t, err)
	defer ds.Close()

	ds.SetData("key1", "value1")
	ds.SetData("key2", 123)

	val1, ok1 := ds.GetData("key1")
	require.True(t, ok1)
	require.Equal(t, "value1", val1)

	val2, ok2 := ds.GetData("key2")
	require.True(t, ok2)
	require.Equal(t, 123, val2)
}

func TestConcurrency(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "datastore_test_")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	ds, err := NewDataStore(tempDir)
	require.NoError(t, err)
	defer ds.Close()

	var wg sync.WaitGroup
	numGoroutines := 10
	numItems := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < numItems; j++ {
				key := "goroutine" + strconv.Itoa(i) + "_item" + strconv.Itoa(j)
				ds.SetData(key, i*1000+j)
			}
		}(i)
	}
	wg.Wait()

	val, ok := ds.GetData("goroutine0_item0")
	require.True(t, ok)
	require.Equal(t, 0, val)
}

func TestPersistenceAfterDebounce(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "datastore_test_")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	ds, err := NewDataStore(tempDir)
	require.NoError(t, err)
	defer ds.Close()

	ds.SetData("testKey", "testValue")
	time.Sleep(700 * time.Millisecond)

	dataPath := filepath.Join(tempDir, "data.json")
	content, err := os.ReadFile(dataPath)
	require.NoError(t, err)

	var stored map[string]interface{}
	err = json.Unmarshal(content, &stored)
	require.NoError(t, err)
	val, ok := stored["testKey"]
	require.True(t, ok)
	require.Equal(t, "testValue", val)
}

func TestPersistenceOnClose(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "datastore_test_")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	ds, err := NewDataStore(tempDir)
	require.NoError(t, err)

	ds.SetData("closeKey", "closeValue")
	ds.Close()

	dataPath := filepath.Join(tempDir, "data.json")
	content, err := os.ReadFile(dataPath)
	require.NoError(t, err)

	var stored map[string]interface{}
	err = json.Unmarshal(content, &stored)
	require.NoError(t, err)
	val, ok := stored["closeKey"]
	require.True(t, ok)
	require.Equal(t, "closeValue", val)
}
