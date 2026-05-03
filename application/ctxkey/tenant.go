package ctxkey

import (
	"context"

	"github.com/google/uuid"
)

type key string

const tenantIDKey key = "tenant_id"

// DefaultTenantID is the fallback used until real auth is implemented.
var DefaultTenantID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

func WithTenantID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, tenantIDKey, id)
}

func TenantID(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(tenantIDKey).(uuid.UUID); ok {
		return id
	}
	return DefaultTenantID
}
