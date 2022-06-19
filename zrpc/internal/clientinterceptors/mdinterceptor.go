package clientinterceptors

import (
	"context"

	"github.com/zeromicro/go-zero/core/md"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func UnaryMdInterceptor(defaultMd md.Metadata) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = injectionMd(ctx, defaultMd)

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func StreamMdInterceptor(defaultMd md.Metadata) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		ctx = injectionMd(ctx, defaultMd)

		return streamer(ctx, desc, cc, method, opts...)
	}
}

func injectionMd(ctx context.Context, defaultMd md.Metadata) context.Context {
	m, b := md.FromContext(ctx)
	if !b {
		m = md.Metadata{}
	}

	defaultMd.Range(func(key string, values ...string) bool {
		m.Append(key, values...)
		return true
	})
	ctx = md.NewMetadataContext(ctx, m)

	outgoingMd, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return ctx
	}

	m.Range(func(key string, values ...string) bool {
		outgoingMd.Append(key, values...)
		return true
	})
	ctx = metadata.NewOutgoingContext(ctx, outgoingMd)

	return ctx
}
