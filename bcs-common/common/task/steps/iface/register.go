package iface

import (
	"sync"
)

var (
	stepMu sync.RWMutex
	steps  = make(map[string]StepWorkerInterface)
)

// Register makes a StepWorkerInterface available by the provided name.
// If Register is called twice with the same name or if StepWorkerInterface is nil,
// it panics.
func Register(name string, step StepWorkerInterface) {
	stepMu.Lock()
	defer stepMu.Unlock()

	if step == nil {
		panic("task: Register step is nil")
	}

	if _, dup := steps[name]; dup {
		panic("task: Register step twice for work " + name)
	}

	steps[name] = step
}

// GetRegisters get all steps instance
func GetRegisters() map[string]StepWorkerInterface {
	stepMu.Lock()
	defer stepMu.Unlock()

	return steps
}
