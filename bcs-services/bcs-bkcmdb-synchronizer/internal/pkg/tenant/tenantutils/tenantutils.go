package tenantutils

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/constants"
)

// WithTenantIdFromContext set tenantID to context
func WithTenantIdFromContext(ctx context.Context, tenantId string) context.Context {
	return context.WithValue(ctx, constants.BkTenantIdHeaderKey, tenantId)
}

// GetTenantIdFromContext get tenantId from context
func GetTenantIdFromContext(ctx context.Context) string {
	tenantId := ""

	if id, ok := ctx.Value(constants.BkTenantIdHeaderKey).(string); ok {
		tenantId = id
	}

	if tenantId == "" {
		tenantId = string(constants.DefaultTenantId)
	}

	return tenantId
}
