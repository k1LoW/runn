package runn

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
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
		wantResMessage  map[string]interface{}
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
						params: map[string]interface{}{
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
					"content-type": {"application/grpc"},
					"3rd":          {"stone"},
					"user-agent":   {fmt.Sprintf("runn/%s grpc-go/%s", version.Version, grpc.Version)},
				},
				Message: grpcstub.Message{
					"name":         "alice",
					"num":          float64(3),
					"request_time": time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC).Format(time.RFC3339Nano),
				},
			},
			map[string]interface{}{
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
						params: map[string]interface{}{
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
					"content-type": {"application/grpc"},
					"101000":       {"lab"},
					"user-agent":   {fmt.Sprintf("runn/%s grpc-go/%s", version.Version, grpc.Version)},
				},
				Message: grpcstub.Message{
					"name":         "alice",
					"num":          float64(3),
					"request_time": time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC).Format(time.RFC3339Nano),
				},
			},
			map[string]interface{}{
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
						params: map[string]interface{}{
							"name":         "alice",
							"num":          3,
							"request_time": time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC),
						},
					},
					{
						op: GRPCOpMessage,
						params: map[string]interface{}{
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
					"content-type": {"application/grpc"},
					"101000":       {"lab"},
					"user-agent":   {fmt.Sprintf("runn/%s grpc-go/%s", version.Version, grpc.Version)},
				},
				Message: grpcstub.Message{
					"name":         "bob",
					"num":          float64(4),
					"request_time": time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC).Format(time.RFC3339Nano),
				},
			},
			map[string]interface{}{
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
						params: map[string]interface{}{
							"name":         "alice",
							"num":          3,
							"request_time": time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC),
						},
					},
					{
						op:     GRPCOpReceive,
						params: map[string]interface{}{},
					},
					{
						op: GRPCOpMessage,
						params: map[string]interface{}{
							"name":         "bob",
							"num":          4,
							"request_time": time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC),
						},
					},
					{
						op:     GRPCOpReceive,
						params: map[string]interface{}{},
					},
					{
						op:     GRPCOpClose,
						params: map[string]interface{}{},
					},
				},
			},
			2,
			2,
			&grpcstub.Request{
				Service: "grpctest.GrpcTestService",
				Method:  "HelloChat",
				Headers: metadata.MD{
					"content-type": {"application/grpc"},
					"101000":       {"lab"},
					"user-agent":   {fmt.Sprintf("runn/%s grpc-go/%s", version.Version, grpc.Version)},
				},
				Message: grpcstub.Message{
					"name":         "bob",
					"num":          float64(4),
					"request_time": time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC).Format(time.RFC3339Nano),
				},
			},
			map[string]interface{}{
				"message":     "hello",
				"num":         float64(35),
				"create_time": time.Date(2022, 6, 25, 5, 24, 47, 382783000, time.UTC).Format(time.RFC3339Nano),
			},
			metadata.MD{
				"content-type": []string{"application/grpc"},
				"hellochat":    []string{"header"},
			},
			metadata.MD{},
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ts := testutil.GRPCServer(t, false)
			o, err := New()
			if err != nil {
				t.Fatal(err)
			}
			r, err := newGrpcRunner("greq", ts.Addr())
			if err != nil {
				t.Fatal(err)
			}
			r.operator = o
			if err := r.Run(ctx, tt.req); err != nil {
				t.Error(err)
			}
			if want := 1; len(r.operator.store.steps) != want {
				t.Errorf("got %v want %v", len(r.operator.store.steps), want)
				return
			}
			{
				got := len(ts.Requests())
				if got != tt.wantReqCount {
					t.Errorf("got %v\nwant %v", got, tt.wantReqCount)
				}
			}
			latest := len(ts.Requests()) - 1
			recvReq := ts.Requests()[latest]
			recvReq.Headers.Delete(":authority")
			if diff := cmp.Diff(recvReq, tt.wantRecvRequest, nil); diff != "" {
				t.Errorf("%s", diff)
			}

			res := r.operator.store.steps[0]["res"].(map[string]interface{})
			{
				got := len(res["messages"].([]map[string]interface{}))
				if got != tt.wantResCount {
					t.Errorf("got %v\nwant %v", got, tt.wantResCount)
				}
			}
			{
				got := res["message"].(map[string]interface{})
				if diff := cmp.Diff(got, tt.wantResMessage, nil); diff != "" {
					t.Errorf("%s", diff)
				}
			}
			{
				got := res["headers"].(metadata.MD)
				if diff := cmp.Diff(got, tt.wantResHeaders, nil); diff != "" {
					t.Errorf("%s", diff)
				}
			}
			{
				got := res["trailers"].(metadata.MD)
				if diff := cmp.Diff(got, tt.wantResTrailers, nil); diff != "" {
					t.Errorf("%s", diff)
				}
			}
		})

		t.Run(fmt.Sprintf("%s with TLS", tt.name), func(t *testing.T) {
			t.Parallel()
			useTLS := true
			ts := testutil.GRPCServer(t, useTLS)
			o, err := New()
			if err != nil {
				t.Fatal(err)
			}
			r, err := newGrpcRunner("greq", ts.Addr())
			if err != nil {
				t.Fatal(err)
			}
			r.operator = o
			r.tls = &useTLS
			r.cacert = testutil.Cacert
			r.cert = testutil.Cert
			r.key = testutil.Key
			r.skipVerify = false
			if err := r.Run(ctx, tt.req); err != nil {
				t.Error(err)
			}
			if want := 1; len(r.operator.store.steps) != want {
				t.Errorf("got %v want %v", len(r.operator.store.steps), want)
				return
			}
			{
				got := len(ts.Requests())
				if got != tt.wantReqCount {
					t.Errorf("got %v\nwant %v", got, tt.wantReqCount)
				}
			}
			latest := len(ts.Requests()) - 1
			recvReq := ts.Requests()[latest]
			recvReq.Headers.Delete(":authority")
			if diff := cmp.Diff(recvReq, tt.wantRecvRequest, nil); diff != "" {
				t.Errorf("%s", diff)
			}

			res := r.operator.store.steps[0]["res"].(map[string]interface{})
			{
				got := len(res["messages"].([]map[string]interface{}))
				if got != tt.wantResCount {
					t.Errorf("got %v\nwant %v", got, tt.wantResCount)
				}
			}
			{
				got := res["message"].(map[string]interface{})
				if diff := cmp.Diff(got, tt.wantResMessage, nil); diff != "" {
					t.Errorf("%s", diff)
				}
			}
			{
				got := res["headers"].(metadata.MD)
				if diff := cmp.Diff(got, tt.wantResHeaders, nil); diff != "" {
					t.Errorf("%s", diff)
				}
			}
			{
				got := res["trailers"].(metadata.MD)
				if diff := cmp.Diff(got, tt.wantResTrailers, nil); diff != "" {
					t.Errorf("%s", diff)
				}
			}
		})
	}
}
