package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/k1LoW/grpcstub"
)

var Cacert = func() []byte {
	b, err := os.ReadFile(filepath.Join(Testdata(), "cacert.pem"))
	if err != nil {
		panic(err)
	}
	return b
}()

var Cert = func() []byte {
	b, err := os.ReadFile(filepath.Join(Testdata(), "cert.pem"))
	if err != nil {
		panic(err)
	}
	return b
}()

var Key = func() []byte {
	b, err := os.ReadFile(filepath.Join(Testdata(), "key.pem"))
	if err != nil {
		panic(err)
	}
	return b
}()

func GRPCServer(t *testing.T, useTLS bool) *grpcstub.Server {
	var ts *grpcstub.Server
	pf := filepath.Join(Testdata(), "grpctest.proto")
	if useTLS {
		ts = grpcstub.NewTLSServer(t, pf, Cacert, Cert, Key, grpcstub.EnableHealthCheck())
	} else {
		ts = grpcstub.NewServer(t, pf, grpcstub.EnableHealthCheck())
	}
	t.Cleanup(func() {
		ts.Close()
	})
	ts.Method("grpctest.GrpcTestService/Hello").
		Header("hello", "header").Trailer("hello", "trailer").
		ResponseString(`{"message":"hello", "num":32, "create_time":"2022-06-25T05:24:43.861872Z"}`)
	ts.Method("grpctest.GrpcTestService/ListHello").
		Header("listhello", "header").Trailer("listhello", "trailer").
		ResponseString(`{"message":"hello", "num":33, "create_time":"2022-06-25T05:24:43.861872Z"}`).
		ResponseString(`{"message":"hello", "num":34, "create_time":"2022-06-25T05:24:44.382783Z"}`)
	ts.Method("grpctest.GrpcTestService/MultiHello").
		Header("multihello", "header").Trailer("multihello", "trailer").
		ResponseString(`{"message":"hello", "num":35, "create_time":"2022-06-25T05:24:45.382783Z"}`)
	ts.Method("grpctest.GrpcTestService/HelloChat").Match(func(r *grpcstub.Request) bool {
		n, ok := r.Message["name"]
		if !ok {
			return false
		}
		return n.(string) == "alice"
	}).Header("hellochat", "header").Trailer("hellochat", "trailer").
		ResponseString(`{"message":"hello", "num":34, "create_time":"2022-06-25T05:24:46.382783Z"}`)
	ts.Method("grpctest.GrpcTestService/HelloChat").Match(func(r *grpcstub.Request) bool {
		n, ok := r.Message["name"]
		if !ok {
			return false
		}
		return n.(string) == "bob"
	}).Header("hellochat-second", "header").Trailer("hellochat-second", "trailer").
		ResponseString(`{"message":"hello", "num":35, "create_time":"2022-06-25T05:24:47.382783Z"}`)
	ts.Method("grpctest.GrpcTestService/HelloChat").Match(func(r *grpcstub.Request) bool {
		n, ok := r.Message["name"]
		if !ok {
			return false
		}
		return n.(string) == "charlie"
	}).Header("hellochat-third", "header").Trailer("hellochat-second", "trailer").
		ResponseString(`{"message":"hello", "num":36, "create_time":"2022-06-25T05:24:48.382783Z"}`)

	return ts
}
