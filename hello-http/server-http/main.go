package main

import (
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	gw "git.vodjk.com/go-grpc/example/proto"
)

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	endpoint := "127.0.0.1:50052"
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := gw.RegisterHelloHttpHandlerFromEndpoint(ctx, mux, endpoint, opts)
	if err != nil {
		return err
	}

	fmt.Println("Listen on 8080")
	http.ListenAndServe(":8080", mux)
	return nil
}

func main() {
	defer glog.Flush()

	if err := run(); err != nil {
		glog.Fatal(err)
	}
}
