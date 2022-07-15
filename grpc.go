package runn

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/golang/protobuf/jsonpb" //nolint
	"github.com/golang/protobuf/proto"  //nolint
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"github.com/jhump/protoreflect/grpcreflect"
	"github.com/k1LoW/runn/version"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	rpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/grpc/status"
)

type grpcOp string

const (
	grpcOpMessage grpcOp = "message"
	grpcOpRecieve grpcOp = "recieve"
	grpcOpClose   grpcOp = "close"
)

type grpcRunner struct {
	name       string
	target     string
	tls        *bool
	cacert     []byte
	cert       []byte
	key        []byte
	skipVerify bool
	cc         *grpc.ClientConn
	grefc      *grpcreflect.Client
	mds        map[string]*desc.MethodDescriptor
	operator   *operator
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
		opts := []grpc.DialOption{
			grpc.WithBlock(),
			grpc.WithUserAgent(fmt.Sprintf("runn/%s", version.Version)),
		}
		useTLS := false
		if strings.HasSuffix(rnr.target, ":443") {
			useTLS = true
		}
		if rnr.tls != nil {
			useTLS = *rnr.tls
		}
		if useTLS {
			tlsc := tls.Config{}
			if rnr.cert != nil {
				certificate, err := tls.X509KeyPair(rnr.cert, rnr.key)
				if err != nil {
					return err
				}
				tlsc.Certificates = []tls.Certificate{certificate}
			}
			if rnr.skipVerify {
				tlsc.InsecureSkipVerify = true
			} else if rnr.cacert != nil {
				pool := x509.NewCertPool()
				if ok := pool.AppendCertsFromPEM(rnr.cacert); !ok {
					return errors.New("failed to append ca certs")
				}
				tlsc.RootCAs = pool
			}
			opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tlsc)))
		} else {
			opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		}

		cc, err := grpc.DialContext(ctx, rnr.target, opts...)
		if err != nil {
			return err
		}
		rnr.cc = cc
	}
	if len(rnr.mds) == 0 {
		stub := rpb.NewServerReflectionClient(rnr.cc)
		rnr.grefc = grpcreflect.NewClient(ctx, stub)
		if err := rnr.resolveAllMethods(ctx); err != nil {
			return err
		}
	}
	key := strings.Join([]string{r.service, r.method}, "/")
	md, ok := rnr.mds[key]
	if !ok {
		return fmt.Errorf("cannot find method: %s", key)
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
	stub := grpcdynamic.NewStubWithMessageFactory(rnr.cc, mf)
	req := mf.NewMessage(md.GetInputType())
	switch {
	case !md.IsServerStreaming() && !md.IsClientStreaming():
		return rnr.invokeUnary(ctx, stub, md, req, r)
	case md.IsServerStreaming() && !md.IsClientStreaming():
		return rnr.invokeServerStreaming(ctx, stub, md, req, r)
	case !md.IsServerStreaming() && md.IsClientStreaming():
		return rnr.invokeClientStreaming(ctx, stub, md, req, r)
	case md.IsServerStreaming() && md.IsClientStreaming():
		return rnr.invokeBidiStreaming(ctx, stub, md, req, r)
	default:
		return errors.New("something strange happened")
	}
}

