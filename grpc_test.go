package runn

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/k1LoW/grpcstub"
	"github.com/k1LoW/runn/version"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestGrpcRunner(t *testing.T) {
	tests := []struct {
		req             *grpcRequest
		wantRecvRequest *grpcstub.Request
		wantResMessage  map[string]interface{}
		wantResHeaders  metadata.MD
	}{
		{
			&grpcRequest{
				service: "grpctest.GrpcTestService",
				method:  "Hello",
				headers: metadata.MD{"3rd": {"stone"}},
				messages: []*grpcMessage{
					{
						op: grpcOpMessage,
						params: map[string]interface{}{
							"name":         "alice",
							"num":          3,
							"request_time": time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC),
						},
					},
				},
			},
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
				"hello":        []string{"world"},
			},
		},
	}
	ctx := context.Background()
	ts := grpcstub.NewServer(t, []string{}, "testdata/grpctest.proto")
	t.Cleanup(func() {
		ts.Close()
	})
	ts.Method("grpctest.GrpcTestService/Hello").
		Header("hello", "world").
		ResponseString(`{"message":"hello", "num":32, "create_time":"2022-06-25T05:24:43.861872Z"}`)
	o, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for i, tt := range tests {
		r, err := newGrpcRunner("greq", ts.Addr(), o)
		if err != nil {
			t.Fatal(err)
		}
		if err := r.Run(ctx, tt.req); err != nil {
			t.Error(err)
		}
		if want := i + 1; len(r.operator.store.steps) != want {
			t.Errorf("got %v want %v", len(r.operator.store.steps), want)
			continue
		}
		if len(ts.Requests()) == 0 {
			t.Fatal("want requests")
		}
		latest := len(ts.Requests()) - 1
		recvReq := ts.Requests()[latest]
		tt.wantRecvRequest.Headers.Append(":authority", ts.Addr())
		if diff := cmp.Diff(recvReq, tt.wantRecvRequest, nil); diff != "" {
			t.Errorf("%s", diff)
		}

		res := r.operator.store.steps[i]["res"].(map[string]interface{})
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
	}
}
