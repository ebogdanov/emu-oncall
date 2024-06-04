package grafana

import (
	"errors"
)

var (
	errEmptyStack = errors.New("state is empty")
)

type DutyState struct {
	Name     string
	elements [][]string
}

func NewDutyState(name string) *DutyState {
	return &DutyState{
		Name:     name,
		elements: [][]string{},
	}
}

// Push adds an element
func (s *DutyState) Push(element []string) {
	s.elements = append(s.elements, element)
}

// Pop removes and returns the top element. Returns an error if empty
func (s *DutyState) Pop() ([]string, error) {
	if len(s.elements) == 0 {
		return []string{}, errEmptyStack
	}

	top := s.elements[len(s.elements)-1]
	s.elements = s.elements[:len(s.elements)-1]
	return top, nil
}

// Peek returns the top element without removing it. Returns an error if empty
func (s *DutyState) Peek() ([]string, error) {
	if len(s.elements) == 0 {
		return []string{}, errEmptyStack
	}
	return s.elements[len(s.elements)-1], nil
}

// IsEmpty checks if is empty
func (s *DutyState) IsEmpty() bool {
	return len(s.elements) == 0
}
