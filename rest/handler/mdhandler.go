package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/core/md"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var httpColorAttributeKey = attribute.Key("http.header.color")

func MdHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		colors := request.Header.Values("color")
		if len(colors) != 0 {
			ctx := request.Context()
			metadata, ok := md.FromContext(ctx)
			if !ok {
				metadata = md.Metadata{}
			}

			span := trace.SpanFromContext(ctx)
			span.SetAttributes(httpColorAttributeKey.StringSlice(colors))

			metadata.Append("color", colors...)
			ctx = md.NewMetadataContext(ctx, metadata)
			request = request.WithContext(ctx)
		}

		next.ServeHTTP(writer, request)
	})
}
