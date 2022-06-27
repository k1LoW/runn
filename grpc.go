package runn

import (
	"context"

	"google.golang.org/grpc"
)

type grpcOp string

const (
	grpcOpMessage grpcOp = "message"
	grpcOpWait    grpcOp = "wait"
	grpcOpExit    grpcOp = "exit"
)

type grpcRunner struct {
	name     string
	target   string
	cc       *grpc.ClientConn
	operator *operator
}

type grpcMessage struct {
	op   grpcOp
	args map[string]interface{}
}

type grpcRequest struct {
	service  string
	method   string
	headers  map[string]string
	trailers map[string]string
	messages []*grpcMessage
}

func newGrpcRunner(name, target string, o *operator) (*grpcRunner, error) {
	return &grpcRunner{
		name:     name,
		target:   target,
		operator: o,
	}, nil
}

func (rnr *grpcRunner) Close() error {
	if rnr.cc == nil {
		return nil
	}
	return rnr.cc.Close()
}

func (rnr *grpcRunner) Run(ctx context.Context, r *grpcRequest) error {
	if rnr.cc == nil {
		cc, err := grpc.Dial(rnr.target, grpc.WithInsecure())
		if err != nil {
			return err
		}
		rnr.cc = cc
	}
	return nil
}
