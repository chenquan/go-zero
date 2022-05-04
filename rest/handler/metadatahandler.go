package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/core/metadata"
)

func MetadataHandler(md metadata.MD) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			ctx := metadata.NewContext(request.Context(), md)

			next.ServeHTTP(writer, request.WithContext(ctx))
		})
	}
}
