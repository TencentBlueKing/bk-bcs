package utils

import (
	"context"
	"fmt"
)

type GrpcTokenAuth struct {
	Token string
}

func (t GrpcTokenAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": fmt.Sprintf("Bearer %s", t.Token),
	}, nil
}

func (GrpcTokenAuth) RequireTransportSecurity() bool {
	return true
}
