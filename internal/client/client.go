package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

// Protocol represents the gRPC protocol variant
type Protocol int

const (
	ProtocolGRPC Protocol = iota
	ProtocolGRPCWeb
	ProtocolConnect
)

// ParseProtocol parses a protocol string
func ParseProtocol(s string) (Protocol, error) {
	switch strings.ToLower(s) {
	case "grpc":
		return ProtocolGRPC, nil
	case "grpc-web":
		return ProtocolGRPCWeb, nil
	case "connect":
		return ProtocolConnect, nil
	default:
		return 0, fmt.Errorf("invalid protocol %q, must be one of: grpc, grpc-web, connect", s)
	}
}

// Client is a dynamic gRPC client
type Client struct {
	address  string
	prefix   string
	protocol Protocol
	headers  map[string]string
	client   *http.Client
}

// NewClient creates a new dynamic gRPC client
func NewClient(address, prefix string, protocol Protocol, headers map[string]string) *Client {
	return &Client{
		address:  strings.TrimSuffix(address, "/"),
		prefix:   strings.TrimSuffix(prefix, "/"),
		protocol: protocol,
		headers:  headers,
		client:   http.DefaultClient,
	}
}

// Call invokes a gRPC method
func (c *Client) Call(ctx context.Context, method protoreflect.MethodDescriptor, input proto.Message) (proto.Message, error) {
	// Build the full URL path
	// gRPC path format: /{package}.{service}/{method}
	svc := method.Parent().(protoreflect.ServiceDescriptor)
	path := fmt.Sprintf("/%s/%s", svc.FullName(), method.Name())

	// Add prefix if specified
	fullURL := c.address
	if c.prefix != "" {
		fullURL += c.prefix
	}
	fullURL += path

	// Create client options based on protocol
	var opts []connect.ClientOption
	switch c.protocol {
	case ProtocolGRPC:
		opts = append(opts, connect.WithGRPC())
	case ProtocolGRPCWeb:
		opts = append(opts, connect.WithGRPCWeb())
	case ProtocolConnect:
		// Connect is the default, no option needed
	}

	// Create output message factory for dynamic messages
	outputDesc := method.Output()

	// Create a dynamic client for this method with a codec that handles dynamic messages
	client := connect.NewClient[dynamicpb.Message, dynamicpb.Message](
		c.client,
		fullURL,
		append(opts, connect.WithCodec(&dynamicCodec{outputDesc: outputDesc}))...,
	)

	// Create the request
	req := connect.NewRequest(input.(*dynamicpb.Message))

	// Add headers
	for k, v := range c.headers {
		req.Header().Set(k, v)
	}

	// Make the call
	resp, err := client.CallUnary(ctx, req)
	if err != nil {
		var connectErr *connect.Error
		if errors.As(err, &connectErr) {
			return nil, fmt.Errorf("gRPC error [%s]: %s", connectErr.Code(), connectErr.Message())
		}
		return nil, err
	}

	return resp.Msg, nil
}

// dynamicCodec is a custom codec that properly handles dynamic protobuf messages
type dynamicCodec struct {
	outputDesc protoreflect.MessageDescriptor
}

func (c *dynamicCodec) Name() string {
	return "proto"
}

func (c *dynamicCodec) Marshal(msg any) ([]byte, error) {
	protoMsg, ok := msg.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("cannot marshal: expected proto.Message, got %T", msg)
	}
	return proto.Marshal(protoMsg)
}

func (c *dynamicCodec) Unmarshal(data []byte, msg any) error {
	protoMsg, ok := msg.(*dynamicpb.Message)
	if !ok {
		return fmt.Errorf("cannot unmarshal: expected *dynamicpb.Message, got %T", msg)
	}

	// Create a new message with the correct descriptor and unmarshal into it
	newMsg := dynamicpb.NewMessage(c.outputDesc)
	if err := proto.Unmarshal(data, newMsg); err != nil {
		return err
	}

	// The protoMsg passed in might be uninitialized (nil internal state).
	// We need to copy all fields from newMsg to protoMsg using reflection.
	newMsgReflect := newMsg.ProtoReflect()
	protoMsgReflect := protoMsg.ProtoReflect()

	// If the target message descriptor is nil, we have an issue - use direct assignment via pointer
	if protoMsgReflect.Descriptor() == nil {
		// The message is completely uninitialized, copy via pointer dereference
		*protoMsg = *newMsg
		return nil
	}

	// Copy all set fields from newMsg to protoMsg
	newMsgReflect.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		protoMsgReflect.Set(fd, v)
		return true
	})

	return nil
}

// JSONToProto converts JSON data to a protobuf message
func JSONToProto(jsonData string, msgDesc protoreflect.MessageDescriptor) (proto.Message, error) {
	msg := dynamicpb.NewMessage(msgDesc)

	unmarshaler := protojson.UnmarshalOptions{
		DiscardUnknown: false,
	}

	if err := unmarshaler.Unmarshal([]byte(jsonData), msg); err != nil {
		return nil, fmt.Errorf("invalid JSON for message type %s: %w", msgDesc.FullName(), err)
	}

	return msg, nil
}

// ProtoToJSON converts a protobuf message to pretty-printed JSON
func ProtoToJSON(msg proto.Message) (string, error) {
	marshaler := protojson.MarshalOptions{
		Multiline:       true,
		Indent:          "  ",
		EmitUnpopulated: false,
	}

	data, err := marshaler.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
