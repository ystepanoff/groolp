package watcher

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/mock"
	"github.com/ystepanoff/groolp/internal/core"
)

// MockWatcher implements AbstractWatcher for testing.
type MockWatcher struct {
	mock.Mock
	events chan fsnotify.Event
	errors chan error
}

func NewMockWatcher() *MockWatcher {
	return &MockWatcher{
		events: make(chan fsnotify.Event),
		errors: make(chan error),
	}
}

func (m *MockWatcher) Add(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockWatcher) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockWatcher) Events() <-chan fsnotify.Event {
	return m.events
}

func (m *MockWatcher) Errors() <-chan error {
	return m.errors
}

// MockTaskManager mocks the TaskManager for testing.
type MockTaskManager struct {
	mock.Mock
}

func (m *MockTaskManager) Register(task *core.Task) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockTaskManager) Run(taskName string) error {
	args := m.Called(taskName)
	return args.Error(0)
}

func (m *MockTaskManager) ListTasks() []*core.Task {
	args := m.Called()
	return args.Get(0).([]*core.Task)
}

func TestWatcher_Start(t *testing.T) {
	mockTM := new(MockTaskManager)
	mockTM.On("Run", "deploy").Return(nil)

	mockWatcher := NewMockWatcher()
	mockWatcher.events = make(chan fsnotify.Event)
	mockWatcher.errors = make(chan error)
	mockWatcher.On("Add", ".").Return(nil)
	mockWatcher.On("Close").Return(nil)

	debounceDuration := 500 * time.Millisecond
	w, _ := NewWatcher(
		mockTM,
		[]string{"."},
		"deploy",
		debounceDuration,
		mockWatcher,
	)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.Start()
	}()

	event := fsnotify.Event{
		Name: "testfile.txt",
		Op:   fsnotify.Write,
	}

	mockWatcher.events <- event
	time.Sleep(2 * debounceDuration)

	mockTM.AssertNumberOfCalls(t, "Run", 1)

	mockTM.ExpectedCalls = nil
	mockTM.On("Run", "deploy").Return(nil)

	mockWatcher.events <- event
	mockWatcher.events <- event
	mockWatcher.events <- event

	time.Sleep(2 * debounceDuration)

	mockTM.AssertNumberOfCalls(t, "Run", 2)

	mockTM.ExpectedCalls = nil
	mockTM.On("Run", "deploy").Return(nil)

	mockWatcher.events <- event

	time.Sleep(2 * debounceDuration)

	mockTM.AssertNumberOfCalls(t, "Run", 3)

	mockWatcher.errors <- errors.New("watcher error")

	close(mockWatcher.events)
	close(mockWatcher.errors)

	time.Sleep(100 * time.Millisecond)

	mockTM.AssertExpectations(t)
	mockWatcher.AssertExpectations(t)

	wg.Wait()
}

func TestWatcher_NoEvents(t *testing.T) {
	mockTM := new(MockTaskManager)

	mockWatcher := NewMockWatcher()
	mockWatcher.On("Add", ".").Return(nil)
	mockWatcher.On("Close").Return(nil)

	debounceDuration := 500 * time.Millisecond
	w, _ := NewWatcher(
		mockTM,
		[]string{"."},
		"deploy",
		debounceDuration,
		mockWatcher,
	)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.Start()
	}()

	close(mockWatcher.events)
	close(mockWatcher.errors)

	time.Sleep(100 * time.Millisecond)

	mockTM.AssertNotCalled(t, "Run", mock.Anything)

	mockTM.AssertExpectations(t)
	mockWatcher.AssertExpectations(t)

	wg.Wait()
}

func TestWatcher_MultipleDebounceCycles(t *testing.T) {
	mockTM := new(MockTaskManager)
	mockTM.On("Run", "deploy").Return(nil)

	mockWatcher := NewMockWatcher()
	mockWatcher.On("Add", ".").Return(nil)
	mockWatcher.On("Close").Return(nil)

	debounceDuration := 500 * time.Millisecond
	w, _ := NewWatcher(
		mockTM,
		[]string{"."},
		"deploy",
		debounceDuration,
		mockWatcher,
	)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.Start()
	}()

	event1 := fsnotify.Event{
		Name: "file1.txt",
		Op:   fsnotify.Write,
	}

	mockWatcher.events <- event1
	mockWatcher.events <- event1
	mockWatcher.events <- event1

	time.Sleep(2 * debounceDuration)

	mockTM.AssertNumberOfCalls(t, "Run", 1)

	mockTM.ExpectedCalls = nil
	mockTM.On("Run", "deploy").Return(nil)

	mockWatcher.events <- event1

	time.Sleep(2 * debounceDuration)
	mockTM.AssertNumberOfCalls(t, "Run", 2)

	close(mockWatcher.events)
	close(mockWatcher.errors)

	time.Sleep(100 * time.Millisecond)

	mockTM.AssertExpectations(t)
	mockWatcher.AssertExpectations(t)

	wg.Wait()
}
