package router

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalRequest(t *testing.T) {
	t.Parallel()
	req := &Request{}
	err := json.Unmarshal([]byte(`{"procedure":"Do","body":{"key":"value"}}`), req)
	require.NoError(t, err)
	assert.Equal(t, &Request{
		Procedure: "Do",
		Body:      []byte(`{"key":"value"}`),
	}, req)
}

func TestHandleUnrecogizedProcedure(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	router := New()
	_, err := router.Handle(ctx, Request{})
	require.Error(t, err)
	assert.Equal(t, "unrecognized procedure ''", err.Error())
}

func TestHandleWithResponseBody(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	body := []byte(`{"body":"response body"}`)
	router := New()
	router.Route("Do", HandlerFunc(func(context.Context, json.RawMessage) (json.RawMessage, error) {
		return body, nil
	}))
	result, err := router.Handle(ctx, Request{Procedure: "Do"})
	require.NoError(t, err)
	assert.Equal(t, &Response{Body: body}, result)
}

func TestHandleWithResponseError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	router := New()
	router.Route("Do", HandlerFunc(func(context.Context, json.RawMessage) (json.RawMessage, error) {
		return nil, assert.AnError
	}))
	result, err := router.Handle(ctx, Request{Procedure: "Do"})
	require.NoError(t, err)
	assert.Equal(t, &Response{
		Error: []byte(assert.AnError.Error()),
	}, result)
}
