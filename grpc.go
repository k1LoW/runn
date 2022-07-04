package runn

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/golang/protobuf/jsonpb" //nolint
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"github.com/jhump/protoreflect/grpcreflect"
	"github.com/k1LoW/runn/version"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	rpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/grpc/status"
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
	grefc    *grpcreflect.Client
	mds      map[string]*desc.MethodDescriptor
	operator *operator
}

type grpcMessage struct {
	op     grpcOp
	params map[string]interface{}
}

type grpcRequest struct {
	service  string
	method   string
	headers  metadata.MD
	messages []*grpcMessage
}

func newGrpcRunner(name, target string, o *operator) (*grpcRunner, error) {
	return &grpcRunner{
		name:     name,
		target:   target,
		mds:      map[string]*desc.MethodDescriptor{},
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
		cc, err := grpc.Dial(rnr.target, grpc.WithInsecure(), grpc.WithUserAgent(fmt.Sprintf("runn/%s", version.Version)))
		if err != nil {
			return err
		}
		rnr.cc = cc
		stub := rpb.NewServerReflectionClient(rnr.cc)
		rnr.grefc = grpcreflect.NewClient(ctx, stub)
		if err := rnr.resolveAllMethods(ctx); err != nil {
			return err
		}
	}
	key := strings.Join([]string{r.service, r.method}, "/")
	md, ok := rnr.mds[key]
	if !ok {
		return fmt.Errorf("can not find method: %s", key)
	}
	var ext dynamic.ExtensionRegistry
	alreadyFetched := map[string]bool{}
	if err := fetchAllExtensions(rnr.grefc, &ext, md.GetInputType(), alreadyFetched); err != nil {
		return err
	}
	if err := fetchAllExtensions(rnr.grefc, &ext, md.GetOutputType(), alreadyFetched); err != nil {
		return err
	}
	mf := dynamic.NewMessageFactoryWithExtensionRegistry(&ext)
	switch {
	case !md.IsServerStreaming() && !md.IsClientStreaming():
		return rnr.invokeUnary(ctx, md, mf, r)
	case md.IsServerStreaming() && !md.IsClientStreaming():
		return errors.New("not implemented")
	case !md.IsServerStreaming() && md.IsClientStreaming():
		return errors.New("not implemented")
	case md.IsServerStreaming() && md.IsClientStreaming():
		return errors.New("not implemented")
	default:
		return errors.New("something strange happened")
	}
}

func (rnr *grpcRunner) invokeUnary(ctx context.Context, md *desc.MethodDescriptor, mf *dynamic.MessageFactory, r *grpcRequest) error {
	if len(r.messages) != 1 {
		return errors.New("unary RPC message should be 1")
	}
	req := mf.NewMessage(md.GetInputType())
	e, err := rnr.operator.expand(r.messages[0].params)
	if err != nil {
		return err
	}
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	if err := jsonpb.Unmarshal(bytes.NewBuffer(b), req); err != nil {
		return err
	}
	stub := grpcdynamic.NewStubWithMessageFactory(rnr.cc, mf)
	var (
		resHeaders  metadata.MD
		resTrailers metadata.MD
	)
	kv := []string{}
	for k, v := range r.headers {
		kv = append(kv, k)
		kv = append(kv, v...)
	}
	ctx = metadata.AppendToOutgoingContext(ctx, kv...)
	res, err := stub.InvokeRpc(ctx, md, req, grpc.Header(&resHeaders), grpc.Trailer(&resTrailers))
	stat, ok := status.FromError(err)
	if !ok {
		return err
	}
	d := map[string]interface{}{
		"status":   stat.Code(),
		"headers":  resHeaders,
		"trailers": resTrailers,
		"message":  nil,
	}
	messages := []map[string]interface{}{}
	if stat.Code() == codes.OK {
		m := new(bytes.Buffer)
		marshaler := jsonpb.Marshaler{
			OrigName: true,
		}
		if err := marshaler.Marshal(m, res); err != nil {
			return err
		}
		var b map[string]interface{}
		if err := json.Unmarshal(m.Bytes(), &b); err != nil {
			return err
		}
		d["message"] = b
		messages = append(messages, b)
		d["messages"] = messages
	}

	rnr.operator.record(map[string]interface{}{
		"res": d,
	})
	return nil
}

func (rnr *grpcRunner) resolveAllMethods(ctx context.Context) error {
	svcs, err := rnr.grefc.ListServices()
	if err != nil {
		return err
	}
	for _, svc := range svcs {
		sd, err := rnr.grefc.ResolveService(svc)
		if err != nil {
			return err
		}
		mds := sd.GetMethods()
		for _, md := range mds {
			key := strings.Join([]string{sd.GetFullyQualifiedName(), md.GetName()}, "/")
			rnr.mds[key] = md
		}
	}
	return nil
}

func fetchAllExtensions(client *grpcreflect.Client, ext *dynamic.ExtensionRegistry, md *desc.MessageDescriptor, alreadyFetched map[string]bool) error {
	msgTypeName := md.GetFullyQualifiedName()
	if alreadyFetched[msgTypeName] {
		return nil
	}
	alreadyFetched[msgTypeName] = true
	if len(md.GetExtensionRanges()) > 0 {
		var fds []*desc.FieldDescriptor
		nums, err := client.AllExtensionNumbersForType(msgTypeName)
		if err != nil {
			return err
		}
		for _, fieldNum := range nums {
			ext, err := client.ResolveExtension(msgTypeName, fieldNum)
			if err != nil {
				return err
			}
			fds = append(fds, ext)
		}
		for _, fd := range fds {
			if err := ext.AddExtension(fd); err != nil {
				return err
			}
		}
	}
	for _, fd := range md.GetFields() {
		if fd.GetMessageType() != nil {
			err := fetchAllExtensions(client, ext, fd.GetMessageType(), alreadyFetched)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
