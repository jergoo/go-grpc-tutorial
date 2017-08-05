package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/jergoo/go-grpc-example/proto"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
)

// 定义helloHttpService并实现约定的接口
type helloHttpService struct{}

// HelloHttpService ...
var HelloHttpService = helloHttpService{}

func (h helloHttpService) SayHello(ctx context.Context, in *pb.HelloHttpRequest) (*pb.HelloHttpReply, error) {
	resp := new(pb.HelloHttpReply)
	resp.Message = "Hello " + in.Name + "."

	return resp, nil
}

// grpcHandlerFunc returns an http.Handler that delegates to grpcServer on incoming gRPC
// connections or otherHandler otherwise. Copied from cockroachdb.
func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO(tamird): point to merged gRPC code rather than a PR.
		// This is a partial recreation of gRPC's internal checks https://github.com/grpc/grpc-go/pull/514/files#diff-95e9a25b738459a2d3030e1e6fa2a718R61
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

func main() {
	endpoint := "127.0.0.1:50052"

	// grpc server
	creds, err := credentials.NewServerTLSFromFile("../../keys/server.pem", "../../keys/server.key")
	if err != nil {
		grpclog.Fatalf("Failed to generate credentials %v", err)
	}
	conn, _ := net.Listen("tcp", endpoint)
	grpcServer := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterHelloHttpServer(grpcServer, HelloHttpService)

	// http-grpc
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	dcreds, err := credentials.NewClientTLSFromFile("../../keys/server.pem", "server name")
	if err != nil {
		grpclog.Fatalf("Failed to create TLS credentials %v", err)
	}
	dopts := []grpc.DialOption{grpc.WithTransportCredentials(dcreds)}
	gwmux := runtime.NewServeMux()
	err = pb.RegisterHelloHttpHandlerFromEndpoint(ctx, gwmux, endpoint, dopts)
	if err != nil {
		fmt.Printf("serve: %v\n", err)
		return
	}

	mux := http.NewServeMux()
	mux.Handle("/", gwmux)

	if err != nil {
		panic(err)
	}

	cert, _ := ioutil.ReadFile("../../keys/server.pem")
	key, _ := ioutil.ReadFile("../../keys/server.key")
	var demoKeyPair *tls.Certificate
	pair, err := tls.X509KeyPair(cert, key)
	if err != nil {
		panic(err)
	}
	demoKeyPair = &pair

	srv := &http.Server{
		Addr:    endpoint,
		Handler: grpcHandlerFunc(grpcServer, mux),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*demoKeyPair},
		},
	}

	fmt.Printf("grpc and https on port: %d\n", 50052)

	err = srv.Serve(tls.NewListener(conn, srv.TLSConfig))

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	return
}
