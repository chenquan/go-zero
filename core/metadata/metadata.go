package metadata

import (
	"context"
	"net/http"

	"google.golang.org/grpc/metadata"
)

type colorKey struct{}

func FromHeader(header http.Header) metadata.MD {
	md := make(metadata.MD, len(header))
	for k, v := range header {
		values := make([]string, len(v))
		copy(values, v)
		md[k] = values
	}

	return md
}

func NewOutgoingContextFromRequest(r *http.Request) context.Context {
	return metadata.NewOutgoingContext(r.Context(), FromHeader(r.Header))
}

func ColorsFromContext(ctx context.Context) []string {
	value := ctx.Value(colorKey{})
	if value != nil {
		return value.([]string)
	}
	md, b := metadata.FromIncomingContext(ctx)
	if b {
		colors := md.Get("color")

		return colors
	}

	return nil
}

//func ColorsFromContext(ctx context.Context) []string {
//	value := ctx.Value(colorKey)
//	if value == nil {
//		return nil
//	}
//
//	return value.([]string)
//}

func ColorsFromMetadataContext(ctx context.Context) context.Context {
	value := ctx.Value(colorKey{})
	if value != nil {
		return ctx
	}

	md, b := metadata.FromIncomingContext(ctx)
	if b {
		colors := md.Get("color")
		return context.WithValue(ctx, colorKey{}, colors)
	}

	return ctx
}
