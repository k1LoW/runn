package runn

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/k1LoW/donegroup"
	"github.com/k1LoW/grpcstub"
	"github.com/k1LoW/runn/testutil"
	"github.com/k1LoW/runn/version"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

func TestGrpcRunner(t *testing.T) {
	tests := []struct {
		name            string
		req             *grpcRequest
		wantReqCount    int
		wantResCount    int
		wantRecvRequest *grpcstub.Request
		wantResMessage  map[string]any
		wantResHeaders  metadata.MD
		wantResTrailers metadata.MD
	}{
		{
			"Unary RPC",
			&grpcRequest{
				service: "grpctest.GrpcTestService",
				method:  "Hello",
				headers: metadata.MD{"3rd": {"stone"}},
				messages: []*grpcMessage{
					{
						op: GRPCOpMessage,
						params: map[string]any{
							"name":         "alice",
							"num":          3,
							"request_time": time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC),
						},
					},
				},
			},
			1,
			1,
			&grpcstub.Request{
				Service: "grpctest.GrpcTestService",
				Method:  "Hello",
				Headers: metadata.MD{
					"content-type":         {"application/grpc"},
					"grpc-accept-encoding": {"gzip"},
					"3rd":                  {"stone"},
					"user-agent":           {fmt.Sprintf("runn/%s grpc-go/%s", version.Version, grpc.Version)},
				},
				Message: grpcstub.Message{
					"name":         "alice",
					"num":          float64(3),
					"request_time": time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC).Format(time.RFC3339Nano),
				},
			},
			map[string]any{
				"message":     "hello",
				"num":         float64(32),
				"create_time": time.Date(2022, 6, 25, 5, 24, 43, 861872000, time.UTC).Format(time.RFC3339Nano),
			},
			metadata.MD{
				"content-type": []string{"application/grpc"},
				"hello":        []string{"header"},
			},
			metadata.MD{
				"hello": []string{"trailer"},
			},
		},
		{
			"Server streaming RPC",
			&grpcRequest{
				service: "grpctest.GrpcTestService",
				method:  "ListHello",
				headers: metadata.MD{"101000": {"lab"}},
				messages: []*grpcMessage{
					{
						op: GRPCOpMessage,
						params: map[string]any{
							"name":         "alice",
							"num":          3,
							"request_time": time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC),
						},
					},
				},
			},
			1,
			2,
			&grpcstub.Request{
				Service: "grpctest.GrpcTestService",
				Method:  "ListHello",
				Headers: metadata.MD{
					"content-type":         {"application/grpc"},
					"grpc-accept-encoding": {"gzip"},
					"101000":               {"lab"},
					"user-agent":           {fmt.Sprintf("runn/%s grpc-go/%s", version.Version, grpc.Version)},
				},
				Message: grpcstub.Message{
					"name":         "alice",
					"num":          float64(3),
					"request_time": time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC).Format(time.RFC3339Nano),
				},
			},
			map[string]any{
				"message":     "hello",
				"num":         float64(34),
				"create_time": time.Date(2022, 6, 25, 5, 24, 44, 382783000, time.UTC).Format(time.RFC3339Nano),
			},
			metadata.MD{
				"content-type": []string{"application/grpc"},
				"listhello":    []string{"header"},
			},
			metadata.MD{
				"listhello": []string{"trailer"},
			},
		},
		{
			"Client streaming RPC",
			&grpcRequest{
				service: "grpctest.GrpcTestService",
				method:  "MultiHello",
				headers: metadata.MD{"101000": {"lab"}},
				messages: []*grpcMessage{
					{
						op: GRPCOpMessage,
						params: map[string]any{
							"name":         "alice",
							"num":          3,
							"request_time": time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC),
						},
					},
					{
						op: GRPCOpMessage,
						params: map[string]any{
							"name":         "bob",
							"num":          4,
							"request_time": time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC),
						},
					},
				},
			},
			2,
			1,
			&grpcstub.Request{
				Service: "grpctest.GrpcTestService",
				Method:  "MultiHello",
				Headers: metadata.MD{
					"content-type":         {"application/grpc"},
					"grpc-accept-encoding": {"gzip"},
					"101000":               {"lab"},
					"user-agent":           {fmt.Sprintf("runn/%s grpc-go/%s", version.Version, grpc.Version)},
				},
				Message: grpcstub.Message{
					"name":         "bob",
					"num":          float64(4),
					"request_time": time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC).Format(time.RFC3339Nano),
				},
			},
			map[string]any{
				"message":     "hello",
				"num":         float64(35),
				"create_time": time.Date(2022, 6, 25, 5, 24, 45, 382783000, time.UTC).Format(time.RFC3339Nano),
			},
			metadata.MD{
				"content-type": []string{"application/grpc"},
				"multihello":   []string{"header"},
			},
			metadata.MD{
				"multihello": []string{"trailer"},
			},
		},
		{
			"Bidirectional streaming RPC",
			&grpcRequest{
				service: "grpctest.GrpcTestService",
				method:  "HelloChat",
				headers: metadata.MD{"101000": {"lab"}},
				messages: []*grpcMessage{
					{
						op: GRPCOpMessage,
						params: map[string]any{
							"name":         "alice",
							"num":          3,
							"request_time": time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC),
						},
					},
					{
						op:     GRPCOpReceive,
						params: map[string]any{},
					},
					{
						op: GRPCOpMessage,
						params: map[string]any{
							"name":         "bob",
							"num":          4,
							"request_time": time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC),
						},
					},
					{
						op:     GRPCOpReceive,
						params: map[string]any{},
					},
					{
						op:     GRPCOpClose,
						params: map[string]any{},
					},
				},
			},
			2,
			2,
			&grpcstub.Request{
				Service: "grpctest.GrpcTestService",
				Method:  "HelloChat",
				Headers: metadata.MD{
					"content-type":         {"application/grpc"},
					"grpc-accept-encoding": {"gzip"},
					"101000":               {"lab"},
					"user-agent":           {fmt.Sprintf("runn/%s grpc-go/%s", version.Version, grpc.Version)},
				},
				Message: grpcstub.Message{
					"name":         "bob",
					"num":          float64(4),
					"request_time": time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC).Format(time.RFC3339Nano),
				},
			},
			map[string]any{
				"message":     "hello",
				"num":         float64(35),
				"create_time": time.Date(2022, 6, 25, 5, 24, 47, 382783000, time.UTC).Format(time.RFC3339Nano),
			},
			metadata.MD{
				"content-type": []string{"application/grpc"},
				"hellochat":    []string{"header"},
			},
			metadata.MD{
				"hellochat":        []string{"trailer"},
				"hellochat-second": []string{"trailer"},
			},
		},
	}

	for _, useTLS := range []bool{true, false} {
		for _, disableReflection := range []bool{true, false} {
			for _, tt := range tests {
				t.Run(fmt.Sprintf("%s (useTLS: %v, disableReflection: %v)", tt.name, useTLS, disableReflection), func(t *testing.T) {
					t.Parallel()
					ctx, cancel := donegroup.WithCancel(context.Background())
					t.Cleanup(cancel)
					ts := testutil.GRPCServer(t, useTLS, disableReflection)
					o, err := New()
					if err != nil {
						t.Fatal(err)
					}
					r, err := newGrpcRunner("greq", ts.Addr())
					if err != nil {
						t.Fatal(err)
					}
					r.tls = &useTLS
					r.cacert = testutil.Cacert
					r.cert = testutil.Cert
					r.key = testutil.Key
					r.bufLocks = append(r.bufLocks, filepath.Join(testutil.Testdata(), "buf.lock"))
					r.skipVerify = false
					if disableReflection {
						r.protos = []string{filepath.Join(testutil.Testdata(), "grpctest.proto")}
					}
					s := newStep(0, "stepKey", o, nil)
					if err := r.run(ctx, tt.req, s); err != nil {
						t.Error(err)
					}
					if want := 1; o.store.StepLen() != want {
						t.Errorf("got %v want %v", o.store.StepLen(), want)
						return
					}
					{
						got := len(ts.Requests())
						if got != tt.wantReqCount {
							t.Errorf("got %v\nwant %v", got, tt.wantReqCount)
							return
						}
					}
					latest := len(ts.Requests()) - 1
					recvReq := ts.Requests()[latest]
					recvReq.Headers.Delete(":authority")
					if diff := cmp.Diff(recvReq, tt.wantRecvRequest, nil); diff != "" {
						t.Error(diff)
					}

					sm := o.store.ToMap()
					sl, ok := sm["steps"].([]map[string]any)
					if !ok {
						t.Fatal("steps not found")
					}
					res, ok := sl[0]["res"].(map[string]any)
					if !ok {
						t.Fatalf("invalid steps res: %v", sl[0]["res"])
					}
					{
						msgs, ok := res["messages"].([]map[string]any)
						if !ok {
							t.Fatalf("invalid res messages: %v", res["messages"])
						}
						got := len(msgs)
						if got != tt.wantResCount {
							t.Errorf("got %v\nwant %v", got, tt.wantResCount)
						}
					}
					{
						got, ok := res["message"].(map[string]any)
						if !ok {
							t.Fatalf("invalid res message: %v", res["message"])
						}
						if diff := cmp.Diff(got, tt.wantResMessage, nil); diff != "" {
							t.Error(diff)
						}
					}
					{
						got, ok := res["headers"].(metadata.MD)
						if !ok {
							t.Fatalf("invalid res headers: %v", res["headers"])
						}
						if diff := cmp.Diff(got, tt.wantResHeaders, nil); diff != "" {
							t.Error(diff)
						}
					}
					{
						got, ok := res["trailers"].(metadata.MD)
						if !ok {
							t.Fatalf("invalid res trailers: %v", res["trailers"])
						}
						if diff := cmp.Diff(got, tt.wantResTrailers, nil); diff != "" {
							t.Error(diff)
						}
					}
				})
			}
		}
	}
}

