package internal

import (
	"strings"

	"github.com/zeromicro/go-zero/core/discov"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
)

func parserAddr(sub *discov.Subscriber) ([]resolver.Address, error) {
	var addrs []resolver.Address
	for _, val := range subset(sub.Values(), subsetSize) {
		valSplit := strings.SplitN(val, "@", 2)

		var attr *attributes.Attributes
		addr := val
		if len(valSplit) == 2 {

			attr = attr.WithValue("metadata", valSplit[1])

			addr = valSplit[0]
		}

		addrs = append(addrs, resolver.Address{
			Addr:       addr,
			Attributes: attr,
		})
	}

	return addrs, nil
}
