package auth

import (
	"context"

	"github.com/jholhewres/whatsgo-api/internal/store"
)

type contextKey string

const (
	instanceKey  contextKey = "instance"
	globalAuthKey contextKey = "global_auth"
)

func SetInstance(ctx context.Context, inst *store.Instance) context.Context {
	return context.WithValue(ctx, instanceKey, inst)
}

func GetInstance(ctx context.Context) *store.Instance {
	inst, _ := ctx.Value(instanceKey).(*store.Instance)
	return inst
}

func SetGlobalAuth(ctx context.Context) context.Context {
	return context.WithValue(ctx, globalAuthKey, true)
}

func IsGlobalAuth(ctx context.Context) bool {
	v, _ := ctx.Value(globalAuthKey).(bool)
	return v
}
