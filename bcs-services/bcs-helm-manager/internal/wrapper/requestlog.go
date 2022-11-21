package wrapper

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/micro/go-micro/v2/metadata"
	"github.com/micro/go-micro/v2/server"
)

// RequestLogWarpper log request
func RequestLogWarpper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
		md, _ := metadata.FromContext(ctx)
		blog.Infof("receive %s, metadata: %v, req: %v", req.Method(), md, req.Body())
		return fn(ctx, req, rsp)
	}
}
