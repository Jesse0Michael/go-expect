package expect

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// StepBuilder constructs a Step fluently.
type StepBuilder struct {
	step Step
}

// HTTPStep creates a StepBuilder for an HTTP request with any method.
func HTTPStep(method, path string) *StepBuilder {
	return &StepBuilder{
		step: Step{
			Request: &HTTPRequest{
				Method: method,
				Path:   path,
				Header: make(map[string]string),
				Query:  make(map[string]string),
			},
			Expect: &HTTPExpect{},
		},
	}
}

// GET creates a step for a GET request.
func GET(path string) *StepBuilder { return HTTPStep("GET", path) }

// POST creates a step for a POST request.
func POST(path string) *StepBuilder { return HTTPStep("POST", path) }

// PUT creates a step for a PUT request.
func PUT(path string) *StepBuilder { return HTTPStep("PUT", path) }

// PATCH creates a step for a PATCH request.
func PATCH(path string) *StepBuilder { return HTTPStep("PATCH", path) }

// DELETE creates a step for a DELETE request.
func DELETE(path string) *StepBuilder { return HTTPStep("DELETE", path) }

// WithConnection sets which named connection this step uses.
func (b *StepBuilder) WithConnection(name string) *StepBuilder {
	b.step.Connection = name
	return b
}

// WithHeader adds a request header (works for both HTTP and gRPC steps).
func (b *StepBuilder) WithHeader(key, value string) *StepBuilder {
	switch req := b.step.Request.(type) {
	case *HTTPRequest:
		req.Header[key] = value
	case *GRPCRequest:
		req.Header[key] = value
	}
	return b
}

// WithQuery adds a query parameter.
func (b *StepBuilder) WithQuery(key, value string) *StepBuilder {
	b.httpReq().Query[key] = value
	return b
}

// WithBody sets the raw request body.
func (b *StepBuilder) WithBody(body []byte) *StepBuilder {
	b.httpReq().Body = body
	return b
}

// WithJSON marshals v to JSON and sets it as the request body,
// also setting Content-Type: application/json.
func (b *StepBuilder) WithJSON(v any) *StepBuilder {
	data, err := json.Marshal(v)
	if err != nil {
		panic("go-expect: WithJSON marshal error: " + err.Error())
	}
	b.httpReq().Body = data
	b.httpReq().Header["Content-Type"] = "application/json"
	return b
}

// ExpectStatus sets the expected HTTP status code.
func (b *StepBuilder) ExpectStatus(code int) *StepBuilder {
	b.httpExpect().Status = code
	return b
}

// ExpectHeader adds an expected response header assertion.
func (b *StepBuilder) ExpectHeader(key, value string) *StepBuilder {
	if b.httpExpect().Header == nil {
		b.httpExpect().Header = make(map[string]string)
	}
	b.httpExpect().Header[key] = value
	return b
}

// ExpectBody sets the expected response body. v may be:
//   - []byte  — exact bytes
//   - string  — exact string
//   - any other value — marshalled to JSON for partial matching
func (b *StepBuilder) ExpectBody(v any) *StepBuilder {
	switch val := v.(type) {
	case []byte:
		b.httpExpect().Body = ExpectBody(val)
	case string:
		b.httpExpect().Body = ExpectBody(val)
	default:
		data, err := json.Marshal(v)
		if err != nil {
			panic("go-expect: ExpectBody marshal error: " + err.Error())
		}
		b.httpExpect().Body = ExpectBody(data)
	}
	return b
}

// Save extracts a field from the JSON response body into a variable for later steps.
func (b *StepBuilder) Save(field, as string) *StepBuilder {
	b.httpExpect().Save = append(b.httpExpect().Save, SaveEntry{Field: field, As: as})
	return b
}

// Build returns the completed Step.
func (b *StepBuilder) Build() Step {
	return b.step
}

// GRPCCall creates a step that invokes a gRPC method using a proto message as the request.
// The message is marshaled to JSON via protojson before sending.
// fullMethod is the full gRPC method path, e.g. "/mypackage.MyService/MyMethod".
func GRPCCall(connection, fullMethod string, req proto.Message) *StepBuilder {
	body, err := protojson.Marshal(req)
	if err != nil {
		panic("go-expect: GRPCCall marshal error: " + err.Error())
	}
	return GRPCRawCall(connection, fullMethod, body)
}

// GRPCRawCall creates a step that invokes a gRPC method using raw JSON bytes as the request body.
func GRPCRawCall(connection, fullMethod string, body []byte) *StepBuilder {
	return &StepBuilder{
		step: Step{
			Connection: connection,
			Request:    &GRPCRequest{FullMethod: fullMethod, Body: body, Header: make(map[string]string)},
			Expect:     &GRPCExpect{},
		},
	}
}

// ExpectGRPCCode sets the expected gRPC status code name (e.g. "OK", "NOT_FOUND").
func (b *StepBuilder) ExpectGRPCCode(code string) *StepBuilder {
	b.grpcExpect().Code = code
	return b
}

// ExpectGRPCBody sets the expected gRPC response body for partial JSON matching.
func (b *StepBuilder) ExpectGRPCBody(v any) *StepBuilder {
	switch val := v.(type) {
	case []byte:
		b.grpcExpect().Body = ExpectBody(val)
	case string:
		b.grpcExpect().Body = ExpectBody(val)
	default:
		data, err := json.Marshal(v)
		if err != nil {
			panic("go-expect: ExpectGRPCBody marshal error: " + err.Error())
		}
		b.grpcExpect().Body = ExpectBody(data)
	}
	return b
}

// SaveGRPC extracts a field from the JSON gRPC response into a variable for later steps.
func (b *StepBuilder) SaveGRPC(field, as string) *StepBuilder {
	b.grpcExpect().Save = append(b.grpcExpect().Save, SaveEntry{Field: field, As: as})
	return b
}

func (b *StepBuilder) httpReq() *HTTPRequest {
	return b.step.Request.(*HTTPRequest)
}

func (b *StepBuilder) httpExpect() *HTTPExpect {
	return b.step.Expect.(*HTTPExpect)
}

func (b *StepBuilder) grpcExpect() *GRPCExpect {
	return b.step.Expect.(*GRPCExpect)
}
