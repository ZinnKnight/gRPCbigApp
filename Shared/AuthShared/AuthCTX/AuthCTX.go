package AuthCTX

import "context"

type UserAuth struct {
	UserID   string
	UserName string
	UserPlan string
}

// Тут тянем OrderService, можно часть логики снести туда, тогда будет крассзависимость

// в метадату
type contextKey struct{}

func PutUser(ctx context.Context, user *UserAuth) context.Context {
	return context.WithValue(ctx, contextKey{}, user)
}

func GetUser(ctx context.Context) (*UserAuth, bool) {
	user, ok := ctx.Value(contextKey{}).(*UserAuth)
	return user, ok
}
