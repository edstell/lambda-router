package router

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// Request is the expected structure of an event which can be routed. It
// contains a Procedure field which names the Handler which should handle the
// request, and the body to be passed to the handler.
type Request struct {
	Procedure string          `json:"procedure"`
	Body      json.RawMessage `json:"body"`
}

// Response is returned from the Router.Handle function and is the form of the
// event which will returned from the lambda invoke function.
// NOTE: Errors returned from a Handler won't be propagated, instead they're
// marshaled to json and streamed as part of the response.
type Response struct {
	Body  json.RawMessage `json:"body,omitempty"`
	Error json.RawMessage `json:"error,omitempty"`
}

// Handler implementations will be called by the Router to handle requests made
// to the associated procedure.
type Handler interface {
	Handle(context.Context, json.RawMessage) (json.RawMessage, error)
}

// The HandlerFunc type is an adapter to allow the use of ordinary functions as
// Lambda handlers. If f is a function with the appropriate signature,
// HandlerFunc(f) is a Handler that calls f.
type HandlerFunc func(context.Context, json.RawMessage) (json.RawMessage, error)

// Handle calls f(ctx, body).
func (f HandlerFunc) Handle(ctx context.Context, body json.RawMessage) (json.RawMessage, error) {
	return f(ctx, body)
}

// Router exposes a 'Handle' method which should be passed to 'lambda.Start'.
// It does the work of unwrapping the request event and passing it to the
// relevant Handler, and wrapping the response, or error returned.
type Router struct {
	Routes       map[string]Handler
	marshalError func(error) (json.RawMessage, error)
}

// Option implementations can mutate the Router to configure how events should
// be handled.
type Option func(*Router)

// New initializes a Router instance with the options passed.
func New(opts ...Option) *Router {
	router := &Router{
		Routes: map[string]Handler{},
		marshalError: func(err error) (json.RawMessage, error) {
			return []byte(err.Error()), nil
		},
	}
	for _, opt := range opts {
		opt(router)
	}
	return router
}

// Route registers the passed Handler to the procedure name.
// NOTE: If multiple Handlers are registered to the same procedure, only the
// last registered will be called.
func (r *Router) Route(procedure string, handler Handler) {
	r.Routes[procedure] = handler
}

// Handle should be passed to 'lambda.Start' to handle inbound requests.
func (r *Router) Handle(ctx context.Context, req Request) (*Response, error) {
	handler, ok := r.Routes[req.Procedure]
	if !ok {
		return nil, errors.New(fmt.Sprintf("unrecognized procedure '%s'", req.Procedure))
	}
	rsp, err := handler.Handle(ctx, req.Body)
	if err != nil {
		body, me := r.marshalError(err)
		if me != nil {
			body = []byte(err.Error())
		}
		return &Response{Error: body}, nil
	}
	return &Response{Body: rsp}, nil
}

// MarshalErrorsWith configures the Router to use the passed function to marshal
// errors returned from a Handler. This allows streaming additional content
// which may be included in your error value. If a custom marshaling function
// isn't provided only the error message  will be propagated to the caller.
// If your marshaling function fails it should return an error; the Router
// will propagate the original err.Error() in this case.
func MarshalErrorsWith(f func(error) (json.RawMessage, error)) Option {
	return func(router *Router) {
		router.marshalError = f
	}
}