func (rnr *grpcRunner) invokeUnary(ctx context.Context, stub grpcdynamic.Stub, md *desc.MethodDescriptor, req proto.Message, r *grpcRequest) error {
	if len(r.messages) != 1 {
		return errors.New("unary RPC message should be 1")
	}
	ctx = setHeaders(ctx, r.headers)
	if err := rnr.setMessage(req, r.messages[0].params); err != nil {
		return err
	}
	var (
		resHeaders  metadata.MD
		resTrailers metadata.MD
	)
	res, err := stub.InvokeRpc(ctx, md, req, grpc.Header(&resHeaders), grpc.Trailer(&resTrailers))
	stat, ok := status.FromError(err)
	if !ok {
		return err
	}
	d := map[string]interface{}{
		"status":   int(stat.Code()),
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

func (rnr *grpcRunner) invokeServerStreaming(ctx context.Context, stub grpcdynamic.Stub, md *desc.MethodDescriptor, req proto.Message, r *grpcRequest) error {
	if len(r.messages) != 1 {
		return errors.New("server streaming RPC message should be 1")
	}
	ctx = setHeaders(ctx, r.headers)
	if err := rnr.setMessage(req, r.messages[0].params); err != nil {
		return err
	}
	stream, err := stub.InvokeRpcServerStream(ctx, md, req)
	if err != nil {
		return err
	}
	d := map[string]interface{}{
		"headers":  metadata.MD{},
		"trailers": metadata.MD{},
		"message":  nil,
	}
	messages := []map[string]interface{}{}
	for err == nil {
		var res proto.Message
		res, err = stream.RecvMsg()
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		stat, ok := status.FromError(err)
		if !ok {
			return err
		}
		d["status"] = int64(stat.Code())
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
		}
	}
	d["messages"] = messages
	if h, err := stream.Header(); err == nil {
		d["headers"] = h
	}
	d["trailers"] = stream.Trailer()

	rnr.operator.record(map[string]interface{}{
		"res": d,
	})

	return nil
}

func (rnr *grpcRunner) invokeClientStreaming(ctx context.Context, stub grpcdynamic.Stub, md *desc.MethodDescriptor, req proto.Message, r *grpcRequest) error {
	ctx = setHeaders(ctx, r.headers)
	stream, err := stub.InvokeRpcClientStream(ctx, md)
	if err != nil {
		return err
	}
	d := map[string]interface{}{
		"headers":  metadata.MD{},
		"trailers": metadata.MD{},
		"message":  nil,
	}
	messages := []map[string]interface{}{}
	for _, m := range r.messages {
		switch m.op {
		case grpcOpMessage:
			if err := rnr.setMessage(req, m.params); err != nil {
				return err
			}
			if err := stream.SendMsg(req); err == io.EOF {
				break
			}
		default:
			return fmt.Errorf("invalid op: %v", m.op)
		}
		req.Reset()
	}
	res, err := stream.CloseAndReceive()
	stat, ok := status.FromError(err)
	if !ok {
		return err
	}
	d["status"] = int64(stat.Code())
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
	}
	d["messages"] = messages
	if h, err := stream.Header(); err == nil {
		d["headers"] = h
	}
	d["trailers"] = stream.Trailer()

	rnr.operator.record(map[string]interface{}{
		"res": d,
	})

	return nil
}

func (rnr *grpcRunner) invokeBidiStreaming(ctx context.Context, stub grpcdynamic.Stub, md *desc.MethodDescriptor, req proto.Message, r *grpcRequest) error {
	ctx = setHeaders(ctx, r.headers)
	stream, err := stub.InvokeRpcBidiStream(ctx, md)
	if err != nil {
		return err
	}
	d := map[string]interface{}{
		"headers":  metadata.MD{},
		"trailers": metadata.MD{},
		"message":  nil,
	}
	messages := []map[string]interface{}{}
	clientClose := false
L:
	for _, m := range r.messages {
		switch m.op {
		case grpcOpMessage:
			if err := rnr.setMessage(req, m.params); err != nil {
				return err
			}
			err = stream.SendMsg(req)
		case grpcOpRecieve:
			res, err := stream.RecvMsg()
			stat, ok := status.FromError(err)
			if !ok {
				return err
			}
			d["status"] = int64(stat.Code())
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
			}
		case grpcOpClose:
			clientClose = true
			err = stream.CloseSend()
			break L
		default:
			return fmt.Errorf("invalid op: %v", m.op)
		}
		req.Reset()
	}
	stat, ok := status.FromError(err)
	if !ok {
		return err
	}
	d["status"] = int64(stat.Code())
	if !clientClose {
		for err == nil {
			res, err := stream.RecvMsg()
			if err == io.EOF {
				break
			}
			stat, ok := status.FromError(err)
			if !ok {
				return err
			}
			d["status"] = int64(stat.Code())
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
			}
		}
	}

	d["messages"] = messages
	if h, err := stream.Header(); err == nil {
		d["headers"] = h
	}
	d["trailers"] = stream.Trailer()

	rnr.operator.record(map[string]interface{}{
		"res": d,
	})

	return nil
}

func setHeaders(ctx context.Context, h metadata.MD) context.Context {
	kv := []string{}
	for k, v := range h {
		kv = append(kv, k)
		kv = append(kv, v...)
	}
	ctx = metadata.AppendToOutgoingContext(ctx, kv...)
	return ctx
}

func (rnr *grpcRunner) setMessage(req proto.Message, message map[string]interface{}) error {
	e, err := rnr.operator.expand(message)
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
