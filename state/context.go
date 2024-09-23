package state

import "context"

type keyType string

const (
	tokenLevelKey keyType = "token-level"
)

func WithTokenLevel(ctx context.Context, level int) context.Context {
	return context.WithValue(ctx, tokenLevelKey, level)
}

func WithNextTokenLevel(ctx context.Context) context.Context {
	level, ok := GetTokenLevel(ctx)
	if !ok {
		return WithTokenLevel(ctx, 0)
	}
	return WithTokenLevel(ctx, level+1)
}

func GetTokenLevel(ctx context.Context) (int, bool) {
	if v := ctx.Value(tokenLevelKey); v != nil {
		return v.(int), true
	}
	return 0, false
}
