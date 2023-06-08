package runn

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/goccy/go-json"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/k1LoW/runn/version"
	"github.com/ktr0731/evans/grpc/grpcreflection"
	"github.com/mitchellh/copystructure"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
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

const (
	grpcStoreStatusKey   = "status"
	grpcStoreHeaderKey   = "headers"
	grpcStoreTrailerKey  = "trailers"
	grpcStoreMessageKey  = "message"
	grpcStoreMessagesKey = "messages"
	grpcStoreResponseKey = "res"
)

type grpcRunner struct {
	name        string
	target      string
	tls         *bool
	cacert      []byte
	cert        []byte
	key         []byte
	skipVerify  bool
	importPaths []string
	protos      []string
	cc          *grpc.ClientConn
	mds         map[string]protoreflect.MethodDescriptor
	operator    *operator
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
	timeout  time.Duration
}

func newGrpcRunner(name, target string) (*grpcRunner, error) {
	return &grpcRunner{
		name:   name,
		target: target,
		mds:    map[string]protoreflect.MethodDescriptor{},
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
			grpc.WithReturnConnectionError(),
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
				certpool, err := x509.SystemCertPool()
				if err != nil {
					// FIXME for Windows
					// ref: https://github.com/golang/go/issues/18609
					certpool = x509.NewCertPool()
				}
				if ok := certpool.AppendCertsFromPEM(rnr.cacert); !ok {
					return errors.New("failed to append cacert")
				}
				tlsc.RootCAs = certpool
			}
			opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tlsc)))
		} else {
			opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		}
		cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		cc, err := grpc.DialContext(cctx, rnr.target, opts...)
		if err != nil {
			return err
		}
		rnr.cc = cc
	}
	if len(rnr.importPaths) > 0 || len(rnr.protos) > 0 {
		if err := rnr.resolveAllMethodsUsingProtos(); err != nil {
			return err
		}
	}
	if len(rnr.mds) == 0 {
		if err := rnr.resolveAllMethodsUsingReflection(ctx); err != nil {
			return err
		}
	}
	key := strings.Join([]string{r.service, r.method}, "/")
	md, ok := rnr.mds[key]
	if !ok {
		return fmt.Errorf("cannot find method: %s", key)
	}
	switch {
	case !md.IsStreamingServer() && !md.IsStreamingClient():
		rnr.operator.capturers.captureGRPCStart(rnr.name, GRPCUnary, r.service, r.method)
		defer rnr.operator.capturers.captureGRPCEnd(rnr.name, GRPCUnary, r.service, r.method)
		return rnr.invokeUnary(ctx, md, r)
	case md.IsStreamingServer() && !md.IsStreamingClient():
		rnr.operator.capturers.captureGRPCStart(rnr.name, GRPCServerStreaming, r.service, r.method)
		defer rnr.operator.capturers.captureGRPCEnd(rnr.name, GRPCServerStreaming, r.service, r.method)
		return rnr.invokeServerStreaming(ctx, md, r)
	case !md.IsStreamingServer() && md.IsStreamingClient():
		rnr.operator.capturers.captureGRPCStart(rnr.name, GRPCClientStreaming, r.service, r.method)
		defer rnr.operator.capturers.captureGRPCEnd(rnr.name, GRPCClientStreaming, r.service, r.method)
		return rnr.invokeClientStreaming(ctx, md, r)
	case md.IsStreamingServer() && md.IsStreamingClient():
		rnr.operator.capturers.captureGRPCStart(rnr.name, GRPCBidiStreaming, r.service, r.method)
		defer rnr.operator.capturers.captureGRPCEnd(rnr.name, GRPCBidiStreaming, r.service, r.method)
		return rnr.invokeBidiStreaming(ctx, md, r)
	default:
		return errors.New("something strange happened")
	}
}

