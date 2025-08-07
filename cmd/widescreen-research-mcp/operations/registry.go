package operations

import (
	"context"
	"fmt"
	"sync"
)

// Operation represents a single operation that can be performed
type Operation struct {
	Name        string
	Description string
	Handler     OperationHandler
}

// OperationHandler is the function signature for operation handlers
type OperationHandler func(ctx context.Context, params map[string]interface{}) (interface{}, error)

// OperationRegistry manages all available operations
type OperationRegistry struct {
	operations map[string]*Operation
	mu         sync.RWMutex
}

// NewOperationRegistry creates a new operation registry
func NewOperationRegistry() *OperationRegistry {
	return &OperationRegistry{
		operations: make(map[string]*Operation),
	}
}

// Register registers a new operation
func (r *OperationRegistry) Register(name string, operation *Operation) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.operations[name] = operation
}

// GetOperation returns an operation by name
func (r *OperationRegistry) GetOperation(name string) *Operation {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.operations[name]
}

// ListOperations returns all registered operations
func (r *OperationRegistry) ListOperations() map[string]*Operation {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	ops := make(map[string]*Operation)
	for k, v := range r.operations {
		ops[k] = v
	}
	return ops
}

// Execute executes an operation by name
func (r *OperationRegistry) Execute(ctx context.Context, name string, params map[string]interface{}) (interface{}, error) {
	op := r.GetOperation(name)
	if op == nil {
		return nil, fmt.Errorf("operation not found: %s", name)
	}
	
	return op.Handler(ctx, params)
}