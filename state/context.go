package state

import "context"

type keyType string

const (
	stateLevelKey keyType = "state-level"
)

func WithStateLevel(ctx context.Context, level int) context.Context {
	return context.WithValue(ctx, stateLevelKey, level)
}

func WithNextStateLevel(ctx context.Context) context.Context {
	level, ok := GetStateLevel(ctx)
	if !ok {
		return WithStateLevel(ctx, 0)
	}
	return WithStateLevel(ctx, level+1)
}

func GetStateLevel(ctx context.Context) (int, bool) {
	if v := ctx.Value(stateLevelKey); v != nil {
		return v.(int), true
	}
	return 0, false
}
