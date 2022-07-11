package runn

import (
	"net/http"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/metadata"
)

func TestParseHTTPRequest(t *testing.T) {
	tests := []struct {
		in      string
		want    *httpRequest
		wantErr bool
	}{
		{
			`
/login:
  post:
    body:
      application/json:
        key: value
`,
			&httpRequest{
				path:      "/login",
				method:    http.MethodPost,
				mediaType: MediaTypeApplicationJSON,
				headers:   map[string]string{},
				body: map[string]interface{}{
					"key": "value",
				},
			},
			false,
		},
		{
			`
/users/k1LoW:
  get: 
    body: null
`,
			&httpRequest{
				path:      "/users/k1LoW",
				method:    http.MethodGet,
				mediaType: "",
				headers:   map[string]string{},
				body:      nil,
			},
			false,
		},
		{
			`
/users/k1LoW:
  get: null
`,
			nil,
			true,
		},
		{
			`
/users/k1LoW:
  post: 
    body: null
`,
			nil,
			true,
		},
	}

	for _, tt := range tests {
		var v map[string]interface{}
		if err := yaml.Unmarshal([]byte(tt.in), &v); err != nil {
			t.Fatal(err)
		}
		got, err := parseHTTPRequest(v)
		if err != nil {
			if !tt.wantErr {
				t.Error(err)
			}
			continue
		}
		if tt.wantErr {
			t.Error("want error")
		}
		opts := cmp.AllowUnexported(httpRequest{})
		if diff := cmp.Diff(got, tt.want, opts); diff != "" {
			t.Errorf("%s", diff)
		}
	}
}

func TestParseDBQuery(t *testing.T) {
	tests := []struct {
		in      string
		want    *dbQuery
		wantErr bool
	}{
		{
			`
query: SELECT * FROM users;
`,
			&dbQuery{
				stmt: "SELECT * FROM users;",
			},
			false,
		},
		{
			`
query: |
  SELECT * FROM users;
`,
			&dbQuery{
				stmt: "SELECT * FROM users;",
			},
			false,
		},
	}

	for _, tt := range tests {
		var v map[string]interface{}
		if err := yaml.Unmarshal([]byte(tt.in), &v); err != nil {
			t.Fatal(err)
		}
		got, err := parseDBQuery(v)
		if err != nil {
			if !tt.wantErr {
				t.Error(err)
			}
			continue
		}
		if tt.wantErr {
			t.Error("want error")
		}
		opts := cmp.AllowUnexported(dbQuery{})
		if diff := cmp.Diff(got, tt.want, opts); diff != "" {
			t.Errorf("%s", diff)
		}
	}
}

func TestParseGrpcRequest(t *testing.T) {
	tests := []struct {
		in      string
		want    *grpcRequest
		wantErr bool
	}{
		{
			`
my.custom.server.Service/Method:
  headers:
    user-agent: "runn/dev"
  message:
    key: value
    foo: bar
`,
			&grpcRequest{
				service: "my.custom.server.Service",
				method:  "Method",
				headers: metadata.MD{
					"user-agent": []string{"runn/dev"},
				},
				messages: []*grpcMessage{
					{
						op: grpcOpMessage,
						params: map[string]interface{}{
							"key": "value",
							"foo": "bar",
						},
					},
				},
			},
			false,
		},
		{
			`
my.custom.server.Service/Method:
  messages:
    - 
      key: value
      foo: bar
    - 
      one: two
`,
			&grpcRequest{
				service: "my.custom.server.Service",
				method:  "Method",
				headers: metadata.MD{},
				messages: []*grpcMessage{
					{
						op: grpcOpMessage,
						params: map[string]interface{}{
							"key": "value",
							"foo": "bar",
						},
					},
					{
						op: grpcOpMessage,
						params: map[string]interface{}{
							"one": "two",
						},
					},
				},
			},
			false,
		},
		{
			`
my.custom.server.Service/Method:
  messages:
    - 
      key: value
    - 
      recieve
    - 
      one: two
    - 
      close
`,
			&grpcRequest{
				service: "my.custom.server.Service",
				method:  "Method",
				headers: metadata.MD{},
				messages: []*grpcMessage{
					{
						op: grpcOpMessage,
						params: map[string]interface{}{
							"key": "value",
						},
					},
					{
						op: grpcOpRecieve,
					},
					{
						op: grpcOpMessage,
						params: map[string]interface{}{
							"one": "two",
						},
					},
					{
						op: grpcOpClose,
					},
				},
			},
			false,
		},
		{
			`
"{{ vars.path }}":
  headers:
    "{{ vars.one }}": "{{ vars.two }}"
  message:
    "{{ vars.one }}": "{{ vars.two }}"
    foo: bar
`,
			&grpcRequest{
				service: "my.custom.server.Service",
				method:  "Method",
				headers: metadata.MD{
					"ichi": []string{"ni"},
				},
				messages: []*grpcMessage{
					{
						op: grpcOpMessage,
						params: map[string]interface{}{
							"{{ vars.one }}": "{{ vars.two }}",
							"foo":            "bar",
						},
					},
				},
			},
			false,
		},
	}

	o, err := New()
	if err != nil {
		t.Fatal(err)
	}
	o.store.vars = map[string]interface{}{"path": "my.custom.server.Service/Method", "one": "ichi", "two": "ni"}

	for _, tt := range tests {
		var v map[string]interface{}
		if err := yaml.Unmarshal([]byte(tt.in), &v); err != nil {
			t.Fatal(err)
		}
		got, err := parseGrpcRequest(v, o.expand)
		if err != nil {
			if !tt.wantErr {
				t.Error(err)
			}
			continue
		}
		if tt.wantErr {
			t.Error("want error")
		}
		opts := cmp.AllowUnexported(grpcRequest{}, grpcMessage{})
		if diff := cmp.Diff(got, tt.want, opts); diff != "" {
			t.Errorf("%s", diff)
		}
	}
}

