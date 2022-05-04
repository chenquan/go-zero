package clientinterceptors

import (
	"context"

	"github.com/zeromicro/go-zero/core/metadata"
	"google.golang.org/grpc"
)

func UnaryMetadataInterceptor(ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	ctx = metadata.NewOutgoingContext(ctx)

	return invoker(ctx, method, req, reply, cc, opts...)
}

func StreamMetadataInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn,
	method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	ctx = metadata.NewOutgoingContext(ctx)

	return streamer(ctx, desc, cc, method, opts...)
}
