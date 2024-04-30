package runn

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strconv"
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
              type: object
              properties:
                username:
                  type: string
                password:
                  type: string
              required:
                - username
                - password
      responses:
        '201':
          description: Created
        '400':
          description: Error
          content:
            application/json:
              schema:
                type: object
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
    put:
      parameters:
        - description: ID
          explode: false
          in: path
          name: id
          required: true
          schema:
            type: string
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                username:
                  type: string
                email:
                  type: string
                  nullable: true
              required:
                - username
                - email
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  username:
                    type: string
                  email:
                    type: string
                    nullable: true
                required:
                  - username
                  - email
              examples:
                Example1:
                  value:
                    username: alice
                    email: alice@example.com
                Example2:
                  value:
                    username: alice
                    email: null
  /users2/{id}:
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
                $ref: "#/components/schemas/UserInfo"
  /users3/{id}:
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
                allOf:
                  - $ref: "#/components/schemas/UserInfo"
                  - $ref: "#/components/schemas/UserAdditionalInfo"
  /users4/{id}:
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
                allOf:
                  - $ref: "#/components/schemas/UserInfo"
                  - $ref: "#/components/schemas/UserAdditionalInfo"
                  - required:
                      - username
                      - email
  /users5:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  users:
                    type: array
                    items:
                      $ref: "#/components/schemas/UserWithAdditionalInfo"
  /users5/{id}:
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
                $ref: "#/components/schemas/UserWithAdditionalInfo"
  /users6/{id}:
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
                allOf:
                  - $ref: "#/components/schemas/UserWithAdditionalInfo"
                  - required:
                      - username
                      - email
  /private:
    get:
      parameters: []
      responses:
        '200':
          description: OK
        '404':
          description: Forbidden
          content:
            application/json:
              schema:
                type: object
                properties:
                  error:
                    type: string
                required:
                  - error
      security:
      - Bearer: []
components:
  securitySchemes:
    Bearer:
      type: http
      scheme: bearer
  schemas:
    UserInfo:
      type: object
      properties:
        username:
          type: string
    UserAdditionalInfo:
      type: object
      properties:
        email:
          type: string
          nullable: true
    UserWithAdditionalInfo:
      allOf:
        - $ref: "#/components/schemas/UserInfo"
        - $ref: "#/components/schemas/UserAdditionalInfo"
