package grpcsrv

import (
	"context"
)

type Auth struct {
}

func NewAuthInterceptor() *Auth {
	return &Auth{}
}

// todo: implement me
func (a *Auth) AuthAndIdentifyTickerFunc(ctx context.Context) (context.Context, error) {
	return ctx, nil
}
