package handlers

import "context"

type ctxKey string

const userIdKey ctxKey = "userID"

func WithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, userIdKey, id)
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIdKey).(string)
	return id, ok
}