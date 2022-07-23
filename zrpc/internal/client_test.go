package internal

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestWithDialOption(t *testing.T) {
	var options ClientOptions
	agent := grpc.WithUserAgent("chrome")
	opt := WithDialOption(agent)
	opt(&options)
	assert.Contains(t, options.DialOptions, agent)
}

func TestWithTimeout(t *testing.T) {
	var options ClientOptions
	opt := WithTimeout(time.Second)
	opt(&options)
	assert.Equal(t, time.Second, options.Timeout)
}

func TestWithNonBlock(t *testing.T) {
	var options ClientOptions
	opt := WithNonBlock()
	opt(&options)
	assert.True(t, options.NonBlock)
}

func TestWithStreamClientInterceptor(t *testing.T) {
	var options ClientOptions
	opt := WithStreamClientInterceptor(func(ctx context.Context, desc *grpc.StreamDesc,
		cc *grpc.ClientConn, method string, streamer grpc.Streamer,
		opts ...grpc.CallOption) (grpc.ClientStream, error) {
		return nil, nil
	})
	opt(&options)
	assert.Equal(t, 1, len(options.DialOptions))
}

func TestWithTransportCredentials(t *testing.T) {
	var options ClientOptions
	opt := WithTransportCredentials(nil)
	opt(&options)
	assert.Equal(t, 1, len(options.DialOptions))
}

func TestWithUnaryClientInterceptor(t *testing.T) {
	var options ClientOptions
	opt := WithUnaryClientInterceptor(func(ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		return nil
	})
	opt(&options)
	assert.Equal(t, 1, len(options.DialOptions))
}

func TestBuildDialOptions(t *testing.T) {
	var c client
	agent := grpc.WithUserAgent("chrome")
	opts := c.buildDialOptions(WithDialOption(agent))
	assert.Contains(t, opts, agent)
}

func TestWithSelector(t *testing.T) {
	var options ClientOptions
	opt := WithSelector("defaultSelector")
	opt(&options)
	assert.Equal(t, "defaultSelector", options.selectorName)
}

func TestWithMetadata(t *testing.T) {
	var options ClientOptions
	opt := WithMetadata(map[string][]string{"a": {"a1"}})
	opt(&options)
	assert.EqualValues(t, map[string][]string{"a": {"a1"}}, options.md)
}
