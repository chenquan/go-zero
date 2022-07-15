package selector

import (
	"sort"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/md"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/balancer"
)

const DefaultSelector = "defaultSelector"

var _ Selector = (*defaultSelector)(nil)

func init() {
	Register(defaultSelector{})
}

type defaultSelector struct{}

func (d defaultSelector) Select(conns []Conn, info balancer.PickInfo) []Conn {
	m, ok := md.FromContext(info.Ctx)
	if !ok {
		return d.getNoColorConns(conns)
	}

	clientColors := m.Get("color")
	if len(clientColors) == 0 {
		return d.getNoColorConns(conns)
	}

	newConns := make([]Conn, 0, len(conns))
	sort.Strings(clientColors)
	for i := len(clientColors) - 1; i >= 0; i-- {
		clientColor := clientColors[i]
		for _, conn := range conns {
			metadataFromGrpcAttributes := conn.Metadata()
			colors := metadataFromGrpcAttributes.Get("color")
			for _, color := range colors {
				if clientColor == color {
					newConns = append(newConns, conn)
				}
			}
		}

		if len(newConns) != 0 {
			spanCtx := trace.SpanFromContext(info.Ctx)
			spanCtx.SetAttributes(ColorAttributeKey.String(clientColor))
			logx.WithContext(info.Ctx).Infow("flow dyeing", logx.Field("color", clientColor), logx.Field("candidateColors", "["+strings.Join(clientColors, ", ")+"]"))

			break
		}
	}

	return newConns
}

func (d defaultSelector) Name() string {
	return DefaultSelector
}

func (d defaultSelector) getNoColorConns(conns []Conn) []Conn {
	var newConns []Conn
	for _, conn := range conns {
		metadataFromGrpcAttributes := conn.Metadata()
		colors := metadataFromGrpcAttributes.Get("color")
		if len(colors) == 0 {
			newConns = append(newConns, conn)
		}
	}

	return newConns
}
