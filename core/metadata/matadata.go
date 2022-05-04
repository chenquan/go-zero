package metadata

import (
	"context"
	"net/http"

	"google.golang.org/grpc/metadata"
)

type MD map[string]string
type mdKey struct{}

var metadataKey = mdKey{}

const metadataKeyStr = "$metadata"

func WithHeader(header http.Header, keys ...string) MD {
	m := MD{}
	for _, key := range keys {
		if val, ok := header[key]; ok {
			m[key] = val[0]
		}
	}

	return m
}

func Metadata(key string, value string) MD {
	return MD{key: value}
}

func New(m map[string]string) MD {
	md := MD{}
	for key, val := range m {
		md[key] = val
	}
	return md
}

// Join joins any number of mds into a single MD.
//
// The order of values for each key is determined by the order in which the mds
// containing those values are presented to Join.
func Join(mds ...MD) MD {
	out := MD{}
	for _, md := range mds {
		for k, v := range md {
			out[k] = v
		}
	}
	return out
}

func FormContext(ctx context.Context) MD {
	md := ctx.Value(metadataKey)
	if md == nil {
		return MD{}
	}
	return md.(MD)
}

func NewContext(ctx context.Context, md MD) context.Context {
	return context.WithValue(ctx, metadataKey, md)
}

func NewOutgoingContext(ctx context.Context) context.Context {
	var grpcMd metadata.MD

	requestMetadata, ok := metadata.FromIncomingContext(ctx)
	if ok {
		grpcMd = requestMetadata.Copy()
	}

	md := FormContext(ctx)
	mds := make([]string, 0, 4)
	for k, v := range md {
		mds = append(mds, k, v)
	}
	grpcMd.Append(metadataKeyStr, mds...)

	return metadata.NewOutgoingContext(ctx, grpcMd)
}

func NewIncomingContext(ctx context.Context) context.Context {
	requestMetadata, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}

	mds := requestMetadata.Get(metadataKeyStr)
	for k, v := range mds {
		ctx = context.WithValue(ctx, k, v)
	}
	return ctx
}