func (rnr *grpcRunner) invokeUnary(ctx context.Context, md protoreflect.MethodDescriptor, r *grpcRequest) error {
	if len(r.messages) != 1 {
		return errors.New("unary RPC message should be 1")
	}
	if r.timeout > 0 {
		cctx, cancel := context.WithTimeout(ctx, r.timeout)
		ctx = cctx
		defer cancel()
	}

	ctx = setHeaders(ctx, r.headers)
	req := dynamicpb.NewMessage(md.Input())

	rnr.operator.capturers.captureGRPCRequestHeaders(r.headers)

	if err := rnr.setMessage(req, r.messages[0].params); err != nil {
		return err
	}

	rnr.operator.capturers.captureGRPCRequestMessage(r.messages[0].params)

	var (
		resHeaders  metadata.MD
		resTrailers metadata.MD
	)
	res := dynamicpb.NewMessage(md.Output())
	err := rnr.cc.Invoke(ctx, toEndpoint(md.FullName()), req, res, grpc.Header(&resHeaders), grpc.Trailer(&resTrailers))
	stat, ok := status.FromError(err)
	if !ok {
		return err
	}

	d := map[string]interface{}{
		string(grpcStoreStatusKey):  int(stat.Code()),
		string(grpcStoreHeaderKey):  resHeaders,
		string(grpcStoreTrailerKey): resTrailers,
		string(grpcStoreMessageKey): nil,
	}

	rnr.operator.capturers.captureGRPCResponseStatus(stat)
	rnr.operator.capturers.captureGRPCResponseHeaders(resHeaders)
	rnr.operator.capturers.captureGRPCResponseTrailers(resTrailers)

	messages := []map[string]interface{}{}
	if stat.Code() == codes.OK {
		b, err := protojson.MarshalOptions{UseProtoNames: true, UseEnumNumbers: true, EmitUnpopulated: true}.Marshal(res)
		if err != nil {
			return err
		}
		var msg map[string]interface{}
		if err := json.Unmarshal(b, &msg); err != nil {
			return err
		}
		d[grpcStoreMessageKey] = msg

		rnr.operator.capturers.captureGRPCResponseMessage(msg)

		messages = append(messages, msg)
		d[grpcStoreMessagesKey] = messages
	} else {
		d[grpcStoreMessageKey] = stat.Message()
	}

	rnr.operator.record(map[string]interface{}{
		string(grpcStoreResponseKey): d,
	})
	return nil
}

func (rnr *grpcRunner) invokeServerStreaming(ctx context.Context, md protoreflect.MethodDescriptor, r *grpcRequest) error {
	if len(r.messages) != 1 {
		return errors.New("server streaming RPC message should be 1")
	}
	if r.timeout > 0 {
		cctx, cancel := context.WithTimeout(ctx, r.timeout)
		ctx = cctx
		defer cancel()
	}

	ctx = setHeaders(ctx, r.headers)
	req := dynamicpb.NewMessage(md.Input())

	rnr.operator.capturers.captureGRPCRequestHeaders(r.headers)

	if err := rnr.setMessage(req, r.messages[0].params); err != nil {
		return err
	}
	rnr.operator.capturers.captureGRPCRequestMessage(r.messages[0].params)

	streamDesc := &grpc.StreamDesc{
		StreamName:    string(md.Name()),
		ServerStreams: md.IsStreamingServer(),
		ClientStreams: md.IsStreamingClient(),
	}

	stream, err := rnr.cc.NewStream(ctx, streamDesc, toEndpoint(md.FullName()))
	if err != nil {
		return err
	}
	if err := stream.SendMsg(req); err != nil {
		return err
	}

	d := map[string]interface{}{
		string(grpcStoreHeaderKey):  metadata.MD{},
		string(grpcStoreTrailerKey): metadata.MD{},
		string(grpcStoreMessageKey): nil,
	}
	messages := []map[string]interface{}{}

	for err == nil {
		res := dynamicpb.NewMessage(md.Output())
		err = stream.RecvMsg(res)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				break
			}
			if errors.Is(err, io.EOF) {
				break
			}
		}
		stat, ok := status.FromError(err)
		if !ok {
			return err
		}
		d[grpcStoreStatusKey] = int64(stat.Code())

		rnr.operator.capturers.captureGRPCResponseStatus(stat)

		if stat.Code() == codes.OK {
			b, err := protojson.MarshalOptions{UseProtoNames: true, UseEnumNumbers: true, EmitUnpopulated: true}.Marshal(res)
			if err != nil {
				return err
			}
			var msg map[string]interface{}
			if err := json.Unmarshal(b, &msg); err != nil {
				return err
			}
			d[grpcStoreMessageKey] = msg

			rnr.operator.capturers.captureGRPCResponseMessage(msg)

			messages = append(messages, msg)
		} else {
			d[grpcStoreMessageKey] = stat.Message()
		}
	}
	d[grpcStoreMessagesKey] = messages
	if h, err := stream.Header(); err == nil {
		d[grpcStoreHeaderKey] = h

		rnr.operator.capturers.captureGRPCResponseHeaders(h)
	}
	t := stream.Trailer()
	d[grpcStoreTrailerKey] = t

	rnr.operator.capturers.captureGRPCResponseTrailers(t)

	rnr.operator.record(map[string]interface{}{
		string(grpcStoreResponseKey): d,
	})

	return nil
}

