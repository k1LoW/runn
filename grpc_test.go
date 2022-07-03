package runn

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/k1LoW/grpcstub"
	"google.golang.org/grpc/metadata"
)

func TestGrpcRunner(t *testing.T) {
	tests := []struct {
		req                   *grpcRequest
		wantRecvLatestMessage grpcstub.Message
	}{
		{
			&grpcRequest{
				service: "grpctest.GrpcTestService",
				method:  "Hello",
				headers: metadata.MD{},
				messages: []*grpcMessage{
					{
						op: grpcOpMessage,
						args: map[string]interface{}{
							"message":     "hello",
							"num":         32,
							"create_time": time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC),
						},
					},
				},
			},
			grpcstub.Message{
				"message":     "hello",
				"num":         32,
				"create_time": time.Date(2022, 2, 22, 22, 22, 22, 22, time.UTC),
			},
		},
	}
	ctx := context.Background()
	ts := grpcstub.NewServer(t, []string{}, "testdata/grpctest.proto")
	t.Cleanup(func() {
		ts.Close()
	})
	ts.Method("grpctest.GrpcTestService/Hello").ResponseString(`{"name":"alice", "num":3, "request_time":"2022-06-25T05:24:43.861872Z"}`)
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
		if diff := cmp.Diff(recvReq.Message, tt.wantRecvLatestMessage, nil); diff != "" {
			t.Errorf("%s", diff)
		}
	}
}
