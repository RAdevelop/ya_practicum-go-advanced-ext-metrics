package audit

import "context"

type ipKey struct{}

// WithIP — кладёт IP клиента в контекст.
func WithIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, ipKey{}, ip)
}

// IPFromContext — достаёт IP клиента из контекста.
func IPFromContext(ctx context.Context) string {
	ip, ok := ctx.Value(ipKey{}).(string)
	if !ok {
		return ""
	}
	return ip
}