func (rnr *grpcRunner) invokeClientStreaming(ctx context.Context, md protoreflect.MethodDescriptor, r *grpcRequest) error {
	if r.timeout > 0 {
		cctx, cancel := context.WithTimeout(ctx, r.timeout)
		ctx = cctx
		defer cancel()
	}

	ctx = setHeaders(ctx, r.headers)

	rnr.operator.capturers.captureGRPCRequestHeaders(r.headers)

	streamDesc := &grpc.StreamDesc{
		StreamName:    string(md.Name()),
		ServerStreams: md.IsStreamingServer(),
		ClientStreams: md.IsStreamingClient(),
	}
	stream, err := rnr.cc.NewStream(ctx, streamDesc, toEndpoint(md.FullName()))
	if err != nil {
		return err
	}
	d := map[string]interface{}{
		string(grpcStoreHeaderKey):  metadata.MD{},
		string(grpcStoreTrailerKey): metadata.MD{},
		string(grpcStoreMessageKey): nil,
	}
	messages := []map[string]interface{}{}
	for _, m := range r.messages {
		switch m.op {
		case GRPCOpMessage:
			req := dynamicpb.NewMessage(md.Input())

			if err := rnr.setMessage(req, m.params); err != nil {
				return err
			}

			rnr.operator.capturers.captureGRPCRequestMessage(m.params)

			err := stream.SendMsg(req)
			if errors.Is(err, context.Canceled) {
				break
			}
			if errors.Is(err, io.EOF) {
				break
			}
		default:
			return fmt.Errorf("invalid op: %v", m.op)
		}
	}
	res := dynamicpb.NewMessage(md.Output())
	if err := stream.CloseSend(); err != nil {
		return err
	}
	err = stream.RecvMsg(res)
	stat, ok := status.FromError(err)
	if !ok {
		return err
	}

	d[grpcStoreStatusKey] = int64(stat.Code())

	rnr.operator.capturers.captureGRPCResponseStatus(stat)

	if stat.Code() == codes.OK {
		b, err := protojson.MarshalOptions{UseProtoNames: true, UseEnumNumbers: true, EmitUnpopulated: true}.Marshal(res)
		if err != nil {
			return err
		}
		var msg map[string]interface{}
		if err := json.Unmarshal(b, &msg); err != nil {
			return err
		}
		d[grpcStoreMessageKey] = msg

		rnr.operator.capturers.captureGRPCResponseMessage(msg)

		messages = append(messages, msg)
	} else {
		d[grpcStoreMessageKey] = stat.Message()
	}

	d[grpcStoreMessagesKey] = messages
	if h, err := stream.Header(); err == nil {
		d[grpcStoreHeaderKey] = h

		rnr.operator.capturers.captureGRPCResponseHeaders(h)
	}
	t := stream.Trailer()
	d[grpcStoreTrailerKey] = t

	rnr.operator.capturers.captureGRPCResponseTrailers(t)

	rnr.operator.record(map[string]interface{}{
		string(grpcStoreResponseKey): d,
	})

	return nil
}

