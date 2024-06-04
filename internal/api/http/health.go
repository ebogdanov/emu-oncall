package http

import (
	"sync"
)

var (
	once    = &sync.Once{}
	checker interface{}
)

func NewHealthChecker(_, _, _ string) interface{} {
	// once.Do(func() {
	//	checker = health.NewChecker(health.CheckerOptions{
	//		Version:   version,
	//		ReleaseID: commit,
	//		ServiceID: hostname,
	//	})
	// })

	return checker
}

func AddHealthCallbacks(names []string, cbs []func()) {
	if checker == nil {
		return
	}
	// checker.AddMultipleCallbacks(names, cbs)
}
