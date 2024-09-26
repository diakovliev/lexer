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
	factoryKey    keyType = "factory"
	receiverKey   keyType = "receiver"
)

// WithHistoryProvider sets the history provider to the context.
func WithHistoryProvider[T any](ctx context.Context, history message.History[T]) context.Context {
	return context.WithValue(ctx, historyKey, history)
}

// GetHistoryProvider returns the history provider from the context. If there is no history in the context,
// it will return nil, false.
func GetHistoryProvider[T any](ctx context.Context) (message.History[T], bool) {
	if v := ctx.Value(historyKey); v != nil {
		return v.(message.History[T]), true
	}
	return nil, false
}

// withFactory sets the factory to the context.
func withFactory[T any](ctx context.Context, factory message.Factory[T]) context.Context {
	return context.WithValue(ctx, factoryKey, factory)
}

// GetFactory returns the factory from the context. If there is no factory in the context,
// it will return nil, false.
func GetFactory[T any](ctx context.Context) (message.Factory[T], bool) {
	if v := ctx.Value(factoryKey); v != nil {
		return v.(message.Factory[T]), true
	}
	return nil, false
}

// withReceiver sets the receiver to the context.
func withReceiver[T any](ctx context.Context, receiver message.Receiver[T]) context.Context {
	return context.WithValue(ctx, receiverKey, receiver)
}

// GetReceiver returns the receiver from the context. If there is no receiver in the context,
// it will return nil, false.
func GetReceiver[T any](ctx context.Context) (message.Receiver[T], bool) {
	if v := ctx.Value(receiverKey); v != nil {
		return v.(message.Receiver[T]), true
	}
	return nil, false
}

// withTokenLevel sets the token level to the context.
func withTokenLevel(ctx context.Context, level int) context.Context {
	return context.WithValue(ctx, tokenLevelKey, level)
}

// WithNextTokenLevel sets the next token level to the context. If there is no token level in the context,
// it will set it to zero. Otherwise, it will increment the current token level by one.
func WithNextTokenLevel(ctx context.Context) context.Context {
	level, ok := GetTokenLevel(ctx)
	if !ok {
		return withTokenLevel(ctx, 0)
	}
	return withTokenLevel(ctx, level+1)
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
