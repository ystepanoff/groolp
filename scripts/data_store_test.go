package scripts_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ystepanoff/groolp/scripts"
)

func TestDataStore_SetGet(t *testing.T) {
	tmpDir := t.TempDir()

	ds := scripts.NewDataStore(tmpDir)

	ds.SetData("foo", "bar")
	val, ok := ds.GetData("foo")
	assert.True(t, ok, "Expected key 'foo' to exist")
	assert.Equal(t, "bar", val, "Expected value to be 'bar' for key 'foo'")

	_, exists := ds.GetData("unknown")
	assert.False(t, exists, "Expected 'unknown' key not to exist")
}

func TestDataStore_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()

	ds := scripts.NewDataStore(tmpDir)
	ds.SetData("version", "v1.0")
	val, ok := ds.GetData("version")
	assert.True(t, ok)
	assert.Equal(t, "v1.0", val)

	ds.SetData("version", "v2.5")
	val2, ok2 := ds.GetData("version")
	assert.True(t, ok2)
	assert.Equal(t, "v2.5", val2)
}

func TestDataStore_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()

	ds := scripts.NewDataStore(tmpDir)
	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			key := "key" + string(rune('A'+idx))
			ds.SetData(key, idx)
		}(i)
	}

	wg.Wait()

	for i := 0; i < goroutines; i++ {
		key := "key" + string(rune('A'+i))
		val, ok := ds.GetData(key)
		assert.True(t, ok, "Expected key %s to exist", key)
		assert.Equal(t, i, val, "Expected value for key %s to be %d", key, i)
	}

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			ds.SetData("sharedKey", idx)
			_, _ = ds.GetData("sharedKey")
		}(i)
	}
	wg.Wait()
}
