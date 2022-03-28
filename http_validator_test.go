package runn

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

const validOpenApi3Spec = `
openapi: 3.0.3
info:
  title: test spec
  version: 0.0.1
paths:
  /users:
    post:
      requestBody:
        content:
          application/json:
            schema:        
              properties:
                username: 
                  type: string
                password: 
                  type: string
              required:
                - username
                - password
      responses:
        '200':
          description: OK
        '400':
          description: Error
          content:
            application/json:
              schema:
                properties:
                  error:
                    type: string
                required:
                  - error
  /users/{id}:
    get:
      parameters:
        - description: ID
          explode: false
          in: path
          name: id
          required: true
          schema:
            type: string
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                properties:
                  data:
                    type: object
                    properties:
                      username:
                        type: string
                    required:
                      - username
                      - email
                required:
                  - data
`

const validOpenApi3SpecPrefix = `
openapi: 3.0.3
info:
  title: test spec
  version: 0.0.1
servers:
  - url: https://example.com/api/v1
paths:
  /users:
    post:
      requestBody:
        content:
          application/json:
            schema:        
              properties:
                username: 
                  type: string
                password: 
                  type: string
              required:
                - username
                - password
      responses:
        '200':
          description: OK
        '400':
          description: Error
          content:
            application/json:
              schema:
                properties:
                  error:
                    type: string
                required:
                  - error
  /users/{id}:
    get:
      parameters:
        - description: ID
          explode: false
          in: path
          name: id
          required: true
          schema:
            type: string
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                properties:
                  data:
                    type: object
                    properties:
                      username:
                        type: string
                    required:
                      - username
                      - email
                required:
                  - data
`

func TestOpenApi3Validator(t *testing.T) {
	tests := []struct {
		opts    []RunnerOption
		req     *http.Request
		res     *http.Response
		wantErr bool
	}{
		{
			[]RunnerOption{RunnerOpenApi3FromData([]byte(validOpenApi3Spec))},
			&http.Request{
				Method: http.MethodPost,
				URL:    pathToURL(t, "/users"),
				Header: http.Header{"Content-Type": []string{"application/json"}},
				Body:   io.NopCloser(strings.NewReader(`{"username": "alice", "password": "passw0rd"}`)),
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Body:       nil,
			},
			false,
		},
		{
			[]RunnerOption{RunnerOpenApi3FromData([]byte(validOpenApi3Spec))},
			&http.Request{
				Method: http.MethodPost,
				URL:    pathToURL(t, "/users"),
				Header: http.Header{"Content-Type": []string{"application/json"}},
				Body:   io.NopCloser(strings.NewReader(`{"username": "alice", "password": "passw0rd"}`)),
			},
			&http.Response{
				StatusCode: http.StatusBadRequest,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"error": "bad request"}`)),
			},
			false,
		},
		// {
		// 	[]RunnerOption{RunnerOpenApi3FromData([]byte(validOpenApi3SpecPrefix)), RunnerPrefix("/api/v1")},
		// 	&http.Request{
		// 		Method: http.MethodPost,
		// 		URL:    pathToURL(t, "/users"),
		// 		Header: http.Header{"Content-Type": []string{"application/json"}},
		// 		Body:   io.NopCloser(strings.NewReader(`{"username": "alice", "password": "passw0rd"}`)),
		// 	},
		// 	&http.Response{
		// 		StatusCode: http.StatusOK,
		// 		Body:       io.NopCloser(strings.NewReader("")),
		// 	},
		// 	false,
		// },
	}
	ctx := context.Background()
	for _, tt := range tests {
		c := &RunnerConfig{}
		for _, opt := range tt.opts {
			if err := opt(c); err != nil {
				t.Fatal(err)
			}
		}
		v, err := NewOpenApi3Validator(c)
		if err != nil {
			t.Fatal(err)
		}
		if err := v.ValidateRequest(ctx, tt.req); err != nil {
			if !tt.wantErr {
				t.Errorf("got error: %v", err)
			}
			continue
		}
		if tt.wantErr {
			t.Error("want error")
			continue
		}
		if err := v.ValidateResponse(ctx, tt.req, tt.res); err != nil {
			if !tt.wantErr {
				t.Errorf("got error: %v", err)
			}
			continue
		}
		if tt.wantErr {
			t.Error("want error")
		}
	}
}

func pathToURL(t *testing.T, p string) *url.URL {
	t.Helper()
	u, err := url.Parse(p)
	if err != nil {
		t.Fatal(err)
	}
	return u
}