func (rnr *grpcRunner) invokeBidiStreaming(ctx context.Context, md protoreflect.MethodDescriptor, r *grpcRequest) error {
	if r.timeout > 0 {
		return errors.New("unsupported timeout: for bidirectional streaming RPC")
	}

	ctx = setHeaders(ctx, r.headers)
	rnr.operator.capturers.captureGRPCRequestHeaders(r.headers)

	streamDesc := &grpc.StreamDesc{
		StreamName:    string(md.Name()),
		ServerStreams: md.IsStreamingServer(),
		ClientStreams: md.IsStreamingClient(),
	}

	stream, err := rnr.cc.NewStream(ctx, streamDesc, toEndpoint(md.FullName()))
	if err != nil {
		return err
	}

	d := map[string]interface{}{
		string(grpcStoreHeaderKey):  metadata.MD{},
		string(grpcStoreTrailerKey): metadata.MD{},
		string(grpcStoreMessageKey): nil,
	}
	messages := []map[string]interface{}{}
	clientClose := false
L:
	for _, m := range r.messages {
		switch m.op {
		case GRPCOpMessage:
			req := dynamicpb.NewMessage(md.Input())
			if err := rnr.setMessage(req, m.params); err != nil {
				return err
			}
			err = stream.SendMsg(req)
			if errors.Is(err, context.Canceled) {
				break L
			}
			if errors.Is(err, io.EOF) {
				break L
			}
			rnr.operator.capturers.captureGRPCRequestMessage(m.params)

			req.Reset()
		case GRPCOpReceive:
			res := dynamicpb.NewMessage(md.Output())
			err := stream.RecvMsg(res)
			if errors.Is(err, context.Canceled) {
				break L
			}
			if errors.Is(err, io.EOF) {
				break L
			}
			stat, ok := status.FromError(err)
			if !ok {
				return err
			}
			d[grpcStoreStatusKey] = int64(stat.Code())

			rnr.operator.capturers.captureGRPCResponseStatus(stat)

			if h, err := stream.Header(); err == nil {
				d[grpcStoreHeaderKey] = h

				rnr.operator.capturers.captureGRPCResponseHeaders(h)
			}
			if stat.Code() == codes.OK {
				b, err := protojson.MarshalOptions{UseProtoNames: true, UseEnumNumbers: true, EmitUnpopulated: true}.Marshal(res)
				if err != nil {
					return err
				}
				var msg map[string]interface{}
				if err := json.Unmarshal(b, &msg); err != nil {
					return err
				}
				d[grpcStoreMessageKey] = msg

				rnr.operator.capturers.captureGRPCResponseMessage(msg)

				messages = append(messages, msg)
			} else {
				d[grpcStoreMessageKey] = stat.Message()
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
		d[grpcStoreStatusKey] = int64(stat.Code())
		d[grpcStoreMessageKey] = stat.Message()

		rnr.operator.capturers.captureGRPCResponseStatus(stat)
	}

	if clientClose {
		for {
			res := dynamicpb.NewMessage(md.Output())
			if err := stream.RecvMsg(res); err != nil {
				if errors.Is(err, context.Canceled) {
					break
				}
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			} else {
				if err := stream.CloseSend(); err != nil {
					return err
				}
			}
		}
	} else {
		if err == nil {
			for {
				res := dynamicpb.NewMessage(md.Output())
				err := stream.RecvMsg(res)
				if errors.Is(err, context.Canceled) {
					break
				}
				if errors.Is(err, io.EOF) {
					break
				}
				stat, ok := status.FromError(err)
				if !ok {
					return err
				}
				d[grpcStoreStatusKey] = int64(stat.Code())

				rnr.operator.capturers.captureGRPCResponseStatus(stat)
				if stat.Code() == codes.OK {
					b, err := protojson.MarshalOptions{UseProtoNames: true, UseEnumNumbers: true, EmitUnpopulated: true}.Marshal(res)
					if err != nil {
						return err
					}
					var msg map[string]interface{}
					if err := json.Unmarshal(b, &msg); err != nil {
						return err
					}
					d[grpcStoreMessageKey] = msg

					rnr.operator.capturers.captureGRPCResponseMessage(msg)

					messages = append(messages, msg)
				} else {
					d[grpcStoreMessageKey] = stat.Message()
				}
			}
		}
	}

	// If the connection is not disconnected here, it will fall into a race condition when retrieving the trailer.
	if err := rnr.cc.Close(); err != nil {
		return err
	}
	rnr.cc = nil

	d[grpcStoreMessagesKey] = messages
	if h, err := stream.Header(); len(d[grpcStoreHeaderKey].(metadata.MD)) == 0 && err == nil {
		d[grpcStoreHeaderKey] = h
	}
	t, ok := dcopy(stream.Trailer()).(metadata.MD)
	if !ok {
		return fmt.Errorf("failed to copy trailers: %s", t)
	}
	d[grpcStoreTrailerKey] = t

	rnr.operator.capturers.captureGRPCResponseTrailers(t)

	rnr.operator.record(map[string]interface{}{
		string(grpcStoreResponseKey): d,
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
	return protojson.Unmarshal(b, req)
}

func (rnr *grpcRunner) resolveAllMethodsUsingReflection(ctx context.Context) error {
	grefc := grpcreflection.NewClient(rnr.cc, map[string][]string{})
	svcs, err := grefc.ListServices()
	if err != nil {
		return err
	}
	for _, svc := range svcs {
		fd, err := grefc.FindSymbol(svc)
		if err != nil {
			return fmt.Errorf("failed to get service descripter of %s: %w", svc, err)
		}
		sd, ok := fd.(protoreflect.ServiceDescriptor)
		if !ok {
			return fmt.Errorf("failed to get service descripter of %s (%v)", svc, fd)
		}
		mds := sd.Methods()
		for i := 0; i < mds.Len(); i++ {
			md := mds.Get(i)
			key := strings.Join([]string{svc, string(md.Name())}, "/")
			rnr.mds[key] = md
		}
	}
	return nil
}

func (rnr *grpcRunner) resolveAllMethodsUsingProtos() error {
	importPaths, protos, err := resolveWildcardPaths(rnr.importPaths, rnr.protos)
	if err != nil {
		return err
	}
	protos, err = protoparse.ResolveFilenames(importPaths, protos...)
	if err != nil {
		return err
	}
	importPaths, protos, accessor, err := resolvePaths(importPaths, protos...)
	if err != nil {
		return err
	}
	p := protoparse.Parser{
		ImportPaths:           importPaths,
		InferImportPaths:      len(importPaths) == 0,
		IncludeSourceCodeInfo: true,
		Accessor:              accessor,
	}
	dfds, err := p.ParseFiles(protos...)
	if err != nil {
		return err
	}
	if err := registerFiles(dfds); err != nil {
		return err
	}

	fds := desc.ToFileDescriptorSet(dfds...)
	files := protoregistry.GlobalFiles
	for _, fd := range fds.File {
		d, err := protodesc.NewFile(fd, files)
		if err != nil {
			return err
		}
		for i := 0; i < d.Services().Len(); i++ {
			svc := d.Services().Get(i)
			for j := 0; j < svc.Methods().Len(); j++ {
				m := svc.Methods().Get(j)
				key := fmt.Sprintf("%s/%s", svc.FullName(), m.Name())
				rnr.mds[key] = m
			}
		}
	}
	return nil
}

func dcopy(in interface{}) interface{} {
	return copystructure.Must(copystructure.Copy(in))
}

func toEndpoint(mn protoreflect.FullName) string {
	splitted := strings.Split(string(mn), ".")
	service := strings.Join(splitted[:len(splitted)-1], ".")
	method := splitted[len(splitted)-1]
	return fmt.Sprintf("/%s/%s", service, method)
}

func registerFiles(fds []*desc.FileDescriptor) (err error) {
	var rf *protoregistry.Files
	rf, err = protodesc.NewFiles(desc.ToFileDescriptorSet(fds...))
	if err != nil {
		return err
	}

	rf.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		if _, err := protoregistry.GlobalFiles.FindFileByPath(fd.Path()); !errors.Is(protoregistry.NotFound, err) {
			return true
		}

		// Skip registration of conflicted descriptors
		conflict := false
		rangeTopLevelDescriptors(fd, func(d protoreflect.Descriptor) {
			if _, err := protoregistry.GlobalFiles.FindDescriptorByName(d.FullName()); err == nil {
				conflict = true
			}
		})
		if conflict {
			return true
		}

		err = protoregistry.GlobalFiles.RegisterFile(fd)
		return (err == nil)
	})

	return err
}

// copy from google.golang.org/protobuf/reflect/protoregistry.
func rangeTopLevelDescriptors(fd protoreflect.FileDescriptor, f func(protoreflect.Descriptor)) {
	eds := fd.Enums()
	for i := eds.Len() - 1; i >= 0; i-- {
		f(eds.Get(i))
		vds := eds.Get(i).Values()
		for i := vds.Len() - 1; i >= 0; i-- {
			f(vds.Get(i))
		}
	}
	mds := fd.Messages()
	for i := mds.Len() - 1; i >= 0; i-- {
		f(mds.Get(i))
	}
	xds := fd.Extensions()
	for i := xds.Len() - 1; i >= 0; i-- {
		f(xds.Get(i))
	}
	sds := fd.Services()
	for i := sds.Len() - 1; i >= 0; i-- {
		f(sds.Get(i))
	}
}

func resolveWildcardPaths(importPaths, protos []string) ([]string, []string, error) {
	resolved := []string{}
	for _, proto := range protos {
		if f, err := os.Stat(proto); err == nil {
			if !f.IsDir() {
				resolved = unique(append(resolved, proto))
				continue
			} else {
				proto = filepath.Join(proto, "*")
			}
		}
		base, pattern := doublestar.SplitPattern(filepath.ToSlash(proto))
		importPaths = unique(append(importPaths, base))
		abs, err := filepath.Abs(base)
		if err != nil {
			return nil, nil, err
		}
		fsys := os.DirFS(abs)
		if err := doublestar.GlobWalk(fsys, pattern, func(p string, d fs.DirEntry) error {
			if d.IsDir() {
				return nil
			}
			resolved = unique(append(resolved, filepath.Join(base, p)))
			return nil
		}); err != nil {
			return nil, nil, err
		}
	}
	return importPaths, resolved, nil
}

func resolvePaths(importPaths []string, protos ...string) ([]string, []string, func(filename string) (io.ReadCloser, error), error) {
	resolvedIPaths := importPaths
	resolvedProtos := []string{}
	for _, p := range protos {
		d, b := filepath.Split(p)
		resolvedIPaths = append(resolvedIPaths, d)
		resolvedProtos = append(resolvedProtos, b)
	}
	resolvedIPaths = unique(resolvedIPaths)
	resolvedProtos = unique(resolvedProtos)
	opened := []string{}
	return resolvedIPaths, resolvedProtos, func(filename string) (io.ReadCloser, error) {
		if contains(opened, filename) { // FIXME: Need to resolvePaths well without this condition
			return io.NopCloser(strings.NewReader("")), nil
		}
		opened = append(opened, filename)
		return os.Open(filename)
	}, nil
}
