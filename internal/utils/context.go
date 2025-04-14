package utils

import (
	"context"

	"github.com/kosttiik/pvz-service/internal/models"
)

type contextKey string

const userContextKey contextKey = "user"

func SetUserContext(ctx context.Context, claims *models.Claims) context.Context {
	return context.WithValue(ctx, userContextKey, claims)
}

func GetUserFromContext(ctx context.Context) *models.Claims {
	claims, ok := ctx.Value(userContextKey).(*models.Claims)
	if !ok {
		return nil
	}

	return claims
}