`

func TestOpenAPI3Validator(t *testing.T) {
	tests := []struct {
		req     *http.Request
		res     *http.Response
		wantErr bool
	}{
		{
			&http.Request{
				Method: http.MethodPost,
				URL:    pathToURL(t, "/users"),
				Header: http.Header{"Content-Type": []string{"application/json"}},
				Body:   io.NopCloser(strings.NewReader(`{"username": "alice", "password": "passw0rd"}`)),
			},
			&http.Response{
				StatusCode: http.StatusCreated,
				Body:       nil,
			},
			false,
		},
		{
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
		{
			&http.Request{
				Method: http.MethodPost,
				URL:    pathToURL(t, "/users"),
				Header: http.Header{"Content-Type": []string{"application/json"}},
				Body:   io.NopCloser(strings.NewReader(`{"username": "alice"}`)),
			},
			&http.Response{
				StatusCode: http.StatusCreated,
				Body:       nil,
			},
			true,
		},
		{
			&http.Request{
				Method: http.MethodPost,
				URL:    pathToURL(t, "/users"),
				Header: http.Header{"Content-Type": []string{"application/json"}},
				Body:   io.NopCloser(strings.NewReader(`{"username": "alice", "password": "passw0rd"}`)),
			},
			&http.Response{
				StatusCode: http.StatusInternalServerError,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"error": "bad request"}`)),
			},
			true,
		},
		{
			&http.Request{
				Method: http.MethodPost,
				URL:    pathToURL(t, "/users"),
				Header: http.Header{"Content-Type": []string{"application/json"}},
				Body:   io.NopCloser(strings.NewReader(`{"username": "alice", "password": "passw0rd"}`)),
			},
			&http.Response{
				StatusCode: http.StatusBadRequest,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"invalid_key": "invalid_value"}`)),
			},
			true,
		},
		{
			&http.Request{
				Method: http.MethodGet,
				URL:    pathToURL(t, "/private"),
				Header: http.Header{"Content-Type": []string{"application/json"}, "Authorization": []string{"Bearer dummy_token"}},
				Body:   nil,
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Body:       nil,
			},
			false,
		},
		{
			// nullable
			&http.Request{
				Method: http.MethodPut,
				URL:    pathToURL(t, "/users/3"),
				Header: http.Header{"Content-Type": []string{"application/json"}},
				Body:   io.NopCloser(strings.NewReader(`{"username": "alice", "email": null}`)),
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"username": "alice", "email": "alice@example.com"}`)),
			},
			false,
		},
		{
			// nullable
			&http.Request{
				Method: http.MethodPut,
				URL:    pathToURL(t, "/users/3"),
				Header: http.Header{"Content-Type": []string{"application/json"}},
				Body:   io.NopCloser(strings.NewReader(`{"username": null, "email": "alice@example.com"}`)),
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"username": "alice", "email": "alice@example.com"}`)),
			},
			true,
		},
		{
			// nullable
			&http.Request{
				Method: http.MethodPut,
				URL:    pathToURL(t, "/users/3"),
				Header: http.Header{"Content-Type": []string{"application/json"}},
				Body:   io.NopCloser(strings.NewReader(`{"username": "alice", "email": "alice@example.com"}`)),
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"username": "alice", "email": null}`)),
			},
			false,
		},
		{
			// nullable
			&http.Request{
				Method: http.MethodPut,
				URL:    pathToURL(t, "/users/3"),
				Header: http.Header{"Content-Type": []string{"application/json"}},
				Body:   io.NopCloser(strings.NewReader(`{"username": "alice", "email": "alice@example.com"}`)),
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"username": null, "email": "alice@example.com"}`)),
			},
			true,
		},
		{
			// nullable
			&http.Request{
				Method: http.MethodPut,
				URL:    pathToURL(t, "/users/3"),
				Header: http.Header{"Content-Type": []string{"application/json"}},
				Body:   io.NopCloser(strings.NewReader(`{"username": "alice", "email": null}`)),
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"username": "alice", "email": null}`)),
			},
			false,
		},
		{
			// nullable
			&http.Request{
				Method: http.MethodGet,
				URL:    pathToURL(t, "/users2/3"),
				Header: http.Header{"Accept": []string{"application/json"}},
				Body:   nil,
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"username": "alice", "email": null}`)),
			},
			false,
		},
		{
			// nullable
			&http.Request{
				Method: http.MethodGet,
				URL:    pathToURL(t, "/users3/3"),
				Header: http.Header{"Accept": []string{"application/json"}},
				Body:   nil,
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"username": "alice", "email": null}`)),
			},
			false,
		},
		{
			// nullable
			&http.Request{
				Method: http.MethodGet,
				URL:    pathToURL(t, "/users4/3"),
				Header: http.Header{"Accept": []string{"application/json"}},
				Body:   nil,
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"username": "alice", "email": null}`)),
			},
			false,
		},
		{
			// nullable
			&http.Request{
				Method: http.MethodGet,
				URL:    pathToURL(t, "/users5/3"),
				Header: http.Header{"Accept": []string{"application/json"}},
				Body:   nil,
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"username": "alice", "email": null}`)),
			},
			false,
		},
		{
			// nullable
			&http.Request{
				Method: http.MethodGet,
				URL:    pathToURL(t, "/users6/3"),
				Header: http.Header{"Accept": []string{"application/json"}},
				Body:   nil,
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"username": "alice", "email": null}`)),
			},
			false,
		},
		{
			// nullable
			&http.Request{
				Method: http.MethodGet,
				URL:    pathToURL(t, "/users5"),
				Header: http.Header{"Accept": []string{"application/json"}},
				Body:   nil,
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"users": [ {"username": "alice", "email": null} ] }`)),
			},
			false,
		},
	}
	ctx := context.Background()
	c := &httpRunnerConfig{}
	opts := []httpRunnerOption{OpenAPI3FromData([]byte(validOpenApi3Spec))}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			t.Fatal(err)
		}
	}
	v, err := newOpenAPI3Validator(c)
	if err != nil {
		t.Fatal(err)
	}
	for i, tt := range tests {
		tt := tt
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			if err := v.ValidateRequest(ctx, tt.req); err != nil {
				if !tt.wantErr {
					t.Errorf("got error: %v", err)
				}
				return
			}
			if err := v.ValidateResponse(ctx, tt.req, tt.res); err != nil {
				if !tt.wantErr {
					t.Errorf("got error: %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Error("want error")
			}
		})
	}
}

const validOpenApi3SpecReproduceIssue882 = `
openapi: 3.0.3
info:
  title: test spec
  version: 0.0.1
paths:
  /messages:
    post:
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Message'
        required: true
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Message'
  /messages/{id}:
    get:
      parameters:
        - description: ID
          in: path
          name: id
          required: true
          schema:
            type: integer
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Message'
components:
  schemas:
    Message:
      type: object
      properties:
        id:
          type: integer
        subject:
          type: string
          nullable: true
        body:
          type: string
          nullable: true
      required:
        - id
        - subject
        - body
`

func TestReusingOpenAPI3ValidatorReproduceIssue882(t *testing.T) {
	tests := []struct {
		req     *http.Request
		res     *http.Response
		wantErr bool
	}{
		{
			&http.Request{
				Method: http.MethodGet,
				URL:    pathToURL(t, "/messages/1"),
				Header: http.Header{"Accept": []string{"application/json"}},
				Body:   nil,
			},
			&http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"id":1,"subject": "foo","body":null}`)), // passing null here is the first key
			},
			false,
		},
		{
			&http.Request{
				Method: http.MethodPost,
				URL:    pathToURL(t, "/messages"),
				Header: http.Header{"Content-Type": []string{"application/json"}, "Accept": []string{"application/json"}},
				Body:   io.NopCloser(strings.NewReader(`{"id":1,"subject":"foo","body":"bar"}`)),
			},
			&http.Response{
				StatusCode: http.StatusCreated,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"id":1,"subject":"foo","body": "bar"}`)),
			},
			false,
		},
	}

	ctx := context.Background()
	c := &httpRunnerConfig{}

	opts := []httpRunnerOption{
		OpenAPI3FromData([]byte(validOpenApi3SpecReproduceIssue882)),
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			t.Fatal(err)
		}
	}

	v, err := newOpenAPI3Validator(c) // reusing the validator is the second key
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		if err := v.ValidateRequest(ctx, tt.req); err != nil {
			if !tt.wantErr {
				t.Errorf("got error: %v", err)
			}
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
