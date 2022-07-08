package md

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/attributes"
)

var _ Carrier = (*Metadata)(nil)

type (
	Metadata    map[string][]string
	metadataKey struct{}
)

func (m Metadata) Append(key string, values ...string) {
	if len(values) == 0 {
		return
	}

	key = strings.ToLower(key)
	m[key] = append(m[key], values...)
}

func (m Metadata) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

func (m Metadata) Range(f func(key string, values ...string) bool) {
	for key, value := range m {
		if !f(key, value...) {
			break
		}
	}
}

func (m Metadata) Set(key string, values ...string) {
	key = strings.ToLower(key)
	m[key] = values
}

func (m Metadata) Get(key string) []string {
	key = strings.ToLower(key)
	return m[key]
}

func FromContext(ctx context.Context) (Metadata, bool) {
	value := ctx.Value(metadataKey{})
	if value == nil {
		return nil, false
	}

	return value.(Metadata), true
}

func MetadataFromGrpcAttributes(attributes *attributes.Attributes) (Metadata, bool) {
	value := attributes.Value("metadata")
	if value != nil {
		return nil, false
	}

	m := Metadata{}
	err := json.Unmarshal([]byte(value.(string)), &m)
	if err != nil {
		logx.Errorf("parsing metadata err:%s, data:%s", err, value.(string))
		return nil, false
	}

	return m, true
}

func NewMetadataContext(ctx context.Context, carrier Carrier) context.Context {
	md := Metadata{}
	for _, k := range carrier.Keys() {
		md[k] = carrier.Get(k)
	}

	return context.WithValue(ctx, metadataKey{}, md)
}
