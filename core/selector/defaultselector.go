package selector

import (
	"sort"

	"github.com/zeromicro/go-zero/core/md"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/balancer"
)

var _ Selector = (*defaultSelector)(nil)

func init() {
	Register(defaultSelector{})
}

type defaultSelector struct{}

func (d defaultSelector) Select(conns []Conn, info balancer.PickInfo) []Conn {
	m, ok := md.FromContext(info.Ctx)
	if !ok {
		return nil
	}

	clientColors := m.Get("color")
	if len(clientColors) == 0 {
		return nil
	}

	newConns := make([]Conn, 0, len(conns))
	sort.Strings(clientColors)
	for i := len(clientColors) - 1; i >= 0; i-- {
		color := clientColors[i]
		for _, conn := range conns {
			metadataFromGrpcAttributes := conn.Metadata()
			colors := metadataFromGrpcAttributes.Get("color")
			for _, c := range colors {
				if color == c {
					newConns = append(newConns, conn)
				}
			}
		}

		if len(newConns) != 0 {
			spanCtx := trace.SpanFromContext(info.Ctx)
			spanCtx.SetAttributes(ColorAttributeKey.String(color))
			break
		}
	}

	return newConns
}

func (d defaultSelector) Name() string {
	return "defaultSelector"
}