func TestGrpcRunnerWithTimeout(t *testing.T) {
	tests := []struct {
		name string
		req  *grpcRequest
	}{
		{
			"Timeout Unary RPC",
			&grpcRequest{
				service: "grpctest.GrpcTestService",
				method:  "Hello",
				headers: metadata.MD{"slow": []string{"enable"}},
				messages: []*grpcMessage{
					{
						op: GRPCOpMessage,
						params: map[string]any{
							"name": "slowhello",
						},
					},
				},
				timeout: 1 * time.Millisecond,
			},
		},
		{
			"Timeout Server streaming RPC",
			&grpcRequest{
				service: "grpctest.GrpcTestService",
				method:  "ListHello",
				headers: metadata.MD{"slow": {"enable"}},
				messages: []*grpcMessage{
					{
						op: GRPCOpMessage,
						params: map[string]any{
							"name": "slowhello",
						},
					},
				},
				timeout: 1 * time.Millisecond,
			},
		},
		{
			"Timeout Client streaming RPC",
			&grpcRequest{
				service: "grpctest.GrpcTestService",
				method:  "MultiHello",
				headers: metadata.MD{"slow": {"enable"}},
				messages: []*grpcMessage{
					{
						op: GRPCOpMessage,
						params: map[string]any{
							"name": "slowhello",
						},
					},
					{
						op: GRPCOpMessage,
						params: map[string]any{
							"name": "slowhello",
						},
					},
				},
				timeout: 1 * time.Millisecond,
			},
		},
		{
			"Timeout grpc.health.v1.Health.Watch (Server streaming RPC)",
			&grpcRequest{
				service: "grpc.health.v1.Health",
				method:  "Watch",
				headers: metadata.MD{},
				messages: []*grpcMessage{
					{
						op: GRPCOpMessage,
						params: map[string]any{
							"service": grpcstub.HealthCheckService_FLAPPING,
						},
					},
				},
				timeout: 1 * time.Millisecond,
			},
		},
	}

	useTLS := false
	ts := testutil.GRPCServer(t, useTLS, false)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := donegroup.WithCancel(context.Background())
			t.Cleanup(cancel)
			o, err := New()
			if err != nil {
				t.Fatal(err)
			}
			r, err := newGrpcRunner("greq", ts.Addr())
			if err != nil {
				t.Fatal(err)
			}
			r.tls = &useTLS

			now := time.Now()
			s := newStep(0, "stepKey", o, nil)
			if err := r.run(ctx, tt.req, s); err != nil {
				t.Error(err)
			}
			got := time.Since(now).Milliseconds()
			if got > (10 * time.Millisecond.Milliseconds()) {
				t.Errorf("got %d msec want 10 msec", time.Since(now).Milliseconds())
				return
			}
			if want := 1; o.store.StepLen() != want {
				t.Errorf("got %v want %v", o.store.StepLen(), want)
				return
			}
		})
	}
}

