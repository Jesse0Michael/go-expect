package expect

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// GRPCRequest invokes a unary gRPC method. Set Message for typed proto invocation
// (from Go code with compiled stubs), or Body for raw JSON invocation (from YAML or
// dynamic tests without compiled stubs). Message takes precedence when both are set.
type GRPCRequest struct {
	// FullMethod is the full gRPC method path, e.g. "/mypackage.MyService/MyMethod".
	FullMethod string
	// Message is the request proto message (typed invocation).
	Message proto.Message
	// Body is the JSON-encoded request body (raw invocation).
	Body []byte
	// Header is outgoing metadata to attach to the call.
	Header map[string]string
}

// Run invokes the gRPC method and returns the raw response bytes (JSON-encoded).
func (r *GRPCRequest) Run(conn *GRPCConnection, vars VarStore) ([]byte, error) {
	cc, err := conn.ClientConn()
	if err != nil {
		return nil, err
	}

	fullMethod := vars.Interpolate(r.FullMethod)

	ctx := context.Background()
	if len(r.Header) > 0 {
		md := metadata.New(nil)
		for k, v := range r.Header {
			md.Append(k, vars.Interpolate(v))
		}
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	if r.Message != nil {
		// Typed invocation: use protojson for marshal/unmarshal.
		respMsg := r.Message.ProtoReflect().New().Interface()
		if err := cc.Invoke(ctx, fullMethod, r.Message, respMsg); err != nil {
			return nil, fmt.Errorf("grpc invoke %q: %w", fullMethod, grpcStatusError(err))
		}
		respBytes, err := protojson.Marshal(respMsg)
		if err != nil {
			return nil, fmt.Errorf("marshal grpc response: %w", err)
		}
		return respBytes, nil
	}

	// Raw JSON invocation: pass bytes through as-is via codec override.
	body := vars.InterpolateBytes(r.Body)
	var respRaw json.RawMessage
	err = cc.Invoke(ctx, fullMethod,
		(*jsonMessage)(&body),
		(*jsonMessage)((*[]byte)(&respRaw)),
		grpc.ForceCodec(jsonCodec{}),
	)
	if err != nil {
		return nil, fmt.Errorf("grpc invoke %q: %w", fullMethod, grpcStatusError(err))
	}
	return respRaw, nil
}

// GRPCExpect validates a gRPC response.
type GRPCExpect struct {
	// Code is the expected gRPC status code name (e.g. "OK", "NOT_FOUND").
	// If empty, any code is accepted.
	Code string
	// Body is the expected response body for partial JSON matching.
	Body ExpectBody
	// Save extracts fields from the JSON response into variables.
	Save []SaveEntry
}

// Validate checks the gRPC response bytes against expectations.
func (e *GRPCExpect) Validate(respBytes []byte, grpcErr error, vars VarStore) error {
	if e.Code != "" {
		st, _ := status.FromError(grpcErr)
		if st.Code().String() != e.Code {
			return fmt.Errorf("unexpected grpc code: %s", st.Code().String())
		}
	} else if grpcErr != nil {
		return fmt.Errorf("unexpected grpc error: %w", grpcErr)
	}

	if e.Body != nil && respBytes != nil {
		if err := e.Body.Validate(respBytes); err != nil {
			return err
		}
	}

	if len(e.Save) > 0 && vars != nil && respBytes != nil {
		saveFromJSON(respBytes, e.Save, vars)
	}

	return nil
}

func grpcStatusError(err error) error {
	if st, ok := status.FromError(err); ok {
		return fmt.Errorf("%s: %s", st.Code(), st.Message())
	}
	return err
}

// jsonMessage is a grpc codec adapter for raw JSON bytes.
type jsonMessage []byte

func (j *jsonMessage) ProtoMessage()            {}
func (j *jsonMessage) Reset()                   { *j = nil }
func (j *jsonMessage) String() string           { return string(*j) }
func (j *jsonMessage) Marshal() ([]byte, error) { return *j, nil }
func (j *jsonMessage) Unmarshal(b []byte) error { *j = b; return nil }

// jsonCodec is a gRPC codec that passes JSON bytes through as-is.
type jsonCodec struct{}

func (jsonCodec) Name() string { return "json" }
func (jsonCodec) Marshal(v any) ([]byte, error) {
	if m, ok := v.(*jsonMessage); ok {
		return []byte(*m), nil
	}
	return json.Marshal(v)
}
func (jsonCodec) Unmarshal(data []byte, v any) error {
	if m, ok := v.(*jsonMessage); ok {
		*m = data
		return nil
	}
	return json.Unmarshal(data, v)
}
