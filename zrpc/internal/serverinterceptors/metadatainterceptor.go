package serverinterceptors

import (
	"context"

	"github.com/zeromicro/go-zero/core/metadata"
	"google.golang.org/grpc"
)

func UnaryMetadataInterceptor(ctx context.Context, req interface{},
	_ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	ctx = metadata.NewIncomingContext(ctx)

	return handler(ctx, req)
}

func StreamMetadataInterceptor(svr interface{}, ss grpc.ServerStream, _ *grpc.StreamServerInfo,
	handler grpc.StreamHandler) error {
	if s, ok := ss.(*serverStream); ok {
		ctx := metadata.NewIncomingContext(ss.Context())
		s.ctx = ctx
	}

	return handler(svr, ss)
}
