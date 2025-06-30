package ctx

import (
	"context"
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
var tk = TaurusContextKey("taurus_context")

// NewTaurusContext creates a new TaurusContext
func NewTaurusContext(requestID string) *TaurusContext {
	return &TaurusContext{requestID: requestID, AtTime: time.Now(), data: make(map[string]interface{})}
}

// Set sets a value in the TaurusContext
func (tc *TaurusContext) Set(key string, value interface{}) {
	tc.data[key] = value
}

// Get gets a value from the TaurusContext
func (tc *TaurusContext) Get(key string) interface{} {
	return tc.data[key]
}

// GetRequestID gets the request ID from the TaurusContext
func (tc *TaurusContext) GetRequestID() string {
	return tc.requestID
}

// WithTaurusContext adds a TaurusContext to the context
func WithTaurusContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, tk, NewTaurusContext(requestID))
}

// GetTaurusContext gets the TaurusContext from the context
func GetTaurusContext(ctx context.Context) *TaurusContext {
	return ctx.Value(tk).(*TaurusContext)
}
