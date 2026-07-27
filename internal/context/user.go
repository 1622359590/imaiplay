package context

import stdcontext "context"

type userKey struct{}

type user struct {
	userID   string
	tenantID string
	email    string
	role     string
}

func WithUser(
	ctx stdcontext.Context,
	userID, tenantID, email, role string,
) stdcontext.Context {
	return stdcontext.WithValue(ctx, userKey{}, user{
		userID: userID, tenantID: tenantID, email: email, role: role,
	})
}

func UserFromContext(
	ctx stdcontext.Context,
) (userID, tenantID, email, role string, ok bool) {
	value, ok := ctx.Value(userKey{}).(user)
	if !ok {
		return "", "", "", "", false
	}
	return value.userID, value.tenantID, value.email, value.role, true
}