func TestParseExecCommand(t *testing.T) {
	tests := []struct {
		in      string
		want    *execCommand
		wantErr bool
	}{
		{
			`
command: echo hello > test.txt
`,
			&execCommand{
				command: "echo hello > test.txt",
			},
			false,
		},
		{
			`
command: echo hello > test.txt
stdin: |
  alice
  bob
  charlie
`,
			&execCommand{
				command: "echo hello > test.txt",
				stdin:   "alice\nbob\ncharlie\n",
			},
			false,
		},
		{
			`
stdin: |
  alice
  bob
  charlie
`,
			nil,
			true,
		},
	}

	for _, tt := range tests {
		var v map[string]interface{}
		if err := yaml.Unmarshal([]byte(tt.in), &v); err != nil {
			t.Fatal(err)
		}
		got, err := parseExecCommand(v)
		if err != nil {
			if !tt.wantErr {
				t.Error(err)
			}
			continue
		}
		if tt.wantErr {
			t.Error("want error")
		}
		opts := cmp.AllowUnexported(execCommand{})
		if diff := cmp.Diff(got, tt.want, opts); diff != "" {
			t.Errorf("%s", diff)
		}
	}
}

func TestTrimDelimiter(t *testing.T) {
	tests := []struct {
		in   map[string]interface{}
		want map[string]interface{}
	}{
		{
			map[string]interface{}{"k": `"Hello"`},
			map[string]interface{}{"k": "Hello"},
		},
		{
			map[string]interface{}{"k": `'Hello'`},
			map[string]interface{}{"k": "Hello"},
		},
		{
			map[string]interface{}{"k": `"'Hello'"`},
			map[string]interface{}{"k": "Hello"},
		},
		{
			map[string]interface{}{"k": `"'He\"llo'"`},
			map[string]interface{}{"k": "He\"llo"},
		},
		{
			map[string]interface{}{"k": `"\"Hello\""`},
			map[string]interface{}{"k": `Hello`},
		},
		{
			map[string]interface{}{"k": []interface{}{
				`"Hello"`,
				2,
			}},
			map[string]interface{}{"k": []interface{}{
				"Hello",
				2,
			}},
		},
	}
	for _, tt := range tests {
		got := trimDelimiter(tt.in)
		if diff := cmp.Diff(got, tt.want, nil); diff != "" {
			t.Errorf("%s", diff)
		}
	}
}

func TestParseServiceAndMethod(t *testing.T) {
	tests := []struct {
		in         string
		wantSvc    string
		wantMethod string
		wantErr    bool
	}{
		{"", "", "", true},
		{"my.custom.server.Service/Method", "my.custom.server.Service", "Method", false},
		{"/my.custom.server.Service/Method", "my.custom.server.Service", "Method", false},
	}
	for _, tt := range tests {
		gotSvc, gotMethod, err := parseServiceAndMethod(tt.in)
		if err != nil {
			if !tt.wantErr {
				t.Error(err)
			}
			continue
		}
		if tt.wantErr {
			t.Error("want error")
			continue
		}
		if gotSvc != tt.wantSvc {
			t.Errorf("got %v\nwant %v", gotSvc, tt.wantSvc)
		}
		if gotMethod != tt.wantMethod {
			t.Errorf("got %v\nwant %v", gotMethod, tt.wantMethod)
		}
	}
}
