package balogan

import "context"

type contextKeyType string

const contextKey contextKeyType = "balogan:context"

func (l *Logger) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey, l)
}

func FromContext(ctx context.Context) (*Logger, bool) {
	if ctx == nil {
		return nil, false
	}
	logger, ok := ctx.Value(contextKey).(*Logger)
	return logger, ok
}
