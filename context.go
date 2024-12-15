package balogan

import "context"

const contextKey = "balogan:context"

func (l *Logger) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey, l)
}

func FromContext(ctx context.Context) (*Logger, bool) {
	logger, ok := ctx.Value(contextKey).(*Logger)
	return logger, ok
}
