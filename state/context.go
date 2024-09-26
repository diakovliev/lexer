package state

import (
	"context"

	"github.com/diakovliev/lexer/message"
)

type keyType string

const (
	tokenLevelKey keyType = "token-level"
	stateNameKey  keyType = "state-name"
	historyKey    keyType = "history"
)

// WithHistory sets the history to the context.
func WithHistory[T any](ctx context.Context, history message.History[T]) context.Context {
	return context.WithValue(ctx, historyKey, history)
}

// GetHistory returns the history from the context. If there is no history in the context, it will return nil.
func GetHistory[T any](ctx context.Context) message.History[T] {
	if v := ctx.Value(historyKey); v != nil {
		return v.(message.History[T])
	}
	return nil
}

// WithTokenLevel sets the token level to the context.
func WithTokenLevel(ctx context.Context, level int) context.Context {
	return context.WithValue(ctx, tokenLevelKey, level)
}

// WithNextTokenLevel sets the next token level to the context. If there is no token level in the context,
// it will set it to zero. Otherwise, it will increment the current token level by one.
func WithNextTokenLevel(ctx context.Context) context.Context {
	level, ok := GetTokenLevel(ctx)
	if !ok {
		return WithTokenLevel(ctx, 0)
	}
	return WithTokenLevel(ctx, level+1)
}

// GetTokenLevel returns the token level from the context. If there is no token level in the context,
// it will return zero and false. Otherwise, it will return the current token level and true.
func GetTokenLevel(ctx context.Context) (int, bool) {
	if v := ctx.Value(tokenLevelKey); v != nil {
		return v.(int), true
	}
	return 0, false
}

// withStateName sets the state name to the context.
func withStateName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, stateNameKey, name)
}

// GetStateName returns the state name from the context. If there is no state name in the context,
// it will return an empty string. Otherwise, it will return the current state name.
func GetStateName(ctx context.Context) string {
	if v := ctx.Value(stateNameKey); v != nil {
		return v.(string)
	}
	return ""
}
