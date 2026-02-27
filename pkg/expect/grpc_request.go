package expect

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// GRPCRequest invokes a unary gRPC method using JSON over gRPC.
// Body is the JSON-encoded request body; an empty body sends {}.
type GRPCRequest struct {
	// FullMethod is the full gRPC method path, e.g. "/mypackage.MyService/MyMethod".
	FullMethod string
	// Body is the JSON-encoded request body.
	Body []byte
	// Header is outgoing metadata to attach to the call.
	Header map[string]string
}

// Run invokes the gRPC method and returns the raw JSON response bytes.
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

	body := vars.InterpolateBytes(r.Body)
	if len(body) == 0 {
		body = []byte("{}")
	}
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
