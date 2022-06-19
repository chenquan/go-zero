package selector

import (
	"context"

	"github.com/zeromicro/go-zero/core/md"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/resolver"
)

const trafficSelect = "trafficselect"

var (
	selectorMap       = make(map[string]Selector)
	ColorAttributeKey = attribute.Key("selector.color")
)

type (
	Selector interface {
		Select(conns []Conn, info balancer.PickInfo) []Conn
		Name() string
	}
	Conn interface {
		Address() resolver.Address
		SubConn() balancer.SubConn
		Metadata() md.Metadata
	}
)

func Register(selector Selector) {
	selectorMap[selector.Name()] = selector
}

func Get(name string) Selector {
	if b, ok := selectorMap[name]; ok {
		return b
	}
	return nil
}

func SelectFromContext(ctx context.Context) []Selector {
	m, b := md.FromContext(ctx)
	if !b {
		return nil
	}

	selectorNames := m.Get(trafficSelect)
	if len(selectorNames) == 0 {
		return nil
	}

	var selectors []Selector
	for _, selectorName := range selectorNames {
		selector := Get(selectorName)
		if selector != nil {
			selectors = append(selectors, selector)
		}
	}

	return selectors
}

func NewSelectorContext(ctx context.Context, selectorName string) context.Context {
	m, b := md.FromContext(ctx)
	if !b {
		m = md.Metadata{}
	}
	m.Set(trafficSelect, selectorName)

	return md.NewMetadataContext(ctx, m)
}

func AppendSelectorContext(ctx context.Context, selectorName string) context.Context {
	OutgoingMd, b := metadata.FromOutgoingContext(ctx)
	if b {
		OutgoingMd.Append(trafficSelect, selectorName)
	} else {
		OutgoingMd = metadata.New(map[string]string{trafficSelect: selectorName})
	}

	return metadata.NewOutgoingContext(ctx, OutgoingMd)
}