func TestGrpcTraceHeader(t *testing.T) {
	tests := []struct {
		name string
		req  *grpcRequest
	}{
		{
			"Unary RPC",
			&grpcRequest{
				service: "grpctest.GrpcTestService",
				method:  "Hello",
				headers: metadata.MD{"3rd": {"stone"}},
				messages: []*grpcMessage{
					{
						op: GRPCOpMessage,
						params: map[string]any{
							"name":         "alice",
							"num":          3,
							"request_time": time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC),
						},
					},
				},
			},
		},
		{
			"Server streaming RPC",
			&grpcRequest{
				service: "grpctest.GrpcTestService",
				method:  "ListHello",
				headers: metadata.MD{"101000": {"lab"}},
				messages: []*grpcMessage{
					{
						op: GRPCOpMessage,
						params: map[string]any{
							"name":         "alice",
							"num":          3,
							"request_time": time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC),
						},
					},
				},
			},
		},
		{
			"Client streaming RPC",
			&grpcRequest{
				service: "grpctest.GrpcTestService",
				method:  "MultiHello",
				headers: metadata.MD{"101000": {"lab"}},
				messages: []*grpcMessage{
					{
						op: GRPCOpMessage,
						params: map[string]any{
							"name":         "alice",
							"num":          3,
							"request_time": time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC),
						},
					},
					{
						op: GRPCOpMessage,
						params: map[string]any{
							"name":         "bob",
							"num":          4,
							"request_time": time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC),
						},
					},
				},
			},
		},
		{
			"Bidirectional streaming RPC",
			&grpcRequest{
				service: "grpctest.GrpcTestService",
				method:  "HelloChat",
				headers: metadata.MD{"101000": {"lab"}},
				messages: []*grpcMessage{
					{
						op: GRPCOpMessage,
						params: map[string]any{
							"name":         "alice",
							"num":          3,
							"request_time": time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC),
						},
					},
					{
						op:     GRPCOpReceive,
						params: map[string]any{},
					},
					{
						op: GRPCOpMessage,
						params: map[string]any{
							"name":         "bob",
							"num":          4,
							"request_time": time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC),
						},
					},
					{
						op:     GRPCOpReceive,
						params: map[string]any{},
					},
					{
						op:     GRPCOpClose,
						params: map[string]any{},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := donegroup.WithCancel(context.Background())
			t.Cleanup(cancel)
			ts := testutil.GRPCServer(t, false, false)
			o, err := New()
			if err != nil {
				t.Fatal(err)
			}
			r, err := newGrpcRunner("greq", ts.Addr())
			if err != nil {
				t.Fatal(err)
			}
			useTLS := false
			trace := true
			r.tls = &useTLS
			r.trace = &trace
			s := newStep(0, "stepKey", o, nil)
			if err := r.run(ctx, tt.req, s); err != nil {
				t.Error(err)
			}
			latest := len(ts.Requests()) - 1
			recvReq := ts.Requests()[latest]
			if len(recvReq.Headers.Get(DefaultTraceHeaderName)) != 1 {
				t.Error("got empty trace header")
			}
		})
	}
}

