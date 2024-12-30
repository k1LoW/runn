package runn

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/k1LoW/donegroup"
	"github.com/k1LoW/grpcstub"
	"github.com/k1LoW/runn/testutil"
	"github.com/k1LoW/runn/version"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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
		useTLS := useTLS
		for _, disableReflection := range []bool{true, false} {
			disableReflection := disableReflection
			for _, tt := range tests {
				tt := tt
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
		tt := tt
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
		tt := tt
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
