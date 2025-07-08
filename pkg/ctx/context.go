package ctx

import (
	"context"
	"errors"
	"time"
)

// TaurusContext is a context for Taurus
type TaurusContext struct {
	requestID string
	data      map[string]interface{}
	AtTime    time.Time
}

// TaurusContextKey is a key for TaurusContext
type TaurusContextKey string

// TaurusContextKey is a key for TaurusContext
var TaurusContextId = TaurusContextKey("taurus_context")

// NewTaurusContext creates a new TaurusContext
func NewTaurusContext(requestID string) *TaurusContext {
	return &TaurusContext{requestID: requestID, AtTime: time.Now(), data: make(map[string]interface{})}
}

// Set sets a value in the TaurusContext
func (tc *TaurusContext) Set(key string, value interface{}) {
	tc.data[key] = value
}

// Get gets a value from the TaurusContext
func (tc *TaurusContext) Get(key string) (interface{}, error) {
	if v, exists := tc.data[key]; exists {
		return v, nil
	}
	return nil, errors.New("key not found")
}

// GetRequestID gets the request ID from the TaurusContext
func (tc *TaurusContext) GetRequestID() string {
	return tc.requestID
}

// WithTaurusContext adds a TaurusContext to the context
func WithTaurusContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, TaurusContextId, NewTaurusContext(requestID))
}

// GetTaurusContext gets the TaurusContext from the context
func GetTaurusContext(ctx context.Context) (*TaurusContext, error) {
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	if v, ok := ctx.Value(TaurusContextId).(*TaurusContext); ok {
		return v, nil
	}
	return nil, errors.New("taurus context not found")
}