func TestGrpcRunnerReflectionWithAuth(t *testing.T) {
	t.Parallel()

	const token = "Bearer runn-test"
	authErr := status.Error(codes.Unauthenticated, "auth required")
	checkAuth := func(ctx context.Context) bool {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return false
		}
		a := md.Get("authorization")
		return len(a) > 0 && a[0] == token
	}

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, h grpc.UnaryHandler) (any, error) {
			if !checkAuth(ctx) {
				return nil, authErr
			}
			return h(ctx, req)
		}),
		grpc.StreamInterceptor(func(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, h grpc.StreamHandler) error {
			if !checkAuth(ss.Context()) {
				return authErr
			}
			return h(srv, ss)
		}),
	)
	healthpb.RegisterHealthServer(srv, health.NewServer())
	reflection.Register(srv)

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() { _ = srv.Serve(l) }()
	t.Cleanup(srv.Stop)
	addr := l.Addr().String()

	t.Run("step headers are forwarded to reflection RPC", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := donegroup.WithCancel(context.Background())
		t.Cleanup(cancel)
		o, err := New()
		if err != nil {
			t.Fatal(err)
		}
		r, err := newGrpcRunner("greq", addr)
		if err != nil {
			t.Fatal(err)
		}
		useTLS := false
		r.tls = &useTLS
		req := &grpcRequest{
			service:  "grpc.health.v1.Health",
			method:   "Check",
			headers:  metadata.MD{"authorization": {token}},
			messages: []*grpcMessage{{op: GRPCOpMessage, params: map[string]any{}}},
		}
		s := newStep(0, "stepKey", o, nil)
		if err := r.run(ctx, req, s); err != nil {
			t.Errorf("run() with auth header failed: %v", err)
		}
	})

	t.Run("reflection RPC fails without step headers", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := donegroup.WithCancel(context.Background())
		t.Cleanup(cancel)
		o, err := New()
		if err != nil {
			t.Fatal(err)
		}
		r, err := newGrpcRunner("greq", addr)
		if err != nil {
			t.Fatal(err)
		}
		useTLS := false
		r.tls = &useTLS
		req := &grpcRequest{
			service:  "grpc.health.v1.Health",
			method:   "Check",
			messages: []*grpcMessage{{op: GRPCOpMessage, params: map[string]any{}}},
		}
		s := newStep(0, "stepKey", o, nil)
		if err := r.run(ctx, req, s); err == nil {
			t.Error("run() without auth header should fail")
		}
	})
}

