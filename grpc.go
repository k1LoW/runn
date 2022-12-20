package runn

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/goccy/go-json"
	"github.com/mitchellh/copystructure"

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

type GRPCType string

const (
	GRPCUnary           GRPCType = "unary"
	GRPCServerStreaming GRPCType = "server"
	GRPCClientStreaming GRPCType = "client"
	GRPCBidiStreaming   GRPCType = "bidi"
)

type GRPCOp string

const (
	GRPCOpMessage GRPCOp = "message"
	GRPCOpReceive GRPCOp = "receive"
	GRPCOpClose   GRPCOp = "close"
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
	op     GRPCOp
	params map[string]interface{}
}

type grpcRequest struct {
	service  string
	method   string
	headers  metadata.MD
	messages []*grpcMessage
}

func newGrpcRunner(name, target string) (*grpcRunner, error) {
	return &grpcRunner{
		name:   name,
		target: target,
		mds:    map[string]*desc.MethodDescriptor{},
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
		useTLS := true
		if strings.HasSuffix(rnr.target, ":80") {
			useTLS = false
		}
		if rnr.tls != nil {
			useTLS = *rnr.tls
		}
		if useTLS {
			tlsc := tls.Config{MinVersion: tls.VersionTLS12}
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
		rnr.grefc = grpcreflect.NewClientV1Alpha(ctx, stub)
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
		rnr.operator.capturers.captureGRPCStart(rnr.name, GRPCUnary, r.service, r.method)
		defer rnr.operator.capturers.captureGRPCEnd(rnr.name, GRPCUnary, r.service, r.method)
		return rnr.invokeUnary(ctx, stub, md, req, r)
	case md.IsServerStreaming() && !md.IsClientStreaming():
		rnr.operator.capturers.captureGRPCStart(rnr.name, GRPCServerStreaming, r.service, r.method)
		defer rnr.operator.capturers.captureGRPCEnd(rnr.name, GRPCServerStreaming, r.service, r.method)
		return rnr.invokeServerStreaming(ctx, stub, md, req, r)
	case !md.IsServerStreaming() && md.IsClientStreaming():
		rnr.operator.capturers.captureGRPCStart(rnr.name, GRPCClientStreaming, r.service, r.method)
		defer rnr.operator.capturers.captureGRPCEnd(rnr.name, GRPCClientStreaming, r.service, r.method)
		return rnr.invokeClientStreaming(ctx, stub, md, req, r)
	case md.IsServerStreaming() && md.IsClientStreaming():
		rnr.operator.capturers.captureGRPCStart(rnr.name, GRPCBidiStreaming, r.service, r.method)
		defer rnr.operator.capturers.captureGRPCEnd(rnr.name, GRPCBidiStreaming, r.service, r.method)
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

	rnr.operator.capturers.captureGRPCRequestHeaders(r.headers)

	if err := rnr.setMessage(req, r.messages[0].params); err != nil {
		return err
	}

	rnr.operator.capturers.captureGRPCRequestMessage(r.messages[0].params)

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

	rnr.operator.capturers.captureGRPCResponseStatus(int(stat.Code()))
	rnr.operator.capturers.captureGRPCResponseHeaders(resHeaders)
	rnr.operator.capturers.captureGRPCResponseTrailers(resTrailers)

	messages := []map[string]interface{}{}
	if stat.Code() == codes.OK {
		m := new(bytes.Buffer)
		marshaler := jsonpb.Marshaler{
			OrigName: true,
		}
		if err := marshaler.Marshal(m, res); err != nil {
			return err
		}
		var msg map[string]interface{}
		if err := json.Unmarshal(m.Bytes(), &msg); err != nil {
			return err
		}
		d["message"] = msg

		rnr.operator.capturers.captureGRPCResponseMessage(msg)

		messages = append(messages, msg)
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

	rnr.operator.capturers.captureGRPCRequestHeaders(r.headers)

	if err := rnr.setMessage(req, r.messages[0].params); err != nil {
		return err
	}

	rnr.operator.capturers.captureGRPCRequestMessage(r.messages[0].params)

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

		rnr.operator.capturers.captureGRPCResponseStatus(int(stat.Code()))

		if stat.Code() == codes.OK {
			m := new(bytes.Buffer)
			marshaler := jsonpb.Marshaler{
				OrigName: true,
			}
			if err := marshaler.Marshal(m, res); err != nil {
				return err
			}
			var msg map[string]interface{}
			if err := json.Unmarshal(m.Bytes(), &msg); err != nil {
				return err
			}
			d["message"] = msg

			rnr.operator.capturers.captureGRPCResponseMessage(msg)

			messages = append(messages, msg)
		}
	}
	d["messages"] = messages
	if h, err := stream.Header(); err == nil {
		d["headers"] = h

		rnr.operator.capturers.captureGRPCResponseHeaders(h)
	}
	t := stream.Trailer()
	d["trailers"] = t

	rnr.operator.capturers.captureGRPCResponseTrailers(t)

	rnr.operator.record(map[string]interface{}{
		"res": d,
	})

	return nil
}

func (rnr *grpcRunner) invokeClientStreaming(ctx context.Context, stub grpcdynamic.Stub, md *desc.MethodDescriptor, req proto.Message, r *grpcRequest) error {
	ctx = setHeaders(ctx, r.headers)

	rnr.operator.capturers.captureGRPCRequestHeaders(r.headers)

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
		case GRPCOpMessage:
			if err := rnr.setMessage(req, m.params); err != nil {
				return err
			}

			rnr.operator.capturers.captureGRPCRequestMessage(m.params)

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

	rnr.operator.capturers.captureGRPCResponseStatus(int(stat.Code()))

	if stat.Code() == codes.OK {
		m := new(bytes.Buffer)
		marshaler := jsonpb.Marshaler{
			OrigName: true,
		}
		if err := marshaler.Marshal(m, res); err != nil {
			return err
		}
		var msg map[string]interface{}
		if err := json.Unmarshal(m.Bytes(), &msg); err != nil {
			return err
		}
		d["message"] = msg

		rnr.operator.capturers.captureGRPCResponseMessage(msg)

		messages = append(messages, msg)
	}
	d["messages"] = messages
	if h, err := stream.Header(); err == nil {
		d["headers"] = h

		rnr.operator.capturers.captureGRPCResponseHeaders(h)
	}
	t := stream.Trailer()
	d["trailers"] = t

	rnr.operator.capturers.captureGRPCResponseTrailers(t)

	rnr.operator.record(map[string]interface{}{
		"res": d,
	})

	return nil
}

func (rnr *grpcRunner) invokeBidiStreaming(ctx context.Context, stub grpcdynamic.Stub, md *desc.MethodDescriptor, req proto.Message, r *grpcRequest) error {
	ctx = setHeaders(ctx, r.headers)

	rnr.operator.capturers.captureGRPCRequestHeaders(r.headers)

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
		case GRPCOpMessage:
			if err := rnr.setMessage(req, m.params); err != nil {
				return err
			}
			err = stream.SendMsg(req)

			rnr.operator.capturers.captureGRPCRequestMessage(m.params)

			req.Reset()
		case GRPCOpReceive:
			res, err := stream.RecvMsg()
			stat, ok := status.FromError(err)
			if !ok {
				return err
			}
			d["status"] = int64(stat.Code())

			rnr.operator.capturers.captureGRPCResponseStatus(int(stat.Code()))

			if h, err := stream.Header(); err == nil {
				d["headers"] = h

				rnr.operator.capturers.captureGRPCResponseHeaders(h)
			}
			if stat.Code() == codes.OK {
				m := new(bytes.Buffer)
				marshaler := jsonpb.Marshaler{
					OrigName: true,
				}
				if err := marshaler.Marshal(m, res); err != nil {
					return err
				}
				var msg map[string]interface{}
				if err := json.Unmarshal(m.Bytes(), &msg); err != nil {
					return err
				}
				d["message"] = msg

				rnr.operator.capturers.captureGRPCResponseMessage(msg)

				messages = append(messages, msg)
			}
		case GRPCOpClose:
			clientClose = true
			err = stream.CloseSend()
			rnr.operator.capturers.captureGRPCClientClose()
			break L
		default:
			return fmt.Errorf("invalid op: %v", m.op)
		}
	}
	stat, ok := status.FromError(err)
	if !ok {
		return err
	}
	if stat.Code() != codes.OK {
		d["status"] = int64(stat.Code())

		rnr.operator.capturers.captureGRPCResponseStatus(int(stat.Code()))
	}

	if clientClose {
		for {
			if _, err := stream.RecvMsg(); err != nil {
				if err != io.EOF {
					return err
				}
				break
			} else {
				if err := stream.CloseSend(); err != nil {
					return err
				}
			}
		}
	} else {
		if err == nil {
			for {
				res, err := stream.RecvMsg()
				if err == io.EOF {
					break
				}
				stat, ok := status.FromError(err)
				if !ok {
					return err
				}
				d["status"] = int64(stat.Code())

				rnr.operator.capturers.captureGRPCResponseStatus(int(stat.Code()))
				if stat.Code() == codes.OK {
					m := new(bytes.Buffer)
					marshaler := jsonpb.Marshaler{
						OrigName: true,
					}
					if err := marshaler.Marshal(m, res); err != nil {
						return err
					}
					var msg map[string]interface{}
					if err := json.Unmarshal(m.Bytes(), &msg); err != nil {
						return err
					}
					d["message"] = msg

					rnr.operator.capturers.captureGRPCResponseMessage(msg)

					messages = append(messages, msg)
				}
			}
		}
	}

	// If the connection is not disconnected here, it will fall into a race condition when retrieving the trailer.
	if err := rnr.cc.Close(); err != nil {
		return err
	}
	rnr.cc = nil

	d["messages"] = messages
	if h, err := stream.Header(); len(d["headers"].(metadata.MD)) == 0 && err == nil {
		d["headers"] = h
	}
	t := dcopy(stream.Trailer()).(metadata.MD)
	d["trailers"] = t

	rnr.operator.capturers.captureGRPCResponseTrailers(t)

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
	e, err := rnr.operator.expandBeforeRecord(message)
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

func dcopy(in interface{}) interface{} {
	return copystructure.Must(copystructure.Copy(in))
}
