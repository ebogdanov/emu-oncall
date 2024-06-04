package grafana

import (
	"errors"
)

var (
	errEntryNotExists = errors.New("entry does not exist")
)

// OnDuty manages multiple stacks identified by string keys.
type OnDuty struct {
	entry map[string]*DutyState
}

// NewOnDuty creates a new OnDuty
func NewOnDuty() *OnDuty {
	return &OnDuty{
		entry: make(map[string]*DutyState),
	}
}

// CreateEntry creates a new stack for the given key.
func (m *OnDuty) CreateEntry(key string) {
	_, exists := m.entry[key]
	if !exists {
		m.entry[key] = NewDutyState(key)
	}
}

func (m *OnDuty) Get(key string) *DutyState {
	_, exists := m.entry[key]
	if !exists {
		m.entry[key] = NewDutyState(key)
	}

	return m.entry[key]
}

// Push pushes an element onto the stack identified by the given key.
func (m *OnDuty) Push(key string, element []string) error {
	stack, exists := m.entry[key]
	if !exists {
		return errEntryNotExists
	}
	stack.Push(element)
	return nil
}