func TestGrpcRunnerReusableLifecycle(t *testing.T) {
	ts := testutil.GRPCServer(t, false, false)

	t.Run("copyOperators sets reusable flag and keeps connection across Close", func(t *testing.T) {
		ctx, cancel := donegroup.WithCancel(context.Background())
		t.Cleanup(cancel)

		o, err := New(Book("testdata/book/always_success.yml"))
		if err != nil {
			t.Fatal(err)
		}

		r, err := newGrpcRunner("greq", ts.Addr())
		if err != nil {
			t.Fatal(err)
		}
		useTLS := false
		r.tls = &useTLS
		o.grpcRunners["greq"] = r

		if r.reusable {
			t.Error("reusable should be false before copyOperators")
		}

		if err := r.connectAndResolve(ctx, o); err != nil {
			t.Fatal(err)
		}
		if r.cc == nil {
			t.Fatal("cc should not be nil after connectAndResolve")
		}

		copied, err := copyOperators([]*operator{o}, nil)
		if err != nil {
			t.Fatal(err)
		}
		copiedRunner := copied[0].grpcRunners["greq"]
		if copiedRunner != r {
			t.Error("copied operator should share the same grpcRunner instance")
		}
		if !r.reusable {
			t.Error("reusable should be true after copyOperators")
		}

		// Close(false) should NOT close reusable runner
		copied[0].Close(false)
		if r.cc == nil {
			t.Error("Close(false) should not close reusable runner connection")
		}

		// Close(true) should also NOT close reusable runner
		copied[0].Close(true)
		if r.cc == nil {
			t.Error("Close(true) should not close reusable runner connection")
		}
	})

	t.Run("Terminate closes reusable runners", func(t *testing.T) {
		ctx, cancel := donegroup.WithCancel(context.Background())
		t.Cleanup(cancel)

		o, err := New()
		if err != nil {
			t.Fatal(err)
		}

		r, err := newGrpcRunner("greq", ts.Addr())
		if err != nil {
			t.Fatal(err)
		}
		useTLS := false
		r.tls = &useTLS
		r.reusable = true
		o.grpcRunners["greq"] = r

		if err := r.connectAndResolve(ctx, o); err != nil {
			t.Fatal(err)
		}
		if r.cc == nil {
			t.Fatal("cc should not be nil after connectAndResolve")
		}

		opn := &operatorN{
			ops: []*operator{o},
		}

		// Close (via runN defer) should NOT close reusable runner
		opn.Close()
		if r.cc == nil {
			t.Error("operatorN.Close() should not close reusable runner connection")
		}

		// Terminate should close reusable runner
		if err := opn.Terminate(); err != nil {
			t.Fatal(err)
		}
		if r.cc != nil {
			t.Error("Terminate() should close reusable runner connection")
		}
	})
}
