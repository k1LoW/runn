package runn

import (
	"net/http"
	"testing"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc/metadata"
)

func TestParseHTTPRequest(t *testing.T) {
	use := true
	notUse := false
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
				headers:   http.Header{},
				body: map[string]any{
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
				headers:   http.Header{},
				body:      nil,
				useCookie: nil,
				trace:     nil,
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
		{
			`
/users/k1LoW:
  get:
    body: null
    useCookie: true
`,
			&httpRequest{
				path:      "/users/k1LoW",
				method:    http.MethodGet,
				mediaType: "",
				headers:   http.Header{},
				body:      nil,
				useCookie: &use,
				trace:     nil,
			},
			false,
		},
		{
			`
/users/k1LoW:
  get:
    body: null
    trace: true
`,
			&httpRequest{
				path:      "/users/k1LoW",
				method:    http.MethodGet,
				mediaType: "",
				headers:   http.Header{},
				body:      nil,
				useCookie: nil,
				trace:     &use,
			},
			false,
		},
		{
			`
/users/k1LoW:
  get:
    body: null
    useCookie: nil
    trace: nil
`,
			nil,
			true,
		},
		{
			`
/users/k1LoW?page=2:
  get:
    body: null
    useCookie: false
    trace: false
`,
			&httpRequest{
				path:      "/users/k1LoW?page=2",
				method:    http.MethodGet,
				mediaType: "",
				headers:   http.Header{},
				body:      nil,
				useCookie: &notUse,
				trace:     &notUse,
			},
			false,
		},
		{
			`
/users/k1LoW:
  get:
    body: null
    useCookie: 1
    trace: true
`,
			nil,
			true,
		},
		{
			`
/users/k1LoW:
  get:
    body: null
    useCookie: true
    trace: "true"
`,
			nil,
			true,
		},
	}

	for _, tt := range tests {
		var v map[string]any
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
			t.Error(diff)
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
		var v map[string]any
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
			t.Error(diff)
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
						op: GRPCOpMessage,
						params: map[string]any{
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
						op: GRPCOpMessage,
						params: map[string]any{
							"key": "value",
							"foo": "bar",
						},
					},
					{
						op: GRPCOpMessage,
						params: map[string]any{
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
      receive
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
						op: GRPCOpMessage,
						params: map[string]any{
							"key": "value",
						},
					},
					{
						op: GRPCOpReceive,
					},
					{
						op: GRPCOpMessage,
						params: map[string]any{
							"one": "two",
						},
					},
					{
						op: GRPCOpClose,
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
						op: GRPCOpMessage,
						params: map[string]any{
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
	o.store.SetVar("path", "my.custom.server.Service/Method")
	o.store.SetVar("one", "ichi")
	o.store.SetVar("two", "ni")

	for _, tt := range tests {
		var v map[string]any
		if err := yaml.Unmarshal([]byte(tt.in), &v); err != nil {
			t.Fatal(err)
		}
		got, err := parseGrpcRequest(v, &step{}, o.expandBeforeRecord)
		if err != nil {
			if !tt.wantErr {
				t.Error(err)
			}
			continue
		}
		if tt.wantErr {
			t.Error("want error")
		}
		opts := []cmp.Option{
			cmp.AllowUnexported(grpcRequest{}, grpcMessage{}),
			cmpopts.IgnoreFields(grpcRequest{}, "mu"),
		}
		if diff := cmp.Diff(got, tt.want, opts...); diff != "" {
			t.Error(diff)
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
		var v map[string]any
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
			t.Error(diff)
		}
	}
}

func TestTrimDelimiter(t *testing.T) {
	tests := []struct {
		in   map[string]any
		want map[string]any
	}{
		{
			map[string]any{"k": `"Hello"`},
			map[string]any{"k": "Hello"},
		},
		{
			map[string]any{"k": `'Hello'`},
			map[string]any{"k": "Hello"},
		},
		{
			map[string]any{"k": `"'Hello'"`},
			map[string]any{"k": "Hello"},
		},
		{
			map[string]any{"k": `"'He\"llo'"`},
			map[string]any{"k": "He\"llo"},
		},
		{
			map[string]any{"k": `"\"Hello\""`},
			map[string]any{"k": `Hello`},
		},
		{
			map[string]any{"k": []any{
				`"Hello"`,
				2,
			}},
			map[string]any{"k": []any{
				"Hello",
				2,
			}},
		},
	}
	for _, tt := range tests {
		got := trimDelimiter(tt.in)
		if diff := cmp.Diff(got, tt.want, nil); diff != "" {
			t.Error(diff)
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

func TestParseDuration(t *testing.T) {
	tests := []struct {
		in      string
		want    time.Duration
		wantErr bool
	}{
		{"0", 0, false},
		{"3", 3 * time.Second, false},
		{"3min", 3 * time.Minute, false},
		{"3xxx", 0, true},
		{"0.5", 500 * time.Millisecond, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			got, err := parseDuration(tt.in)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("got %v\n", err)
				}
				return
			}
			if tt.wantErr {
				t.Error("want err\n")
				return
			}
			if diff := cmp.Diff(got, tt.want, nil); diff != "" {
				t.Error(diff)
			}
		})
	}
}
